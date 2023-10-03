package main

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip10"
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
	ie := pool.QuerySingle(ctx, relays, filter)
	cancel()

	if ie != nil {
		store.SaveEvent(ctx, ie.Event)
		eventCache.Set(ie.Event.ID, ie.Event, 1)
	} else {
		// cache this even if it's nil so we don't keep trying to fetch it
		eventCache.SetWithTTL(ie.Event.ID, ie.Event, 1, CACHE_TTL_NOT_FOUND)
	}

	return ie.Event
}

func getStatusHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/api/v1/statuses/"):]
	evt := loadEvent(r.Context(), id, nil, nil)
	if evt == nil {
		jsonError(w, "couldn't find event", 404)
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
	data := createStatusBody{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		jsonError(w, "invalid request data", 400)
		return
	}

	if data.Visibility != "public" {
		jsonError(w, "only public supported for now", 422)
		return
	}

	if data.Poll != nil {
		jsonError(w, "polls are not supported", 422)
		return
	}

	if len(data.MediaIds) > 0 {
		jsonError(w, "media uploads not yet supported", 422)
		return
	}

	evt := &nostr.Event{
		CreatedAt: nostr.Now(),
		Kind:      1,
		Tags:      make(nostr.Tags, 0, 4),
		Content:   data.Status,
	}

	if data.Sensitive && data.SpoilerText != "" {
		evt.Tags = append(evt.Tags, nostr.Tag{"content-warning", data.SpoilerText})
	} else if data.Sensitive {
		evt.Tags = append(evt.Tags, nostr.Tag{"content-warning"})
	} else if data.SpoilerText != "" {
		evt.Tags = append(evt.Tags, nostr.Tag{"subject", data.SpoilerText})
	}

	if data.InReplyToId != "" {
		// try to fetch the event we're repĺying to
		parent := loadEvent(r.Context(), data.InReplyToId, nil, nil)
		if parent == nil {
			evt.Tags = append(evt.Tags, nostr.Tag{"e", data.InReplyToId, "reply"})
		} else {
			root := nip10.GetThreadRoot(evt.Tags)
			if root == nil {
				root = nip10.GetImmediateReply(evt.Tags)
			}
			if root != nil {
				root = &nostr.Tag{"e", (*root)[1], "root"}
			}

			// copy 'p' tags
			totalPs := 0
			for _, tag := range evt.Tags {
				if len(tag) < 2 {
					continue
				}
				if tag[0] == "p" && totalPs < 4 {
					// TODO include better hint for p if we have one
					evt.Tags.AppendUnique(tag)
					totalPs++
				}
			}

			// if root exists
			// TODO include hints for all of these
			if root != nil {
				// include root
				evt.Tags = append(evt.Tags, nostr.Tag{"e", (*root)[1], "", "root"})
				// and then the parent as "reply"
				evt.Tags = append(evt.Tags, nostr.Tag{"e", parent.ID, "", "reply"})
			} else {
				// include parent as root
				evt.Tags = append(evt.Tags, nostr.Tag{"e", parent.ID, "", "root"})
			}
		}
	}

	switch data.Visibility {
	case "public":
		var err error
		evt, err = publish(r.Context(), evt)
		if err != nil {
			jsonError(w, err.Error(), 500)
			return
		}
	case "unlisted":
	case "private":
	case "direct":
	}

	json.NewEncoder(w).Encode(toStatus(r.Context(), evt))
}
