package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/nbd-wtf/go-nostr"
)

func homeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.URL)
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 {
		limit = 20
	}

	var keys []string
	if follows := loadContactList(r.Context(), profile.pubkey); follows != nil {
		keys = make([]string, len(*follows))
		for i, f := range *follows {
			keys[i] = f.Pubkey
		}
	}
	keys = append(keys, profile.pubkey)

	events, err := store.QueryEvents(r.Context(), nostr.Filter{
		Authors: keys,
		Limit:   limit,
	})
	if err != nil {
		http.Error(w, "error querying internal db", 500)
		return
	}

	statuses := make([]*Status, 0, limit)
	for evt := range events {
		statuses = append(statuses, toStatus(r.Context(), evt))
	}

	json.NewEncoder(w).Encode(statuses)
}
