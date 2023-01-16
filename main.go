package main

import (
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fiatjaf/norad2"
)

var nr *norad2.Core

func main() {
	p := tea.NewProgram(initialModel())

	config, err := handleConfig()
	if err != nil {
		log.Fatal("failed to parse config: " + err.Error())
	}

	store := InitStore(config.DataDir)
	nr = norad2.New(store, norad2.Options{
		FallbackRelays: []string{"wss://nostr-pub.wellorder.net"},
		AlwaysCheck:    []string{},
		SafeRelays:     []string{},
	})

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
