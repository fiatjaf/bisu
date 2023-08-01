package main

import (
	"context"

	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip19"
)

type Profile struct {
	event *nostr.Event

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

func loadProfile(ctx context.Context, pubkey string) (*Profile, error) {
	return &Profile{Name: "zzz"}, nil
}
