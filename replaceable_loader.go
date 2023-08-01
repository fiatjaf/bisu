package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/arriqaaq/flashdb"
	"github.com/graph-gophers/dataloader/v7"
	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nson"
	"golang.org/x/exp/slices"
)

var REPLACEABLE_CACHE_TTL = int64((time.Hour * 18).Seconds())

var replaceableLoaders = make(map[int]*dataloader.Loader[string, *nostr.Event])

type EventResult dataloader.Result[*nostr.Event]

func initializeDataloaders() {
	replaceableLoaders[0] = createReplaceableDataloader(0)
	replaceableLoaders[3] = createReplaceableDataloader(3)
	replaceableLoaders[10000] = createReplaceableDataloader(10000)
	replaceableLoaders[10002] = createReplaceableDataloader(10002)
}

func createReplaceableDataloader(kind int) *dataloader.Loader[string, *nostr.Event] {
	return dataloader.NewBatchedLoader(
		func(
			ctx context.Context,
			pubkeys []string,
		) []*dataloader.Result[*nostr.Event] {
			return batchLoadReplaceableEvents(ctx, kind, pubkeys)
		},
		dataloader.WithBatchCapacity[string, *nostr.Event](400),
		dataloader.WithClearCacheOnBatch[string, *nostr.Event](),
		dataloader.WithWait[string, *nostr.Event](time.Millisecond*400),
	)
}

func batchLoadReplaceableEvents(
	ctx context.Context,
	kind int,
	pubkeys []string,
) []*dataloader.Result[*nostr.Event] {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	batchSize := len(pubkeys)
	results := make([]*dataloader.Result[*nostr.Event], batchSize)
	keyPositions := make(map[string]int)           // { [pubkey]: slice_index }
	relayFilters := make(map[string]*nostr.Filter) // { [relayUrl]: filter }

	for i, pubkey := range pubkeys {
		evt := loadReplaceableEventFromCache(ctx, pubkey, kind)
		if evt != nil {
			results[i] = &dataloader.Result[*nostr.Event]{Data: evt}
			continue
		}

		// save attempts here so we don't try the same failed query over and over
		if doneRecently(shouldFetchReplKey(kind, pubkey), time.Hour*1) {
			results[i] = &dataloader.Result[*nostr.Event]{
				Error: fmt.Errorf("last attempt failed, waiting more to try again"),
			}
			continue
		}

		// build batched queries for the external relays
		keyPositions[pubkey] = i // this is to help us know where to save the result later

		// gather relays we'll use for this pubkey
		relays := determineRelaysToQuery(ctx, pubkey, kind)

		// by default we will return an error (this will be overwritten when we find an event)
		results[i] = &dataloader.Result[*nostr.Event]{
			Error: fmt.Errorf("couldn't find a kind %d event anywhere %v", kind, relays),
		}

		for _, relay := range relays {
			// each relay will have a custom filter
			filter, ok := relayFilters[relay]
			if !ok {
				filter = &nostr.Filter{
					Kinds:   []int{kind},
					Authors: make([]string, 0, batchSize-i /* this and all pubkeys after this can be added */),
				}
				relayFilters[relay] = filter
			}
			filter.Authors = append(filter.Authors, pubkey)
		}
	}

	// process and cache results as we get them
	newEvents := make(chan *nostr.Event)
	defer close(newEvents)
	go func() {
		for evt := range newEvents {
			b, _ := nson.Marshal(evt)
			flash.Update(func(txn *flashdb.Tx) error {
				return txn.SetEx(replaceableKey(evt.Kind, evt.PubKey), b, REPLACEABLE_CACHE_TTL)
			})
		}
	}()

	// query all relays with the prepared filters
	multiSubs := batchReplaceableRelayQueries(ctx, relayFilters)
	for {
		select {
		case evt, more := <-multiSubs:
			if !more {
				return results
			}

			// insert this event at the desired position
			pos := keyPositions[evt.PubKey] // @unchecked: it must succeed because it must be a key we passed
			if results[pos].Data == nil || results[pos].Data.CreatedAt < evt.CreatedAt {
				results[pos] = &dataloader.Result[*nostr.Event]{Data: evt}
				newEvents <- evt
			}
		case <-ctx.Done():
			return results
		}
	}
}

func determineRelaysToQuery(ctx context.Context, pubkey string, kind int) []string {
	// search in specific relays for user
	relays := fetchOutboxRelaysForUser(ctx, pubkey, 1, false)

	// use a different set of extra relays depending on the kind
	switch kind {
	case 0:
		relays = append(relays, profileRelays...)
	case 3:
		relays = append(relays, contactListRelays...)
	case 10002:
		relays = append(relays, relayListRelays...)
	}

	for len(relays) < 3 {
		next++
		relays = append(relays, defaultRelays[next%len(defaultRelays)])
	}

	// save attempts
	if kind != 10002 {
		go func() {
			for _, relay := range relays {
				if !slices.Contains(profileRelays, relay) && !slices.Contains(relayListRelays, relay) {
					saveLastAttempted(ctx, pubkey, relay)
				}
			}
		}()
	}

	return relays
}

// batchReplaceableRelayQueries subscribes to multiple relays using a different filter for each and returns
// a single channel with all results. it closes on EOSE or when all the expected events were returned.
//
// the number of expected events is given by the number of pubkeys in the .Authors filter field.
// because of that, batchReplaceableRelayQueries is only suitable for querying replaceable events -- and
// care must be taken to not include the same pubkey more than once in the filter .Authors array.
func batchReplaceableRelayQueries(
	ctx context.Context,
	relayFilters map[string]*nostr.Filter,
) <-chan *nostr.Event {
	bg := ctx
	all := make(chan *nostr.Event)

	wg := sync.WaitGroup{}
	wg.Add(len(relayFilters))
	for url, filter := range relayFilters {
		go func(url string, filter *nostr.Filter) {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(ctx, time.Second*4)
			defer cancel()

			n := len(filter.Authors)
			rl, _ := pool.EnsureRelay(url)
			if rl == nil {
				return
			}
			sub, _ := rl.Subscribe(ctx, nostr.Filters{*filter}, nostr.WithLabel("dl-repl"))
			if sub == nil {
				return
			}

			received := 0
			for {
				select {
				case evt, more := <-sub.Events:
					if !more {
						// ctx canceled, sub.Events is closed
						return
					}

					all <- evt

					if evt.Kind != 10002 &&
						!slices.Contains(relayListRelays, sub.Relay.URL) &&
						!slices.Contains(profileRelays, sub.Relay.URL) {
						// associate relays
						go func() {
							saveLastFetched(bg, evt, sub.Relay.URL)
							grabRelaysFromEvent(bg, evt)
						}()
					}

					received++
					if received >= n {
						// we got all events we asked for, unless the relay is shitty and sent us two from the same
						return
					}
				case <-sub.EndOfStoredEvents:
					// close here
					return
				}
			}
		}(url, filter)
	}

	go func() {
		wg.Wait()
		close(all)
	}()

	return all
}

func loadReplaceableEventFromCache(ctx context.Context, pubkey string, kind int) *nostr.Event {
	k := replaceableKey(kind, pubkey)
	evt := &nostr.Event{}
	flash.View(func(txn *flashdb.Tx) error {
		if val, _ := txn.Get(k); val != "" {
			err := nson.Unmarshal(val, evt)
			if err != nil {
				evt = nil
			}
		} else {
			evt = nil
		}
		return nil
	})
	return evt
}
