package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nbd-wtf/go-nostr/nip19"
)

func handleCommand(text string) (tea.Cmd, error) {
	var key string
	var relays []string
	spl := strings.Split(text, " ")
	if spl[0] == "/follow" {
		prefix, value, err := nip19.Decode(spl[1])
		if (prefix == "nprofile" || prefix == "npub") && err == nil {
			switch t := value.(type) {
			case string:
				key = t
			case nip19.ProfilePointer:
				key = t.PublicKey
				relays = t.Relays
			}
		} else {
			return nil, fmt.Errorf("invalid command: %w", err)
		}
	}

	return func() tea.Msg {
		store.FollowKey(config.PublicKey, key)
		for _, url := range relays {
			store.IncrementRelayScoreForPubkey(config.PublicKey, url, 1)
		}
		return UpdateFollows
	}, nil
}
