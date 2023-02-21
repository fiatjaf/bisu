package main

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"sort"
	"time"

	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip04"
)

// TODO: keep track of hint event relays, not only seen

func PublishToRelays(ctx context.Context, event nostr.Event, urls []string) error {
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 3*time.Second)
		defer cancel()
	}

	good := make(chan struct{})
	for _, url := range urls {
		go func(url string) {
			err := pool.publish(ctx, url, event)
			if err == nil {
				good <- struct{}{}
			}
		}(url)
	}

	select {
	case <-good:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("failed to publish on any of the relays")
	}
}

func GetReplaceableEvent(ctx context.Context, pubkey string, kind int) *nostr.Event {
	cached := store.GetReplaceableEvent(ctx, pubkey, kind)
	if cached != nil {
		return cached
	}

	// search in specific relays for user, then on fallback relays
	for _, relays := range [][]string{
		store.GetTopRelaysForPubkey(pubkey, 3),
		fallback(3),
	} {
		select {
		case <-ctx.Done():
			return nil
		default:
			if evt, err := race(
				func(relay string) (*nostr.Event, error) {
					events := pool.querySync(ctx, relay, nostr.Filters{
						nostr.Filter{
							Limit:   1,
							Kinds:   []int{kind},
							Authors: []string{pubkey},
						},
					})
					if len(events) == 0 {
						return nil, fmt.Errorf("not found on relay %s", relay)
					}

					sort.Slice(events, sortEventsDesc(events))
					return events[0], nil
				},
				relays,
			); err == nil {
				// we got this event from some relay
				// cache it
				store.CacheSingleEvent(evt)
				store.CacheReplaceableEvent(evt)

				// TODO increment relay data for user

				// then return
				return evt
			}
		}
	}

	return nil
}

func GetEvent(ctx context.Context, id string, relays ...string) *nostr.Event {
	cached := store.GetEvent(ctx, id)
	if cached != nil {
		return cached
	}

	// search on the relays that were given to us and if that fails search on fallback relays
	relaysGroups := make([][]string, 2)
	relaysGroups[0] = relays
	relaysGroups[1] = fallback(5)

	for _, relays := range relaysGroups {
		select {
		case <-ctx.Done():
			return nil
		default:
			if evt, err := race(
				func(relay string) (*nostr.Event, error) {
					events := pool.querySync(ctx, relay, nostr.Filters{
						nostr.Filter{IDs: []string{id}},
					})
					if len(events) == 0 {
						return nil, fmt.Errorf("not found on relay %s", relay)
					}

					evt := events[0]
					// store this url on it so it is cached with it for future use
					evt.SetExtra("relay", nostr.NormalizeURL(relay))

					return evt, nil
				},
				relays,
			); err == nil {
				// cache it
				store.CacheSingleEvent(evt)

				// TODO increment relay score for user

				// then return
				return evt
			}
		}
	}

	return nil
}

func GetCachedHomeFeedEvents(fromPubkey string, until int, limit int) []*nostr.Event {
	pubkeys := store.GetFollowedKeys(fromPubkey)

	if len(pubkeys) == 0 {
		return nil
	}

	// amount to try to fetch from each key
	nkeys := float64(len(pubkeys))
	fromEach := int(math.Ceil(float64(limit) * math.Pow(nkeys, 0.8) / nkeys)) // this will yield a reasonable number

	// do a separate query for each pubkey, in parallel
	ch := make(chan []*nostr.Event)
	waiting := len(pubkeys)
	for _, pk := range pubkeys {
		go func(pk string) {
			b := store.GetProfileEvents(pk, until, fromEach)
			ch <- b
			waiting--
			if waiting == 0 {
				close(ch)
			}
		}(pk)
	}

	// aggregate all results here
	events := make([]*nostr.Event, 0, fromEach*len(pubkeys))
	for batch := range ch {
		events = append(events, batch...)
	}

	// then sort and trim
	sort.Slice(events, sortEventsDesc(events))

	if len(events) <= limit {
		return events
	}
	return events[0:limit]
}

func GetProfileEvents(pubkey string, until int, limit int) []*nostr.Event {
	return store.GetProfileEvents(pubkey, until, limit)
}

func GetReplies(id string, until int, limit int) []*nostr.Event {
	return store.GetReplies(id, until, limit)
}

func GetChatMessages(pubkey1, pubkey2 string, until int, limit int) []*nostr.Event {
	return store.GetChatEvents(chatIdFromPubkeys(pubkey1, pubkey2), until, limit)
}

func GetMostRecentChats(pubkey string) []*nostr.Event {
	return store.GetMostRecentChats(pubkey)
}

func SubscribeHomeFeed(ctx context.Context, pubkey string) chan *nostr.Event {
	pubkeys := store.GetFollowedKeys(pubkey)
	events := pool.addHomeFeedListener(ctx, pubkeys)
	unique := softUniquenessFilter(events)
	return unique
}

func SubscribeProfileEvents(ctx context.Context, pubkey string) chan *nostr.Event {
	// subscribe and listen to relays
	relays := store.GetTopRelaysForPubkey(pubkey, 4)
	if len(relays) < 3 {
		relays = append(relays, fallback(3-len(relays))...)
	}

	return pool.subMany(ctx, relays, nostr.Filters{
		{
			Limit:   100, // only for the initial query
			Authors: []string{pubkey},
			Kinds:   []int{1},
			Since:   store.GetLatestTimestampForProfile(pubkey),
		},
	})
}

func SubscribeReplies(ctx context.Context, id string) chan *nostr.Event {
	// subscribe and listen to relays
	return pool.subMany(ctx, config.SafeRelays, nostr.Filters{
		{
			Limit: 100, // only for the initial query
			Tags:  nostr.TagMap{"e": []string{id}},
			Kinds: []int{1},
			Since: store.GetLatestTimestampForReplies(id),
		},
	})
}

func SubscribeChatEvents(ctx context.Context, to string, from string) chan *nostr.Event {
	// subscribe and listen to relays
	relays := store.GetTopRelaysForPubkey(from, 2)
	if len(relays) < 2 {
		relays = append(relays, fallback(3-len(relays))...)
	}

	// subscribe and listen to relays
	return pool.subMany(ctx, relays, nostr.Filters{
		{
			Limit:   100, // only for the initial query
			Tags:    nostr.TagMap{"p": []string{to}},
			Authors: []string{from},
			Kinds:   []int{4},
			Since:   store.GetLatestTimestampForChatMessages(chatIdFromPubkeys(to, from)),
		},
	})
}

func CacheEvent(evt *nostr.Event) {
	// cache them both individually and as part of their profile list of events and reply lists
	store.CacheSingleEvent(evt)

	// profile list
	store.AddEventToProfileList(evt)

	// replies lists
	store.AddEventToRepliesLists(evt)
}

func CacheChatMessage(evt *nostr.Event, privateKey string) {
	// cache full decrypted event under the list for that chat
	var peerPubkey string
	if derivedPublicKey, _ := nostr.GetPublicKey(privateKey); derivedPublicKey != evt.PubKey {
		peerPubkey = evt.Tags.GetFirst([]string{"p", ""}).Value()
	} else {
		peerPubkey = evt.PubKey
	}
	sharedSecret, _ := nip04.ComputeSharedSecret(privateKey, peerPubkey)
	evt.Content, _ = nip04.Decrypt(evt.Content, sharedSecret)

	store.SaveUnencryptedEventOnChatHistory(chatId(evt), evt)
}

func GetRelayRecommendationForPubkey(pubkey string) string {
	relays := store.GetTopRelaysForPubkey(pubkey, 1)
	if len(relays) > 0 {
		return relays[0]
	}
	return ""
}

func fallback(n int) []string {
	if n > len(config.FallbackRelays) {
		n = len(config.FallbackRelays)
	}

	rand.Shuffle(len(config.FallbackRelays), func(i, j int) {
		config.FallbackRelays[j], config.FallbackRelays[i] = config.FallbackRelays[i], config.FallbackRelays[j]
	})

	relays := make([]string, 0, n)
	relays = append(relays, config.FallbackRelays[0:n]...)

	return relays
}
