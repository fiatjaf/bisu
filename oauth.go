package main

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/nbd-wtf/go-nostr"
)

type token struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
	CreatedAt   int64  `json:"created_at"`
}

func oauthHandler(w http.ResponseWriter, r *http.Request) {
	var redir string
	if r.Method == "POST" {
		r.ParseForm()
		redir = r.FormValue("redirect_uri")
	} else if r.Method == "GET" {
		redir = r.URL.Query().Get("redirect_uri")
	}

	re, err := url.Parse(redir)
	if err != nil {
		http.Error(w, "invalid redirect_uri", 400)
		return
	}

	qs := re.Query()

	// TODO maybe we need something else here -- for now it's just to trigger an action from the trontend
	qs.Add("code", "bla")

	re.RawQuery = qs.Encode()
	http.Redirect(w, r, re.String(), http.StatusFound)
}

func createTokenHandler(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(token{
		AccessToken: "_",
		TokenType:   "Bearer",
		Scope:       "read write follow push",
		CreatedAt:   int64(nostr.Now()),
	})
}

func appCredentialsHandler(w http.ResponseWriter, r *http.Request) {
}
