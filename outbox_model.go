package main

import (
	"context"

	"github.com/nbd-wtf/go-nostr"
)

func fetchOutboxRelaysForUser(ctx context.Context, pubkey string, n int, strict bool) []string {
	// TODO
	return []string{}
}

func fetchInboxRelaysForUser(ctx context.Context, pubkey string, n int, strict bool) []string {
	// TODO
	return []string{}
}

func saveLastAttempted(ctx context.Context, pubkey string, relay string) {
	// TODO
}

func grabRelaysFromEvent(ctx context.Context, evt *nostr.Event) {
	// TODO
}

func saveLastFetched(ctx context.Context, evt *nostr.Event, relay string) {
	// TODO
}
