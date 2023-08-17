package main

import (
	"context"

	"github.com/nbd-wtf/go-nostr"
)

type Follow struct {
	Pubkey  string
	Relay   string
	Petname string
}

func loadContactList(ctx context.Context, pubkey string) *[]Follow {
	if follows, ok := contactListsCache.Get(pubkey); ok {
		return follows
	}

	if evt := loadReplaceableEvent(ctx, pubkey, 3); evt == nil {
		log.Debug().Str("pubkey", pubkey).Msg("failed to load contact list event")
		return nil
	} else {
		follows := make([]Follow, 0, len(evt.Tags))
		for _, tag := range evt.Tags {
			if len(tag) >= 2 && tag[0] == "p" {
				follow := Follow{Pubkey: tag[1], Relay: tag.Relay()}
				if len(tag) >= 4 {
					follow.Petname = tag[3]
				}
				follows = append(follows, follow)
			}
		}
		contactListsCache.Set(pubkey, &follows, 1)
		return &follows
	}
}

func loadRelaysList(ctx context.Context, pubkey string) (read []string, write []string) {
	if evt := loadReplaceableEvent(ctx, pubkey, 10002); evt != nil {
		for _, tag := range evt.Tags {
			if len(tag) < 2 {
				continue
			}

			relay := nostr.NormalizeURL(tag[1])

			if len(tag) > 2 && tag[2] == "read" {
				read = append(read, relay)
			} else if len(tag) > 2 && tag[2] == "write" {
				write = append(write, relay)
			} else if len(tag) == 2 {
				read = append(read, relay)
				write = append(write, relay)
			}
		}
	}

	return read, write
}
