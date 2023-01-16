package main

import (
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/nbd-wtf/go-nostr"
)

type Model struct {
	channels []nostr.Event
	selected int

	body  viewport.Model
	input textinput.Model
}

func initialModel() Model {
	body := viewport.New(34, 20)
	input := textinput.New()
	input.Width = 34
	input.Focus()

	return Model{
		channels: []nostr.Event{},
		selected: 0,

		body:  body,
		input: input,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.EnterAltScreen
}
