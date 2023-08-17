package main

import (
	"time"

	ristretto "github.com/fiatjaf/generic-ristretto"
	"github.com/nbd-wtf/go-nostr"
)

type hex32Cache[V any] struct {
	Cache *ristretto.Cache[string, V]
}

func newHex32Cache[V any](max int64) *hex32Cache[V] {
	cache, _ := ristretto.NewCache(&ristretto.Config[string, V]{
		NumCounters: max * 10,
		MaxCost:     max,
		BufferItems: 64,
		KeyToHash:   func(key string) (uint64, uint64) { return h32(key), 0 },
	})
	return &hex32Cache[V]{Cache: cache}
}

func (s hex32Cache[V]) Get(k string) (v V, ok bool)     { return s.Cache.Get(k) }
func (s hex32Cache[V]) Delete(k string)                 { s.Cache.Del(k) }
func (s hex32Cache[V]) Set(k string, v V, c int64) bool { return s.Cache.Set(k, v, c) }
func (s hex32Cache[V]) SetWithTTL(k string, v V, c int64, d time.Duration) bool {
	return s.Cache.SetWithTTL(k, v, c, d)
}

func h32(key string) uint64 { return shortUint64(key) }

var (
	eventCache        = newHex32Cache[*nostr.Event](36_000)
	metadataCache     = newHex32Cache[*Profile](8_000)
	contactListsCache = newHex32Cache[*[]Follow](8_000)
)
