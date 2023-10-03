package main

import (
	"context"
	"fmt"

	"github.com/nbd-wtf/go-nostr"
)

func publish(ctx context.Context, evt *nostr.Event) (*nostr.Event, error) {
	if evt.Tags == nil {
		evt.Tags = nostr.Tags{}
	}

	if err := evt.Sign(sk); err != nil {
		return nil, fmt.Errorf("failed to sign event")
	}

	if err := store.SaveEvent(ctx, evt); err != nil {
		return nil, fmt.Errorf("failed to save event")
	}

	successes := 0
	for _, relay := range writeRelays {
		r, err := pool.EnsureRelay(relay)
		if err != nil {
			log.Warn().Err(err).Str("relay", relay).Msg("failed to ensure relay when publishing")
			continue
		}
		status, err := r.Publish(r.Context(), *evt)
		if err == nil && status == nostr.PublishStatusSucceeded {
			log.Debug().Str("id", evt.ID).Str("relay", relay).Msg("event published")
			successes++
		} else {
			log.Warn().Str("id", evt.ID).Str("relay", relay).Err(err).Msg("event probably failed to be published")
		}
	}

	if successes == 0 {
		return nil, fmt.Errorf("failed to publish to any relay of %v", writeRelays)
	}

	return evt, nil
}
