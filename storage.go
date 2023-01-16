package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"log"
	"path/filepath"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/nbd-wtf/go-nostr"
	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schema string

type Store struct {
	db *sqlx.DB
}

func InitStore(datadir string) *Store {
	dbpath := filepath.Join(datadir, "bisu")
	db, err := sqlx.Connect("sqlite", dbpath)
	if err != nil {
		log.Fatalln(err)
	}

	db.MustExec(schema)

	return &Store{db: db}
}

const (
	maxPerProfile = 80
)

func (s *Store) IncrementRelayScoreForPubkey(pubkey string, relay string, score int) error {
	s.db.Exec("INSERT INTO relay_scores (pubkey, url, score) VALUES ($1, $2, $3) ON CONFLICT (pubkey, url) DO UPDATE SET score = score + $3",
		pubkey, relay, score)
	return nil
}

func (s *Store) FetchTopRelaysForPubkey(pubkey string, limit int) []string {
	var relays []string
	s.db.Select(&relays, "SELECT url FROM relay_scores WHERE pubkey = $1 LIMIT $2", pubkey, limit)
	return relays
}

func (s *Store) CacheSingleEvent(event *nostr.Event) {
	b, _ := json.Marshal(event)
	s.db.Exec("INSERT OR IGNORE INTO events (id, body) VALUES ($1, $2)", event.ID, b)
}

func (s *Store) CacheReplaceableEvent(event *nostr.Event) {
	var b []byte
	s.db.Get(&b, "SELECT body FROM event WHERE pubkey = $1 AND kind = $2", event.PubKey, event.Kind)
	var curr nostr.Event
	var currId string
	if json.Unmarshal(b, &curr); curr.CreatedAt.Before(event.CreatedAt) {
		s.CacheSingleEvent(event)
		currId = event.ID
	} else {
		currId = curr.ID
	}
	s.db.Exec("INSERT OR IGNORE INTO replaceable_events (pubkey, kind, id) VALUES ($1, $2, $3)",
		event.PubKey, event.Kind, currId)
}

func (s *Store) AddEventToProfileList(event *nostr.Event) {
	var count int
	s.db.Get(&count, "SELECT count(*) FROM profile_events WHERE pubkey = $1", event.PubKey)
	if count >= maxPerProfile {
		var oldest int64
		s.db.Get(&oldest, "SELECT max(date) FROM profile_events WHERE pubkey = $1", event.PubKey)
		if oldest < event.CreatedAt.Unix() {
			return
		}
		s.db.Exec("DELETE FROM profile_events WHERE pubkey = $1 AND date = $2", event.PubKey, oldest)
	}
	s.db.Exec("INSERT INTO profile_events (pubkey, id, date) VALUES ($1, $2, $3)",
		event.PubKey, event.ID, event.CreatedAt.Unix())
}

func (s *Store) AddEventToRepliesLists(event *nostr.Event) {
	for _, ref := range event.Tags.GetAll([]string{"e", ""}) {
		var count int
		s.db.Get(&count, "SELECT count(*) FROM reply_events WHERE root = $1", ref.Value())
		if count >= maxPerProfile {
			var oldest int64
			s.db.Get(&oldest, "SELECT max(date) FROM reply_events WHERE root = $1", ref.Value())
			if oldest < event.CreatedAt.Unix() {
				return
			}
			s.db.Exec("DELETE FROM reply_events WHERE root = $1 AND date = $2", ref.Value(), oldest)
		}
		s.db.Exec("INSERT INTO reply_events (root, id, date) VALUES ($1, $2, $3)",
			ref.Value(), event.ID, event.CreatedAt.Unix())
	}
}

func (s *Store) SaveUnencryptedEventOnChatHistory(chatId string, event *nostr.Event) {
	return
}

func (s *Store) GetEvent(ctx context.Context, id string) *nostr.Event {
	var b []byte
	var evt nostr.Event
	s.db.GetContext(ctx, &b, "SELECT body FROM events WHERE id = $1", id)
	json.Unmarshal(b, &evt)
	return &evt
}

func (s *Store) GetReplaceableEvent(ctx context.Context, pubkey string, kind int) *nostr.Event {
	var id string
	s.db.GetContext(ctx, &id, "SELECT id FROM replaceable_events WHERE pubkey = $1 AND kind = $2",
		pubkey, kind)
	return s.GetEvent(ctx, id)
}

func (s *Store) GetProfileEvents(pubkey string, until int, limit int) []*nostr.Event {
	ids := make([]string, 0, limit)
	s.db.Select(ids, "SELECT id FROM profile_events WHERE pubkey = $1 AND date < $2 ORDER BY date DESC LIMIT $3",
		pubkey, until, limit)
	events := make([]*nostr.Event, len(ids))
	for i, id := range ids {
		events[i] = s.GetEvent(context.Background(), id)
	}
	return events
}

func (s *Store) GetReplies(id string, until int, limit int) []*nostr.Event {
	ids := make([]string, 0, limit)
	s.db.Select(ids, "SELECT id FROM reply_events WHERE root = $1 AND date < $2 ORDER BY date DESC LIMIT $3",
		id, until, limit)
	events := make([]*nostr.Event, len(ids))
	for i, id := range ids {
		events[i] = s.GetEvent(context.Background(), id)
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
	s.db.Get(&last, "SELECT min(date) FROM reply_events WHERE root = $1", id)
	ts := time.Unix(last, 0)
	return &ts
}

func (s *Store) GetLatestTimestampForProfile(pubkey string) *time.Time {
	var last int64
	s.db.Get(&last, "SELECT min(date) FROM profile_events WHERE pubkey = $1", pubkey)
	ts := time.Unix(last, 0)
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
