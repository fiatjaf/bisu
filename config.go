package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
)

type Config struct {
	DataDir         string            `json:"-"`
	PrivateKey      string            `json:"privatekey,omitempty"`
	Following       []Follow          `json:"following"`
	Relays          map[string]Policy `json:"relays,omitempty"`
	FallbackRelays  []string          `json:"fallback_relays,omitempty"`
	SafeRelays      []string          `json:"safe_relays,omitempty"`
	WriteableRelays []string          `json:"writeable_relays,omitempty"`
}

type Follow struct {
	Key    string   `json:"key"`
	Name   string   `json:"name,flow,omitempty"`
	Relays []string `json:"relays,flow,omitempty"`
}

type Policy struct {
	Read  bool `json:"read"`
	Write bool `json:"write"`
}

func handleConfig() (*Config, error) {
	var config Config

	flag.StringVar(&config.DataDir, "datadir", "~/.config/nostr",
		"Base directory for configurations and data from Nostr.")
	flag.Parse()
	config.DataDir, _ = homedir.Expand(config.DataDir)
	os.Mkdir(config.DataDir, 0700)

	path := filepath.Join(config.DataDir, "config.json")
	f, err := os.Open(path)
	if err != nil {
		saveConfig(path, config)
		f, _ = os.Open(path)
	}
	f, _ = os.Open(path)
	err = json.NewDecoder(f).Decode(&config)
	if err != nil {
		return nil, fmt.Errorf("invalid json (%s): %w", path, err)
	}

	if len(config.FallbackRelays) == 0 {
		for relay, policy := range config.Relays {
			if policy.Read {
				config.FallbackRelays = append(config.FallbackRelays, relay)
			}
		}
	}

	if len(config.FallbackRelays) == 0 {
		config.FallbackRelays = []string{
			"wss://nostr-pub.wellorder.net",
			"wss://nostr-relay.wlvs.space",
			"wss://relay.damus.io",
			"wss://relay.nostr.info",
			"wss://nostr-2.zebedee.cloud",
			"wss://nostr.fmt.wiz.biz",
			"wss://nostr.v0l.io",
			"wss://nostr-relay.untethr.me",
		}
	}

	if len(config.WriteableRelays) == 0 {
		for relay, policy := range config.Relays {
			if policy.Write {
				config.WriteableRelays = append(config.WriteableRelays, relay)
			}
		}
	}

	if len(config.WriteableRelays) == 0 {
		config.WriteableRelays = []string{
			"wss://nostr-pub.wellorder.net",
			"wss://nostr-relay.wlvs.space",
			"wss://relay.damus.io",
			"wss://relay.nostr.info",
			"wss://nostr-2.zebedee.cloud",
			"wss://nostr.fmt.wiz.biz",
			"wss://nostr.v0l.io",
			"wss://nostr-relay.untethr.me",
		}
	}

	if len(config.SafeRelays) == 0 {
		config.SafeRelays = []string{
			"wss://relay.minds.com/nostr/v1/ws",
			"wss://nostr-verified.wellorder.net",
			"wss://relay.damus.io",
		}
	}

	return &config, nil
}

func saveConfig(path string, config Config) {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal("can't open config file " + path + ": " + err.Error())
		return
	}

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")

	enc.Encode(config)
}
