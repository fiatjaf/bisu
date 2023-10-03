package main

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/tidwall/gjson"
)

type urls struct {
	StreamingAPI string `json:"streaming_api"`
}

type instance struct {
	URI              string `json:"uri"`
	Title            string `json:"title"`
	Description      string `json:"description"`
	ShortDescription string `json:"short_description"`
	Registrations    bool   `json:"registrations"`
	MaxTootChars     int    `json:"max_toot_chars"`
	Urls             urls   `json:"urls"`
	Version          string `json:"version"`
	Rules            []any  `json:"rules"`
}

type app struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Website      string `json:"website"`
	RedirectURI  any    `json:"redirect_uri"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	VapidKey     string `json:"vapid_key"`
}

func instanceHandler(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(instance{
		URI:              srv.Addr,
		Title:            "bisu",
		Description:      "nostr personal homeserver",
		ShortDescription: "nostr homeserver",
		Registrations:    false,
		MaxTootChars:     90000,
		Version:          "2.7.2 (compatible; bisu 0.0.0)",
		Rules:            []any{},
		Urls: urls{
			StreamingAPI: "ws://" + srv.Addr,
		},
	})
}

func createAppHandler(w http.ResponseWriter, r *http.Request) {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		jsonError(w, "wrong body: "+err.Error(), 400)
		return
	}
	defer r.Body.Close()

	json.NewEncoder(w).Encode(app{
		ID:           "2",
		Name:         "bisu",
		Website:      "https://_/",
		RedirectURI:  gjson.GetBytes(b, "redirect_uris").Value(),
		ClientID:     "BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB",
		ClientSecret: "BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB",
		VapidKey:     "BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB=",
	})
}

func appCredentialsHandler(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(app{
		ID:           "2",
		Name:         "bisu",
		Website:      "https://_/",
		RedirectURI:  "urn:ietf:wg:oauth:2.0:oob",
		ClientID:     "BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB",
		ClientSecret: "BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB",
		VapidKey:     "BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB=",
	})
}
