package main

import (
	"context"
)

type Follow struct {
	Pubkey  string `json:"pubkey"`
	Relay   string `json:"relay"`
	Petname string `json:"petname"`
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
