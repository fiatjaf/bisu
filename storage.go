package main

import (
	"context"
	"database/sql"
	_ "embed"
	"encoding/json"
	"log"
	"path/filepath"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/nbd-wtf/go-nostr"
)

//go:embed schema.sql
var schema string

type Store struct {
	db *sqlx.DB
}

func InitStore(datadir string) *Store {
	dbpath := filepath.Join(datadir, "bisu")
	db, err := sqlx.Connect("sqlite3", "file:"+dbpath+"?cache=shared")
	if err != nil && err != sql.ErrNoRows {
		log.Fatalln(err)
	}

	db.MustExec(schema)
	db.SetMaxOpenConns(1)

	return &Store{db: db}
}

const (
	maxPerProfile = 80
)

func (s *Store) IncrementRelayScoreForPubkey(pubkey string, relay string, score int) error {
	_, err := s.db.Exec("INSERT INTO relay_scores (pubkey, url, score) VALUES ($1, $2, $3) ON CONFLICT (pubkey, url) DO UPDATE SET score = score + $3",
		pubkey, relay, score)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("failed to save relay score: %s", err.Error())
		return err
	}
	return nil
}

func (s *Store) GetTopRelaysForPubkey(pubkey string, limit int) []string {
	var relays []string
	err := s.db.Select(&relays, "SELECT url FROM relay_scores WHERE pubkey = $1 LIMIT $2", pubkey, limit)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("failed to get top relays: %s", err.Error())
		return nil
	}
	return relays
}

func (s *Store) CacheSingleEvent(event *nostr.Event) {
	b, _ := json.Marshal(event)
	_, err := s.db.Exec("INSERT OR IGNORE INTO events (id, body) VALUES ($1, $2)", event.ID, b)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("failed to cache single event: %s", err.Error())
	}
}

func (s *Store) CacheReplaceableEvent(event *nostr.Event) {
	var b []byte
	err := s.db.Get(&b, "SELECT body FROM event WHERE pubkey = $1 AND kind = $2", event.PubKey, event.Kind)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("failed to check existence of replaceable event: %s", err.Error())
	}
	var curr nostr.Event
	var currId string
	if json.Unmarshal(b, &curr); curr.CreatedAt.Before(event.CreatedAt) {
		s.CacheSingleEvent(event)
		currId = event.ID
	} else {
		currId = curr.ID
	}
	_, err = s.db.Exec(
		"INSERT OR IGNORE INTO replaceable_events (pubkey, kind, id) VALUES ($1, $2, $3)",
		event.PubKey, event.Kind, currId)
	if err != nil {
		log.Printf("failed to cache replaceable event: %s", err.Error())
	}
}

func (s *Store) AddEventToProfileList(event *nostr.Event) {
	var count int
	err := s.db.Get(&count, "SELECT count(*) FROM profile_events WHERE pubkey = $1", event.PubKey)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("failed to count profile events: %s", err.Error())
	}
	if count >= maxPerProfile {
		var oldest int64
		err := s.db.Get(&oldest, "SELECT min(date) FROM profile_events WHERE pubkey = $1", event.PubKey)
		if err != nil && err != sql.ErrNoRows {
			log.Printf("failed to get oldest for profile: %s", err.Error())
		}
		if oldest < event.CreatedAt.Unix() {
			return
		}
		_, err = s.db.Exec("DELETE FROM profile_events WHERE pubkey = $1 AND date = $2", event.PubKey, oldest)
		if err != nil && err != sql.ErrNoRows {
			log.Printf("failed to cleanup profile cache: %s", err.Error())
		}
	}
	_, err = s.db.Exec("INSERT INTO profile_events (pubkey, id, date) VALUES ($1, $2, $3)",
		event.PubKey, event.ID, event.CreatedAt.Unix())
	if err != nil && err != sql.ErrNoRows {
		log.Printf("failed to save profile event: %s", err.Error())
	}
}

func (s *Store) AddEventToRepliesLists(event *nostr.Event) {
	for _, ref := range event.Tags.GetAll([]string{"e", ""}) {
		var count int
		err := s.db.Get(&count, "SELECT count(*) FROM reply_events WHERE root = $1", ref.Value())
		if err != nil && err != sql.ErrNoRows {
			log.Printf("failed to count reply events: %s", err.Error())
		}
		if count >= maxPerProfile {
			var oldest int64
			err := s.db.Get(&oldest, "SELECT min(date) FROM reply_events WHERE root = $1", ref.Value())
			if err != nil && err != sql.ErrNoRows {
				log.Printf("failed to get oldest for replies: %s", err.Error())
			}
			if oldest < event.CreatedAt.Unix() {
				return
			}
			_, err = s.db.Exec("DELETE FROM reply_events WHERE root = $1 AND date = $2", ref.Value(), oldest)
			if err != nil && err != sql.ErrNoRows {
				log.Printf("failed to cleanup reply cache: %s", err.Error())
			}
		}
		_, err = s.db.Exec("INSERT INTO reply_events (root, id, date) VALUES ($1, $2, $3)",
			ref.Value(), event.ID, event.CreatedAt.Unix())
		if err != nil && err != sql.ErrNoRows {
			log.Printf("failed to cache reply event: %s", err.Error())
		}
	}
}

func (s *Store) SaveUnencryptedEventOnChatHistory(chatId string, event *nostr.Event) {
	return
}

func (s *Store) GetEvent(ctx context.Context, id string) *nostr.Event {
	var b []byte
	var evt nostr.Event
	err := s.db.GetContext(ctx, &b, "SELECT body FROM events WHERE id = $1", id)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("failed to get event: %s", err.Error())
	}
	json.Unmarshal(b, &evt)
	return &evt
}

func (s *Store) GetReplaceableEvent(ctx context.Context, pubkey string, kind int) *nostr.Event {
	var id string
	err := s.db.GetContext(ctx, &id, "SELECT id FROM replaceable_events WHERE pubkey = $1 AND kind = $2",
		pubkey, kind)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("failed to get replaceable event: %s", err.Error())
	}
	return s.GetEvent(ctx, id)
}

func (s *Store) GetProfileEvents(pubkey string, until int, limit int) []*nostr.Event {
	ids := make([]string, 0, limit)
	err := s.db.Select(&ids,
		"SELECT id FROM profile_events WHERE pubkey = $1 AND date < $2 ORDER BY date DESC LIMIT $3",
		pubkey, until, limit)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("failed to profile events: %s", err.Error())
	}
	events := make([]*nostr.Event, 0, len(ids))
	for _, id := range ids {
		evt := s.GetEvent(context.Background(), id)
		if evt != nil {
			events = append(events, evt)
		}
	}
	return events
}

func (s *Store) GetReplies(id string, until int, limit int) []*nostr.Event {
	ids := make([]string, 0, limit)
	err := s.db.Select(&ids,
		"SELECT id FROM reply_events WHERE root = $1 AND date < $2 ORDER BY date DESC LIMIT $3",
		id, until, limit)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("failed to get replies: %s", err.Error())
	}
	events := make([]*nostr.Event, 0, len(ids))
	for _, id := range ids {
		evt := s.GetEvent(context.Background(), id)
		if evt != nil {
			events = append(events, evt)
		}
	}
	return events
}

func (s *Store) GetChatEvents(chatId string, until int, limit int) []*nostr.Event {
	return nil
}

func (s *Store) GetMostRecentChats(pubkey string) []*nostr.Event {
	return nil
}

func (s *Store) GetLatestTimestampForReplies(id string) *time.Time {
	var last int64
	err := s.db.Get(&last, "SELECT min(date) FROM reply_events WHERE root = $1", id)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("failed to get latest date for replies: %s", err.Error())
		return nil
	}
	ts := time.Unix(last, 0)
	return &ts
}

func (s *Store) GetLatestTimestampForProfile(pubkey string) *time.Time {
	var last int64
	err := s.db.Get(&last, "SELECT min(date) FROM profile_events WHERE pubkey = $1", pubkey)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("failed to get latest date for profile: %s", err.Error())
		return nil
	}
	ts := time.Unix(last, 0)
	log.Print("latest ", pubkey, " ", ts.Unix())
	return &ts
}

func (s *Store) GetLatestTimestampForChatMessages(chatId string) *time.Time {
	return nil
}

func (s *Store) FollowKey(_, followedKey string) error {
	_, err := s.db.Exec("INSERT INTO follows (pubkey) VALUES ($1)", followedKey)
	return err
}

func (s *Store) GetFollowedKeys(_ string) []string {
	var keys []string
	s.db.Select(&keys, "SELECT pubkey FROM follows")
	return keys
}
