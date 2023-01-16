package main

import (
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fiatjaf/norad2"
)

var (
	nr     *norad2.Core
	store  *Store
	config *Config
)

func main() {
	var err error
	config, err = handleConfig()
	if err != nil {
		log.Fatal("failed to parse config: " + err.Error())
	}

	store = InitStore(config.DataDir)
	for _, f := range config.Following {
		store.FollowKey(config.PublicKey, f.Key)
		for _, url := range f.Relays {
			store.IncrementRelayScoreForPubkey(f.Key, url, 1)
		}
	}

	nr = norad2.New(store, norad2.Options{
		FallbackRelays: config.FallbackRelays,
		AlwaysCheck:    []string{},
		SafeRelays:     config.SafeRelays,
	})

	m := initialModel()
	log.Print("x")
	m.homefeed = nr.GetCachedHomeFeedEvents(config.PublicKey, 99999999999999, 100)
	log.Print("y")
	p := tea.NewProgram(m)

	f, err := tea.LogToFile("debug.log", "debug")
	if err != nil {
		fmt.Println("fatal:", err)
		os.Exit(1)
	}
	defer f.Close()

	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
