package main

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/nbd-wtf/go-nostr"
	"mvdan.cc/xurls/v2"
)

var urlMatcher = xurls.Strict()

func loadEvent(ctx context.Context, id string, relayHints []string, authorHint *string) *nostr.Event {
	if evt, ok := eventCache.Get(id); ok {
		return evt
	}

	filter := nostr.Filter{IDs: []string{id}}

	if ch, err := store.QueryEvents(ctx, filter); err == nil {
		if evt := <-ch; evt != nil {
			eventCache.Set(evt.ID, evt, 1)
			return evt
		}
	}

	max := len(relayHints)
	if max < 3 {
		max = 3
	}

	relays := make([]string, 0, max)
	for i := 0; i < max; i++ {
		var relay string
		if len(relayHints) > i {
			relay = relayHints[i]
		} else {
			serial++
			relay = defaultRelays[serial&len(defaultRelays)]
		}
		relays = append(relays, relay)
	}

	if authorHint != nil {
		// TODO: gather relays from NIP-65
	}

	ctx, cancel := context.WithTimeout(ctx, time.Second*4)
	evt := pool.QuerySingle(ctx, relays, filter)
	cancel()

	if evt != nil {
		store.SaveEvent(ctx, evt)
		eventCache.Set(evt.ID, evt, 1)
	} else {
		// cache this even if it's nil so we don't keep trying to fetch it
		eventCache.SetWithTTL(evt.ID, evt, 1, CACHE_TTL_NOT_FOUND)
	}

	return evt
}

func getStatusHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/api/v1/statuses/"):]
	evt := loadEvent(r.Context(), id, nil, nil)
	if evt == nil {
		http.Error(w, "couldn't find event", 404)
		return
	}
	status := toStatus(r.Context(), evt)
	json.NewEncoder(w).Encode(status)
}

type createStatusBody struct {
	InReplyToId string   `json:"inReplyToId"`
	Language    string   `json:"language"`
	MediaIds    []string `json:"mediaIds"`
	Poll        *struct {
		Options    string `json:"options"`
		ExpiresIn  int    `json:"expiresIn"`
		Multiple   bool   `json:"multiple"`
		HideTotals bool   `json:"hideTotals"`
	} `json:"poll"`
	ScheduledAt string `json:"scheduledAt"`
	Sensitive   bool   `json:"sensitive"`
	SpoilerText string `json:"spoilerText"`
	Status      string `json:"status"`
	Visibility  string `json:"visibility"`
}

func createStatusHandler(w http.ResponseWriter, r *http.Request) {
	body := createStatusBody{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request body", 400)
		return
	}

	if body.Visibility != "public" {
		http.Error(w, "only public supported for now", 422)
		return
	}

	if body.Poll != nil {
		http.Error(w, "polls are not supported", 422)
		return
	}

	if len(body.MediaIds) > 0 {
		http.Error(w, "media uploads not yet supported", 422)
		return
	}

	evt := nostr.Event{
		CreatedAt: nostr.Now(),
		Kind:      1,
		Tags:      make(nostr.Tags, 0, 4),
	}

	if body.Sensitive && body.SpoilerText != "" {
		evt.Tags = append(evt.Tags, nostr.Tag{"content-warning", body.SpoilerText})
	} else if body.Sensitive {
		evt.Tags = append(evt.Tags, nostr.Tag{"content-warning"})
	} else if body.SpoilerText != "" {
		evt.Tags = append(evt.Tags, nostr.Tag{"subject", body.SpoilerText})
	}

	if err := evt.Sign(sk); err != nil {
		http.Error(w, "failed to sign event", 500)
		return
	}

	if err := store.SaveEvent(r.Context(), &evt); err != nil {
		http.Error(w, "failed to save event", 500)
		return
	}

	switch body.Visibility {
	case "public":
		for _, relay := range writeRelays {
			r, err := pool.EnsureRelay(relay)
			if err != nil {
				log.Warn().Err(err).Str("relay", relay).Msg("failed to ensure relay when publishing")
				continue
			}
			status, err := r.Publish(r.Context(), evt)
			if err == nil && status == nostr.PublishStatusSucceeded {
				log.Debug().Str("id", evt.ID).Str("relay", relay).Msg("event published")
			} else {
				log.Warn().Str("id", evt.ID).Str("relay", relay).Err(err).Msg("event probably failed to be published")
			}
		}
	case "unlisted":
	case "private":
	case "direct":
	}
}
