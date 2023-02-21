package main

import (
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

var (
	store   *Store
	pool    *relayPool
	config  *Config
	program *tea.Program
)

func main() {
	f, err := tea.LogToFile("debug.log", "debug")
	if err != nil {
		fmt.Println("fatal:", err)
		os.Exit(1)
	}
	defer f.Close()

	config, err = handleConfig()
	if err != nil {
		log.Fatal("failed to parse config: " + err.Error())
	}

	store = initStore(config.DataDir)
	for _, f := range config.Following {
		store.FollowKey(config.PublicKey, f.Key)
		for _, url := range f.Relays {
			store.IncrementRelayScoreForPubkey(f.Key, url, 1)
		}
	}

	pool = newRelayPool()

	program = tea.NewProgram(initialModel())

	if _, err := program.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
