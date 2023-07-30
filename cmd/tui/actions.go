package main

import (
	"context"
	"log"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nbd-wtf/go-nostr"
)

func initialFetch(m *Model) func() tea.Msg {
	return func() tea.Msg {
		go func() {
			for event := range SubscribeHomeFeed(context.Background(), config.PublicKey) {
				log.Print("event: ", []any{event}, " ", event.ID[0:3])
				program.Send(event)
			}
		}()

		events := GetCachedHomeFeedEvents(config.PublicKey, 99999999999999, 100)
		homefeed, _ := m.homefeed.Update(events)
		return homefeed
	}
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

		return PublishToRelays(context.Background(), evt, config.WriteableRelays)
	}
}
