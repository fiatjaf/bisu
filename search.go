package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/nbd-wtf/go-nostr"
)

type searchResponse struct {
	Accounts []*Account `json:"accounts"`
	Statuses []any      `json:"statuses"`
	Hashtags []any      `json:"hashtags"`
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	filter := nostr.Filter{Search: r.URL.Query().Get("q")}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 20
	}
	filter.Limit = limit
	switch r.URL.Query().Get("type") {
	case "accounts":
		filter.Kinds = []int{0}
	}
	events := pool.SubManyEose(r.Context(), searchRelays, nostr.Filters{filter})
	accounts := make([]*Account, 0, limit)
	i := 0
	for evt := range events {
		accounts = append(accounts, toAccount(r.Context(), toProfile(evt.Event), nil))
		i++
		if i >= limit {
			break
		}
	}

	json.NewEncoder(w).Encode(searchResponse{
		Accounts: accounts,
		Statuses: make([]any, 0),
		Hashtags: make([]any, 0),
	})
}
