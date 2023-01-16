package main

import (
	"context"
	"log"
	"path/filepath"
	"time"

	"github.com/cayleygraph/cayley"
	"github.com/cayleygraph/cayley/graph"
	_ "github.com/cayleygraph/cayley/graph/kv/bolt"
	"github.com/nbd-wtf/go-nostr"
)

type Store struct {
	db *cayley.Handle
}

func InitStore(datadir string) *Store {
	dbpath := filepath.Join(datadir, "bcayley")
	graph.InitQuadStore("bolt", dbpath, nil)
	db, err := cayley.NewMemoryGraph()
	if err != nil {
		log.Fatalln(err)
	}
	return &Store{db: db}
}

func (s Store) IncrementRelayScoreForPubkey(pubkey string, relay string, score int) error {
	return nil
}

func (s Store) FetchTopRelaysForPubkey(pubkey string, limit int) []string {
	return []string{"wss://nostr-pub.wellorder.net"}
}

func (s Store) CacheSingleEvent(event *nostr.Event) {
}

func (s Store) CacheReplaceableEvent(event *nostr.Event) {
}

func (s Store) AddEventToProfileList(event *nostr.Event) {
}

func (s Store) AddEventToRepliesLists(event *nostr.Event) {
}

func (s Store) SaveUnencryptedEventOnChatHistory(chatId string, event *nostr.Event) {
}

func (s Store) GetEvent(ctx context.Context, id string) *nostr.Event {
	return nil
}

func (s Store) GetReplaceableEvent(ctx context.Context, pubkey string, kind int) *nostr.Event {
	return nil
}

func (s Store) GetProfileEvents(pubkey string, until int, limit int) []*nostr.Event {
	return nil
}

func (s Store) GetReplies(id string, until int, limit int) []*nostr.Event {
	return nil
}

func (s Store) GetChatEvents(chatId string, until int, limit int) []*nostr.Event {
	return nil
}

func (s Store) GetMostRecentChats(pubkey string) []*nostr.Event {
	return nil
}

func (s Store) GetLatestTimestampForProfile(pubkey string) *time.Time {
	return nil
}

func (s Store) GetLatestTimestampForReplies(id string) *time.Time {
	return nil
}

func (s Store) GetLatestTimestampForChatMessages(chatId string) *time.Time {
	return nil
}

func (s Store) GetFollowedKeys(pubkey string) []string {
	return nil
}
