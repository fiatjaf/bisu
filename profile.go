package main

import (
	"context"
	"encoding/json"

	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip19"
)

type Profile struct {
	pubkey string
	event  *nostr.Event

	Name        string `json:"name,omitempty"`
	About       string `json:"about,omitempty"`
	NIP05       string `json:"nip05,omitempty"`
	LUD06       string `json:"lud06,omitempty"`
	LUD16       string `json:"lud16,omitempty"`
	Banner      string `json:"banner,omitempty"`
	Picture     string `json:"picture,omitempty"`
	Website     string `json:"website,omitempty"`
	DisplayName string `json:"display_name,omitempty"`

	// this will be nil when we haven't attempted to fetch yet or if it's invalid
	validatedNip05 *string `json:"-"`
}

func (p Profile) handle() string {
	handle := p.Name
	if handle == "" {
		npub, _ := nip19.EncodePublicKey(p.event.PubKey)
		handle = npub
	}
	return handle
}

func loadProfile(ctx context.Context, pubkey string) *Profile {
	if profile, ok := metadataCache.Get(pubkey); ok {
		return profile
	}

	metadataEvent := loadReplaceableEvent(ctx, pubkey, 0)
	if metadataEvent == nil {
		log.Debug().Str("pubkey", pubkey).Msg("failed to load metadata event, storing nil on cache")
		metadataCache.SetWithTTL(pubkey, &Profile{pubkey: pubkey}, 1, CACHE_TTL_NOT_FOUND)
		return nil
	}

	metadata := toProfile(metadataEvent)
	log.Debug().Interface("profile", metadata).Msg("found metadata, storing on cache")
	metadataCache.Set(pubkey, metadata, 1)
	return metadata
}

func toProfile(metadataEvent *nostr.Event) *Profile {
	metadata := &Profile{
		pubkey: metadataEvent.PubKey,
		event:  metadataEvent,
	}

	if err := json.Unmarshal([]byte(metadata.event.Content), metadata); err != nil {
		log.Debug().Str("event", metadata.event.String()).Msg("metadata event has invalid json, storing nil on cache")
		metadataCache.SetWithTTL(metadata.pubkey, metadata, 1, CACHE_TTL_NOT_FOUND)
		return nil
	}

	return metadata
}
