package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/nbd-wtf/go-nostr"
)

func homeHandler(w http.ResponseWriter, r *http.Request) {
	qs := r.URL.Query()
	limit, _ := strconv.Atoi(qs.Get("limit"))
	if limit == 0 {
		limit = 20
	}
	maxId := qs.Get("max_id")

	var keys []string
	if follows := loadContactList(r.Context(), profile.pubkey); follows != nil {
		keys = make([]string, len(*follows))
		for i, f := range *follows {
			keys[i] = f.Pubkey
		}
	}
	keys = append(keys, profile.pubkey)

	var until *nostr.Timestamp
	if maxId != "" {
		max := loadEvent(r.Context(), maxId, nil, nil)
		until = &max.CreatedAt
	}

	events, err := store.QueryEvents(r.Context(), nostr.Filter{
		Kinds:   []int{1},
		Authors: keys,
		Limit:   limit,
		Until:   until,
	})
	if err != nil {
		jsonError(w, "error querying internal db: "+err.Error(), 500)
		return
	}

	statuses := make([]*Status, 0, limit)
	for evt := range events {
		statuses = append(statuses, toStatus(r.Context(), evt))
	}

	json.NewEncoder(w).Encode(statuses)
}
