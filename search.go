package main

import (
	"encoding/json"
	"net/http"
)

type searchResponse struct {
	accounts []any `json:"accounts"`
	statuses []any `json:"statuses"`
	hashtags []any `json:"hashtags"`
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: search in relays that provide search

	json.NewEncoder(w).Encode(searchResponse{
		accounts: make([]any, 0),
		statuses: make([]any, 0),
		hashtags: make([]any, 0),
	})
}
