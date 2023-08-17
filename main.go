package main

import (
	"bufio"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	_ "embed"

	"github.com/arriqaaq/flashdb"
	"github.com/fiatjaf/khatru/plugins/storage/badgern"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/mitchellh/go-homedir"
	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip19"
	"github.com/rs/cors"
	"github.com/rs/zerolog"
)

//go:embed schema.sql
var schema string

var (
	log                     = zerolog.New(os.Stderr).Output(zerolog.ConsoleWriter{Out: os.Stderr})
	srv                     http.Server
	profile                 *Profile
	readRelays, writeRelays []string
	sk                      string
	flash                   *flashdb.FlashDB
	store                   badgern.BadgerBackend
	db                      *sqlx.DB
	serial                  = 0
)

func main() {
	// load default stuff
	datadir, _ := homedir.Expand("~/.config/bisu")
	os.MkdirAll(datadir, 0700)
	path := filepath.Join(datadir, "key")
	b, err := ioutil.ReadFile(path)
	save := false
	if err != nil {
		scanner := bufio.NewScanner(os.Stdin)
		save = true
		fmt.Printf("paste your private key: ")
		if scanner.Scan() {
			text := scanner.Text()
			prefix, value, _ := nip19.Decode(text)
			if prefix == "nsec" {
				text = value.(string)
			}
			b, _ = hex.DecodeString(text)
		} else {
			log.Fatal().Err(err).Msg("can't can't read from stdin")
			return
		}
	}
	if len(b) != 32 {
		log.Fatal().Err(err).Str("path", path).Msg("private key is not 32 bytes")
		return
	}
	if save {
		os.WriteFile(path, b, 0600)
	}
	sk = hex.EncodeToString(b)
	pk, err := nostr.GetPublicKey(sk)
	if err != nil {
		log.Fatal().Err(err).Str("key", sk).Msg("private key is invalid")
		return
	}

	// start sqlite
	if sql, err := sqlx.Open("sqlite3", filepath.Join(datadir, "params.sqlite3")); err != nil {
		log.Fatal().Err(err).Msg("failed to open sqlite3")
		return
	} else {
		db = sql

		sql.Exec(schema)
	}

	// start flashdb
	flash, err = flashdb.New(&flashdb.Config{Path: filepath.Join(datadir, "flash.db"), EvictionInterval: 10})
	if err != nil {
		log.Fatal().Err(err).Msg("unable to start flashdb")
		return
	}

	// start internal event storage
	store.Path = filepath.Join(datadir, "events.db")
	store.MaxLimit = 100000000
	store.Init()

	// initialize data loaders stuff
	initializeDataloaders()

	// cleanup old stuff from event storage
	go func() {
		ctx := context.Background()
		sevenMonthsAgo := nostr.Now() - 60*60*24*30*7
		events, _ := store.QueryEvents(ctx, nostr.Filter{Until: &sevenMonthsAgo})
		for evt := range events {
			if evt.PubKey != profile.pubkey {
				store.DeleteEvent(ctx, evt)
			}
		}
	}()

	// load user metadata
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	profile = loadProfile(ctx, pk)
	if profile == nil {
		// generate a new profile
		event := &nostr.Event{
			Content:   `{}`,
			CreatedAt: nostr.Now(),
			Kind:      0,
		}
		event.Sign(sk)
		profile = &Profile{
			pubkey: pk,
			event:  event,
		}
	}
	readRelays, writeRelays = loadRelaysList(ctx, pk)
	if len(readRelays) == 0 {
		readRelays = append(readRelays, "wss://nostr.mom")
		readRelays = append(readRelays, "wss://relayable.org")
	}
	if len(writeRelays) == 0 {
		writeRelays = append(writeRelays, "wss://nostr.mom")
		writeRelays = append(writeRelays, "wss://nostr-pub.wellorder.net")
	}
	cancel()

	// start listening to relays
	go startListening()

	// routes
	mux := http.NewServeMux()
	// 	mux.HandleFunc("/api/v1/streaming", streamingHandler)
	// 	mux.HandleFunc("/api/v1/streaming/", streamingHandler)
	// 	mux.HandleFunc("/users/:username", actorHandler)
	// 	mux.HandleFunc("/nodeinfo/:version", nodeInfoSchemaHandler)
	mux.HandleFunc("/api/v1/instance", instanceHandler)
	mux.HandleFunc("/api/v1/apps/verify_credentials", appCredentialsHandler)
	mux.HandleFunc("/api/v1/apps", createAppHandler)
	mux.HandleFunc("/oauth/token", createTokenHandler)
	mux.HandleFunc("/oauth/revoke", constantHandler(map[string]any{}))
	mux.HandleFunc("/oauth/authorize", oauthHandler)
	mux.HandleFunc("/api/v1/acccounts",
		constantHandler(map[string]string{
			"error": "this is meant to be run locally, not logged in from the external world",
		}),
	)
	mux.HandleFunc("/api/v1/accounts/verify_credentials", verifyCredentialsHandler)
	//	mux.HandleFunc("/api/v1/accounts/update_credentials", requireAuth, updateCredentialsHandler)
	//	mux.HandleFunc("/api/v1/accounts/search", accountSearchHandler)
	//	mux.HandleFunc("/api/v1/accounts/lookup", accountLookupHandler)
	//	mux.HandleFunc("/api/v1/accounts/relationships", relationshipsHandler)
	//	mux.HandleFunc("/api/v1/accounts/:pubkey{[0-9a-f]{64}}/statuses", accountStatusesHandler)
	//	mux.HandleFunc("/api/v1/accounts/:pubkey{[0-9a-f]{64}}", accountHandler)
	//	mux.HandleFunc("/api/v1/statuses/:id{[0-9a-f]{64}}/context", contextHandler)
	//	mux.HandleFunc("/api/v1/statuses/:id{[0-9a-f]{64}}/favourite", favouriteHandler)
	mux.HandleFunc("/api/v1/statuses", createStatusHandler)
	mux.HandleFunc("/api/v1/statuses/", getStatusHandler)
	mux.HandleFunc("/api/v1/timelines/home", homeHandler)
	// mux.HandleFunc("/api/v1/timelines/public", publicHandler)
	mux.HandleFunc("/api/v1/preferences", constantHandler(map[string]any{
		"posting:default:visibility": "public",
		"posting:default:sensitive":  false,
		"posting:default:language":   nil,
		"reading:expand:media":       "default",
		"reading:expand:spoilers":    false,
	}))
	mux.HandleFunc("/api/v1/search", searchHandler)
	mux.HandleFunc("/api/v2/search", searchHandler)
	mux.HandleFunc("/api/pleroma/frontend_configurations", constantHandler(map[string]any{}))
	//	mux.HandleFunc("/api/v1/trends/tags", trendingTagsHandler)
	//	mux.HandleFunc("/api/v1/trends", trendingTagsHandler)

	// not yet implemented
	mux.HandleFunc("/api/v1/notifications", constantHandler([]any{}))
	mux.HandleFunc("/api/v1/bookmarks", constantHandler([]any{}))
	mux.HandleFunc("/api/v1/custom_emojis", constantHandler([]any{}))
	mux.HandleFunc("/api/v1/accounts/search", constantHandler([]any{}))
	mux.HandleFunc("/api/v1/filters", constantHandler([]any{}))
	mux.HandleFunc("/api/v1/blocks", constantHandler([]any{}))
	mux.HandleFunc("/api/v1/mutes", constantHandler([]any{}))
	mux.HandleFunc("/api/v1/domain_blocks", constantHandler([]any{}))
	mux.HandleFunc("/api/v1/markers", constantHandler(map[string]any{}))
	mux.HandleFunc("/api/v1/conversations", constantHandler([]any{}))
	mux.HandleFunc("/api/v1/favourites", constantHandler([]any{}))
	mux.HandleFunc("/api/v1/lists", constantHandler([]any{}))

	// listen for http with graceful shutdown over sigterm etc
	srv = http.Server{
		Handler: cors.AllowAll().Handler(mux),
		Addr:    fmt.Sprintf("127.0.0.1:7001"),
	}
	sigs := make(chan os.Signal)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		srv.Shutdown(ctx)
	}()

	log.Info().Msg("listening at http://" + srv.Addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal().Err(err).Msg("error serving http")
		return
	}
}

func constantHandler(val any) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(val)
	}
}
