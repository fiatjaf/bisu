package main

import (
	"encoding/json"
	"net/http"

	"github.com/nbd-wtf/go-nostr"
)

func verifyCredentialsHandler(w http.ResponseWriter, r *http.Request) {
	acct := toAccount(r.Context(), profile, &ToAccountOpts{WithSource: true})
	json.NewEncoder(w).Encode(acct)
}

func updateCredentialsHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(1 << 24)
	if err != nil {
		jsonError(w, "invalid multipart form", 400)
		return
	}

	profile.Name = r.FormValue("display_name")
	profile.About = r.FormValue("note")
	profile.Website = r.FormValue("website")

	j, _ := json.Marshal(profile)

	evt, err := publish(r.Context(), &nostr.Event{
		CreatedAt: nostr.Now(),
		Kind:      0,
		Content:   string(j),
	})
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}

	profile.event = evt
	json.NewEncoder(w).Encode(toAccount(r.Context(), profile, nil))
}
