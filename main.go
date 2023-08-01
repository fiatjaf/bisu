package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/cors"
	"github.com/rs/zerolog"
)

var (
	log = zerolog.New(os.Stderr).Output(zerolog.ConsoleWriter{Out: os.Stderr})
	srv http.Server
)

func main() {
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
	//	mux.HandleFunc("/api/v1/acccounts", createAccountHandler)
	//	mux.HandleFunc("/api/v1/accounts/verify_credentials", requireAuth, verifyCredentialsHandler)
	//	mux.HandleFunc("/api/v1/accounts/update_credentials", requireAuth, updateCredentialsHandler)
	//	mux.HandleFunc("/api/v1/accounts/search", accountSearchHandler)
	//	mux.HandleFunc("/api/v1/accounts/lookup", accountLookupHandler)
	//	mux.HandleFunc("/api/v1/accounts/relationships", relationshipsHandler)
	//	mux.HandleFunc("/api/v1/accounts/:pubkey{[0-9a-f]{64}}/statuses", accountStatusesHandler)
	//	mux.HandleFunc("/api/v1/accounts/:pubkey{[0-9a-f]{64}}", accountHandler)
	//	mux.HandleFunc("/api/v1/statuses/:id{[0-9a-f]{64}}/context", contextHandler)
	//	mux.HandleFunc("/api/v1/statuses/:id{[0-9a-f]{64}}", statusHandler)
	//	mux.HandleFunc("/api/v1/statuses/:id{[0-9a-f]{64}}/favourite", favouriteHandler)
	//	mux.HandleFunc("/api/v1/statuses", requireAuth, createStatusHandler)
	//	mux.HandleFunc("/api/v1/timelines/home", requireAuth, homeHandler)
	//	mux.HandleFunc("/api/v1/timelines/public", publicHandler)
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
