package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/nbd-wtf/go-nostr"
)

func fetchInboxRelaysForUser(ctx context.Context, pubkey string, n int, strict bool) []string {
	relays := make([]string, 0, n)
	db.SelectContext(ctx, &relays, `
SELECT relay
FROM pubkey_relays
WHERE pubkey = $1
  AND last_kind3_inbox > 0
  OR last_nip65_inbox > 0
ORDER BY greatest(last_kind3_inbox, last_nip65_inbox)
LIMIT $2
    `, pubkey, n)

	if !strict {
		// fill in with relays that everybody uses
		for len(relays) < n {
			n++
			relays = append(relays, defaultRelays[n%len(defaultRelays)])
		}
	}

	return relays
}

func fetchOutboxRelaysForUser(ctx context.Context, pubkey string, n int, strict bool) []string {
	strictLine := ""
	if strict {
		strictLine = " AND (last_fetched_success > 0 OR last_nip65_outbox > 0 OR last_nip65_outbox > 0 OR last_kind3_outbox > 0 OR last_hint_tag > 0 OR last_hint_nprofile > 0 OR last_hint_nip05 > 0)"
	}

	relays := make([]string, 0, n+5)
	db.SelectContext(ctx, &relays, `
SELECT relay
FROM pubkey_relays
WHERE pubkey = $1`+strictLine+`
ORDER BY (
  last_fetched_attempt * -5 + -- this has negative power
  last_fetched_success * 5 + -- which should be compensated by this
  last_nip65_outbox * 2.5 + -- nip65 adoption gives us a very certain outcome
  last_kind3_outbox * 2 + -- but kind3 is not that bad either
  last_hint_tag * 0.5 + -- hints are not worth a lot
  last_hint_nprofile * 0.7 +
  last_hint_nip05 * 1 + -- except when is set by the people themselves
  random() * extract(epoch from now())
) DESC
LIMIT $2
    `, pubkey, n)

	if !strict {
		// fill in with relays that everybody uses
		for len(relays) < n {
			relays = append(relays, defaultRelays[n%len(defaultRelays)])
		}
	}

	return relays
}

func saveLastAttempted(ctx context.Context, pubkey string, relay string) {
	serial++
	if serial%10 != 0 {
		// skip 9 of 10 write attempts to this
		return
	}

	_, err := db.ExecContext(ctx, `SELECT set_if_more_recent($1, $2, 'last_fetched_attempt', $3)`,
		pubkey, nostr.NormalizeURL(relay), time.Now().Unix())
	if err != nil {
		log.Error().Err(err).Str("pubkey", pubkey).Str("relay", relay).Msg("failed to save last attempted")
	}
}

func grabRelaysFromEvent(ctx context.Context, evt *nostr.Event) {
	for _, tag := range evt.Tags {
		if len(tag) >= 3 {
			if tag[0] == "p" {
				if relay := nostr.NormalizeURL(tag[2]); relay != "" {
					if nostr.IsValidPublicKeyHex(tag[1]) {
						saveTagHint(ctx, tag[1], relay, evt.CreatedAt)
					}
				}
			}
		}
	}

	if evt.Kind == 3 {
		relays := make(map[string]struct {
			Write bool `json:"write"`
			Read  bool `json:"read"`
		})
		if err := json.Unmarshal([]byte(evt.Content), &relays); err == nil {
			for relay, policy := range relays {
				relay = nostr.NormalizeURL(relay)
				if relay == "" || relay == "wss://feeds.nostr.band" || strings.HasPrefix(relay, "wss://filter.nostr.wine") {
					continue
				}
				if policy.Read {
					saveKind3Inbox(ctx, evt.PubKey, relay, evt.CreatedAt)
				}
				if policy.Write {
					saveKind3Outbox(ctx, evt.PubKey, relay, evt.CreatedAt)
				}
			}
		}
	}

	if evt.Kind == 10002 {
		for _, tag := range evt.Tags {
			if len(tag) >= 1 && tag[0] == "r" {
				if len(tag) >= 2 {
					relay := tag[1]
					relay = nostr.NormalizeURL(relay)
					if relay == "" {
						continue
					}
					if relay == "" || relay == "wss://feeds.nostr.band" || strings.HasPrefix(relay, "wss://filter.nostr.wine") {
						continue
					}

					read := true
					write := true

					if len(tag) >= 3 {
						if tag[2] == "read" {
							write = false
						}
						if tag[2] == "write" {
							read = false
						}
					}

					if read {
						saveNip65Inbox(ctx, evt.PubKey, relay, evt.CreatedAt)
					}
					if write {
						saveNip65Outbox(ctx, evt.PubKey, relay, evt.CreatedAt)
					}
				}
			}
		}
	}
}

func saveLastFetched(ctx context.Context, pubkey string, relay string) {
	serial++
	if serial%10 != 0 {
		// skip 9 of 10 write attempts to this
		return
	}

	_, err := db.ExecContext(
		ctx,
		`SELECT set_if_more_recent($1, $2, 'last_fetched_success', $3)`,
		pubkey, nostr.NormalizeURL(relay), nostr.Now())
	if err != nil {
		log.Error().Err(err).Str("pubkey", pubkey).Str("relay", relay).Msg("failed to save last attempted")
	}
}

func getLastFetched(ctx context.Context, pubkey string, relay string) *nostr.Timestamp {
	return nil
}

func saveNprofileHint(ctx context.Context, pubkey string, relay string, when nostr.Timestamp) {
	_, err := db.ExecContext(
		ctx,
		`SELECT set_if_more_recent($1, $2, 'last_hint_nprofile', $3)`,
		pubkey, nostr.NormalizeURL(relay), when)
	if err != nil {
		log.Error().Err(err).Str("pubkey", pubkey).Str("relay", relay).Msg("failed to save nprofile hint")
	}
}

func saveNip05Hint(ctx context.Context, pubkey string, relay string, when nostr.Timestamp) {
	_, err := db.ExecContext(
		ctx,
		`SELECT set_if_more_recent($1, $2, 'last_hint_nip05', $3)`,
		pubkey, nostr.NormalizeURL(relay), when)
	if err != nil {
		log.Error().Err(err).Str("pubkey", pubkey).Str("relay", relay).Msg("failed to save nip05 hint")
	}
}

func saveTagHint(ctx context.Context, pubkey string, relay string, when nostr.Timestamp) {
	_, err := db.ExecContext(
		ctx,
		`SELECT set_if_more_recent($1, $2, 'last_hint_tag', $3)`,
		pubkey, relay, when)
	if err != nil {
		log.Error().Err(err).Str("pubkey", pubkey).Str("relay", relay).Msg("failed to save tag hint")
	}
}

func saveNip65Outbox(ctx context.Context, pubkey string, relay string, when nostr.Timestamp) {
	_, err := db.ExecContext(
		ctx,
		`SELECT set_if_more_recent($1, $2, 'last_nip65_outbox', $3)`,
		pubkey, relay, when)
	if err != nil {
		log.Error().Err(err).Str("pubkey", pubkey).Str("relay", relay).Msg("failed to save nip65 outbox")
	}
}

func saveNip65Inbox(ctx context.Context, pubkey string, relay string, when nostr.Timestamp) {
	_, err := db.ExecContext(
		ctx,
		`SELECT set_if_more_recent($1, $2, 'last_nip65_inbox', $3)`,
		pubkey, relay, when)
	if err != nil {
		log.Error().Err(err).Str("pubkey", pubkey).Str("relay", relay).Msg("failed to save nip65 inbox")
	}
}

func saveKind3Outbox(ctx context.Context, pubkey string, relay string, when nostr.Timestamp) {
	_, err := db.ExecContext(
		ctx,
		`SELECT set_if_more_recent($1, $2, 'last_kind3_outbox', $3)`,
		pubkey, relay, when)
	if err != nil {
		log.Error().Err(err).Str("pubkey", pubkey).Str("relay", relay).Msg("failed to save kind3 outbox")
	}
}

func saveKind3Inbox(ctx context.Context, pubkey string, relay string, when nostr.Timestamp) {
	_, err := db.ExecContext(
		ctx,
		`SELECT set_if_more_recent($1, $2, 'last_kind3_inbox', $3)`,
		pubkey, relay, when)
	if err != nil {
		log.Error().Err(err).Str("pubkey", pubkey).Str("relay", relay).Msg("failed to save kind3 inbox")
	}
}

func setIfMoreRecent(ctx context.Context, pubkey string, relay string, column string, when nostr.Timestamp) error {
	if when == 0 {
		return fmt.Errorf("when is zero")
	}

	var current int64
	err := db.GetContext(ctx, &current,
		`SELECT `+column+` FROM pubkey_relays WHERE pubkey = $1 AND relay = $2`,
		pubkey, relay)
	if err != nil {
		return err
	}

	if int64(when) > current {
		_, err = db.ExecContext(ctx, `INSERT INTO relays (pubkey, relay, `+column+`) VALUES ($1, $2, $3)`,
			pubkey, relay, when)
		fmt.Println("~>", err)
	}
	return err
}

func startListening() {
	ctx := context.Background()

	pfollows := loadContactList(ctx, profile.pubkey)
	var follows []Follow
	if pfollows != nil {
		follows = *pfollows
	}
	follows = append(follows, Follow{Pubkey: profile.pubkey})

	log.Debug().Int("n", len(follows)).Msg("listening to notes from all the people we follow")
	queries := make(map[string]nostr.Filter)
	for _, follow := range follows {
		relays := fetchOutboxRelaysForUser(ctx, follow.Pubkey, 3, false)
		for _, r := range relays {
			filter, ok := queries[r]
			if !ok {
				filter.Kinds = []int{1}
				filter.Authors = make([]string, 0, 20)
				filter.Limit = 200
				now := nostr.Now()
				filter.Since = &now
			}
			filter.Authors = append(filter.Authors, follow.Pubkey)
			lastFetched := getLastFetched(ctx, follow.Pubkey, r)
			if lastFetched != nil && *lastFetched < *filter.Since {
				filter.Since = lastFetched
			}
			queries[r] = filter
		}
	}

	// dispatch all queries
	for r, filter := range queries {
		go func(r string, filter nostr.Filter) {
			relay, err := pool.EnsureRelay(r)
			if err != nil {
				log.Warn().Err(err).Str("relay", r).Msg("failed to connect")
				return
			}
			sub, err := relay.Subscribe(ctx, nostr.Filters{filter}, nostr.WithLabel("background"))
			if err != nil {
				log.Warn().Err(err).Str("relay", r).Stringer("filter", filter).Msg("failed to subscribe")
				return
			}

			for evt := range sub.Events {
				log.Debug().Stringer("event", evt).Msg("got event")
				store.SaveEvent(ctx, evt)
			}
		}(r, filter)
	}
}
