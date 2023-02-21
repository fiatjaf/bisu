package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	syncmap "github.com/SaveTheRbtz/generic-sync-map-go"
	"github.com/charmbracelet/bubbles/list"
	"github.com/nbd-wtf/go-nostr"
)

// takes a channel that may emit repeated events and returns one that only emits unique events.
// however the uniqueness cache for each event is cleaned after 30 seconds, so it might emit a
// repeated event if it comes 30 seconds after the first.
func softUniquenessFilter(events chan *nostr.Event) chan *nostr.Event {
	unique := make(chan *nostr.Event)
	eventsSeen := syncmap.MapOf[string, struct{}]{}

	go func() {
		for evt := range events {
			if _, seen := eventsSeen.Load(evt.ID); seen {
				continue
			} else {
				unique <- evt
				eventsSeen.Store(evt.ID, struct{}{})
				go func(id string) {
					time.Sleep(30 * time.Second)
					eventsSeen.Delete(id)
				}(evt.ID)
			}
		}
	}()

	return unique
}

// run the same function multiple times, each with one of the provided args, and returns
// the first result from any of those.
func race[A any, R any, T func(A) (*R, error)](task T, args []A) (*R, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("no arguments to race on")
	}

	first := sync.Once{}
	result := make(chan *R)
	tries := sync.WaitGroup{}

	tries.Add(len(args))

	for i, arg := range args {
		go func(idx int, arg A) {
			r, err := task(arg)
			if err == nil {
				first.Do(func() {
					result <- r
				})
			}
			tries.Done()
		}(i, arg)
	}

	failed := make(chan struct{})
	go func() {
		tries.Wait()
		failed <- struct{}{}
	}()

	select {
	case r := <-result:
		return r, nil
	case <-failed:
		return nil, fmt.Errorf("all racing calls failed")
	}
}

func sortEventsDesc(events []*nostr.Event) func(int, int) bool {
	return func(i, j int) bool {
		return events[i].CreatedAt.After(events[j].CreatedAt)
	}
}

// the chat id is just a convenience. it is given by the first 7 characters of the hash of
// the two public keys involved in a chat, sorted deterministically
func chatId(evt *nostr.Event) string {
	pks1 := evt.PubKey
	pks2 := evt.Tags.GetFirst([]string{"p", ""}).Value()
	return chatIdFromPubkeys(pks1, pks2)
}

func chatIdFromPubkeys(pks1 string, pks2 string) string {
	pubkey1, _ := hex.DecodeString(pks1)
	pubkey2, _ := hex.DecodeString(pks2)

	var payload []byte
	if pks1 > pks2 {
		payload = append(pubkey2, pubkey1...)
	} else {
		payload = append(pubkey1, pubkey2...)
	}

	hash := sha256.Sum256(payload)
	return hex.EncodeToString(hash[:])[0:7]
}

func takeFirst[S any](seq []S, n int) []S {
	if len(seq) < n {
		return seq
	}
	return seq[0:n]
}

func takeFirstString(seq string, n int) string {
	if len(seq) < n {
		return seq
	}
	return seq[0:n]
}

func insertItemIntoDescendingList(sortedArray []list.Item, event item) []list.Item {
	start := 0
	end := len(sortedArray) - 1
	var mid int
	position := start

	if end < 0 {
		return []list.Item{event}
	} else if event.CreatedAt.Before(sortedArray[end].(item).CreatedAt) {
		return append(sortedArray, event)
	} else if event.CreatedAt.After(sortedArray[start].(item).CreatedAt) {
		newArr := make([]list.Item, len(sortedArray)+1)
		newArr[0] = event
		copy(newArr[1:], sortedArray)
		return newArr
	} else if event.CreatedAt.Equal(sortedArray[start].(item).CreatedAt) {
		position = start
	} else {
		for {
			if end <= start+1 {
				position = end
				break
			}
			mid = int(start + (end-start)/2)
			if sortedArray[mid].(item).CreatedAt.After(event.CreatedAt) {
				start = mid
			} else if sortedArray[mid].(item).CreatedAt.Before(event.CreatedAt) {
				end = mid
			} else {
				position = mid
				break
			}
		}
	}

	if sortedArray[position].(item).ID != event.ID {
		newArr := make([]list.Item, len(sortedArray)+1)
		copy(newArr[:position], sortedArray[:position])
		newArr[position] = event
		copy(newArr[position+1:], sortedArray[position:])
		return newArr
	}

	return sortedArray
}
