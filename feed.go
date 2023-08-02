package main

import (
	"encoding/json"
	"net/http"
)

func homeHandler(w http.ResponseWriter, r *http.Request) {
	notes := make([]Note, 0, 20)

	statuses := make([]*Status, len(notes))
	for i, note := range notes {
		statuses[i] = note.toStatus(r.Context())
	}

	json.NewEncoder(w).Encode(statuses)
}
