package main

import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nbd-wtf/go-nostr"
)

func initialFetch() tea.Msg {
	return nil
}

func publishNote(text string) tea.Cmd {
	return func() tea.Msg {
		evt := nostr.Event{
			PubKey:    config.PublicKey,
			Content:   text,
			Kind:      1,
			CreatedAt: time.Now(),
			Tags:      nostr.Tags{},
		}
		evt.ID = evt.GetID()
		err := evt.Sign(config.PrivateKey)
		if err != nil {
			return err
		}

		return nr.PublishToRelays(context.Background(), evt, config.WriteableRelays)
	}
}
