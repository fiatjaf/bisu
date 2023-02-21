package main

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	syncmap "github.com/SaveTheRbtz/generic-sync-map-go"
	"github.com/nbd-wtf/go-nostr"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

type relayPool struct {
	relays syncmap.MapOf[string, *nostr.Relay]
	homeFeedSubscriptions
}

type homeFeedSubscriptions struct {
	sync.RWMutex // use a global mutex for everything instead of a bunch of syncmaps and syncslices

	// which channels should each pubkey event be emitted to
	pubkey2channels map[string][]chan *nostr.Event

	// which relays and subscriptions we're using to listen for each pubkey
	pubkey2relays map[string]map[*nostr.Relay]*nostr.Subscription

	// global map of all channels and homefeed subscriptions that are open
	relay2subscription map[*nostr.Relay]*nostr.Subscription

	// filters to use on the next subscription trigger
	// (i.e. when someone wants to add a key to a subscription we will use the opportunity to remove
	//  old keys that we don't want to track anymore)
	// if this is not defined for a given subscription just reuse the current filters
	filtersPending map[*nostr.Subscription]nostr.Filters
}

func newRelayPool() *relayPool {
	return &relayPool{
		homeFeedSubscriptions: homeFeedSubscriptions{
			pubkey2channels:    make(map[string][]chan *nostr.Event),
			pubkey2relays:      make(map[string]map[*nostr.Relay]*nostr.Subscription),
			relay2subscription: make(map[*nostr.Relay]*nostr.Subscription),
			filtersPending:     make(map[*nostr.Subscription]nostr.Filters),
		},
	}
}

func (pool *relayPool) ensureRelay(url string) *nostr.Relay {
	url = nostr.NormalizeURL(url)

	relay, ok := pool.relays.Load(url)
	if !ok {
		var err error
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		relay, err = nostr.RelayConnect(ctx, url)
		if err != nil {
			return nil
		}
		pool.relays.Store(url, relay)
	}
	return relay
}

func (pool *relayPool) publish(ctx context.Context, url string, event nostr.Event) error {
	relay := pool.ensureRelay(url)
	if relay == nil {
		return fmt.Errorf("failed to ensure relay: %s", url)
	}

	status := relay.Publish(ctx, event)
	if status == nostr.PublishStatusFailed {
		return fmt.Errorf("publish failed: %d", status)
	}
	return nil
}

func (pool *relayPool) sub(ctx context.Context, url string, filters nostr.Filters) *nostr.Subscription {
	relay := pool.ensureRelay(url)
	if relay == nil {
		return nil
	}

	return relay.Subscribe(ctx, filters)
}

// opens a subscription with the given filters to multiple relays
// it also performs automatic caching of any events found, including updating the relay's score per pubkey
func (pool *relayPool) subMany(ctx context.Context, urls []string, filters nostr.Filters) chan *nostr.Event {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel() // ensure ctx.Done() happens

	uniqueEvents := make(chan *nostr.Event)
	seenAlready := syncmap.MapOf[string, struct{}]{}
	subs := make([]*nostr.Subscription, len(urls))
	for i, url := range urls {
		url = nostr.NormalizeURL(url)
		sub := pool.sub(ctx, url, filters)

		if sub == nil {
			continue
		}

		subs[i] = sub
		go func(url string) {
			for evt := range sub.Events {
				// increase the score of this relay for this pubkey
				store.IncrementRelayScoreForPubkey(evt.PubKey, url, seenScore(evt))

				// dispatch unique events to client
				if _, ok := seenAlready.LoadOrStore(evt.ID, struct{}{}); !ok {
					uniqueEvents <- evt
				}

				// store the relay in which we found this in the event object before caching
				evt.SetExtra("relay", url)

				CacheEvent(evt)
			}
		}(url)
	}

	// close subscriptions when context is canceled
	go func() {
		<-ctx.Done()
		for _, sub := range subs {
			sub.Unsub()
		}
	}()

	return uniqueEvents
}

// does a synchronous query to one relay -- it won't open a connection if one already exists
// it also performs automatic updating of relay-per-pubkey scores, but not any other caching
func (pool *relayPool) querySync(ctx context.Context, url string, filters nostr.Filters) []*nostr.Event {
	if _, ok := ctx.Deadline(); !ok {
		// if no timeout is set, force it to 5 seconds
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
	}

	relay := pool.ensureRelay(url)
	if relay == nil {
		return nil
	}

	sub := relay.Subscribe(ctx, filters)
	var events []*nostr.Event

	defer func() {
		// increment this relay's score for the pubkeys we got events from
		for _, evt := range events {
			store.IncrementRelayScoreForPubkey(evt.PubKey, relay.URL, seenScore(evt))
		}
	}()

	for {
		select {
		case evt := <-sub.Events:
			events = append(events, evt)
		case <-sub.EndOfStoredEvents:
			return events
		case <-ctx.Done():
			return events
		}
	}
}

func (pool *relayPool) addHomeFeedListener(
	ctx context.Context,
	pubkeys []string,
) chan *nostr.Event {
	pool.homeFeedSubscriptions.Lock()
	defer pool.homeFeedSubscriptions.Unlock()

	// create a channel to receive unique updates from these pubkeys
	ch := make(chan *nostr.Event)

	// first we wollprepare all the subscriptions with filters, to only unleash them at the end
	preparedRelayActions := make(map[*nostr.Relay]func())

	// now for each pubkey
	for _, pubkey := range pubkeys {
		// add this channel to the list of channels that will be pinged when an event occur
		channels, ok := pool.pubkey2channels[pubkey]
		if !ok {
			channels = make([]chan *nostr.Event, 0, 5)
		}
		pool.pubkey2channels[pubkey] = append(channels, ch)

		// check if there are already subscriptions open for this pubkey
		if _, ok := pool.pubkey2relays[pubkey]; ok {
			// there are already some, so we're good, do nothing
			// TODO check if we should add more subscriptions
		} else {
			// there are none, will open some
			pool.pubkey2relays[pubkey] = make(map[*nostr.Relay]*nostr.Subscription)

			// check what relays we should use
			relays := store.GetTopRelaysForPubkey(pubkey, 3)
			if len(relays) == 0 {
				relays = fallback(2)
			}

			// we'll have one filter for each pubkeys
			// (this has probably the same cost as putting all authors in the same filter and relays will probably be ok with it),
			// therefore we can pregenerate the filter for this pubkey here and use it in all cases that follow
			filter := nostr.Filter{
				Kinds:   []int{1},
				Authors: []string{pubkey},
				Since:   store.GetLatestTimestampForProfile(pubkey),
				Limit:   100, // just for the initial stored-events query
			}

			// check if we already have subscriptions open for these relays
			for _, url := range relays {
				relay := pool.ensureRelay(url)
				if relay == nil {
					continue
				}

				sub, subscriptionExists := pool.relay2subscription[relay]
				if !subscriptionExists {
					// a subscription for this relay doesn't exist
					// create one
					filters := nostr.Filters{filter}
					sub = relay.PrepareSubscription()
					sub.Filters = filters

					preparedRelayActions[relay] = func() {
						sub.Fire(ctx)

						// handle events from this subscription (this will persist even if we change the subscription)
						go pool.handleHomeFeedEvents(sub, pubkey)
					}
				} else {
					// there is already a subscription to this relay
					// grab pending filters if they exist
					// (this makes them effective since we're going live with them now)
					filters, pendingFiltersExist := pool.filtersPending[sub]
					if !pendingFiltersExist {
						// fallback to live filters
						filters = sub.Filters
					} else {
						// we can delete the pending filters now
						delete(pool.filtersPending, sub)
					}

					// append this pubkey's filter
					filters = append(filters, filter)

					sub.Filters = filters

					if _, exists := preparedRelayActions[relay]; !exists {
						preparedRelayActions[relay] = func() {
							sub.Fire(ctx)
						}
					}
				}

				// we are sure this map exists because we just created it
				pool.pubkey2relays[pubkey][relay] = sub

				// set this subscription in the global map -- idempotent
				pool.relay2subscription[relay] = sub
			}
		}
	}

	for _, action := range preparedRelayActions {
		action()
	}

	// fmt.Fprintln(os.Stderr, "=== AFTER ADDING", pubkeys, "===")
	// pool.homeFeedSubscriptions.printSummary()
	// fmt.Fprintln(os.Stderr, "====================")

	// when the context is canceled, schedule these keys to be removed from the subscriptions on the next update
	go func() {
		<-ctx.Done()
		pool.homeFeedSubscriptions.Lock()
		defer pool.homeFeedSubscriptions.Unlock()

		for _, pubkey := range pubkeys {
			// remove the channel for this listener from everywhere
			channels := pool.pubkey2channels[pubkey] // @unchecked, this must succeed
			idx := slices.Index(channels, ch)        // @unchecked, this element must exist
			pool.pubkey2channels[pubkey] = append(channels[0:idx], channels[idx+1:]...)

			// if no one is listening for this pubkey anymore,
			if len(pool.pubkey2channels[pubkey]) == 0 {
				// remove its slice from the map
				delete(pool.pubkey2channels, pubkey)

				// and schedule it to be removed from all filters
				relay2sub := pool.pubkey2relays[pubkey] // @unchecked, this must succeed
				for relay, sub := range relay2sub {
					nextFilters, hasFilterPending := pool.filtersPending[sub]
					if !hasFilterPending {
						nextFilters = sub.Filters
					}

					// remove the filter that has this pubkey (we use one filter for each pubkey)
					idx := slices.IndexFunc(nextFilters, func(f nostr.Filter) bool {
						return f.Authors[0] == pubkey
					})
					if idx != -1 {
						nextFilters = slices.Delete(nextFilters, idx, idx+1)
					}

					// if this sub is empty of authors after we remove, end it now and remove it
					if len(nextFilters) == 0 {
						delete(pool.filtersPending, sub)
						delete(pool.relay2subscription, relay)
						sub.Unsub()
					} else {
						// otherwise the changes will be stored as pending until they are effective
						pool.filtersPending[sub] = nextFilters
					}
				}
			}
		}

		// fmt.Fprintln(os.Stderr, "=== AFTER REMOVING", pubkeys, "===")
		// pool.homeFeedSubscriptions.printSummary()
		// fmt.Fprintln(os.Stderr, "======================")

		close(ch)
	}()

	// return the channel
	return ch
}

func (pool *relayPool) handleHomeFeedEvents(sub *nostr.Subscription, pubkey string) {
	for evt := range sub.Events {
		// every time an event comes from this sub,
		// cache it in all manners possible
		go CacheEvent(evt)
		// fetch all channels that are interested in it
		pool.homeFeedSubscriptions.RLock()
		channels := pool.pubkey2channels[pubkey] // @unchecked, this must succeed
		pool.homeFeedSubscriptions.RUnlock()
		// and dispatch it to them
		for _, target := range channels {
			target <- evt
		}
	}
}

func (hf *homeFeedSubscriptions) printSummary() {
	fmt.Fprintln(os.Stderr, "relays:")
	for relay, sub := range hf.relay2subscription {
		fmt.Fprintf(os.Stderr, "  - %s\n", relay)
		for _, f := range sub.Filters {
			fmt.Fprintf(os.Stderr, "    - %s (%d)\n", f.Authors[0], f.Since.Unix())
			if pending, hasPending := hf.filtersPending[sub]; hasPending {
				fmt.Fprintf(os.Stderr, "      ~ pending ~")
				for _, pf := range pending {
					fmt.Fprintf(os.Stderr, "    - %s (%d)\n", pf.Authors[0], pf.Since.Unix())
				}
			}
		}
	}
	fmt.Fprintln(os.Stderr, "pubkeys:")
	for k, v := range hf.pubkey2relays {
		fmt.Fprintf(os.Stderr, "  - %s: %s\n", k, maps.Keys(v))
		fmt.Fprintf(os.Stderr, "    channels: %d\n", len(hf.pubkey2channels[k]))
	}
}
