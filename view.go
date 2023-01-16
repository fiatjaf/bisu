package main

import "github.com/charmbracelet/lipgloss"

func (m Model) View() string {
	var channels string
	for _, channel := range m.channels {
		channels += channel.ID
	}

	return lipgloss.JoinVertical(0,
		"== channels ==",
		lipgloss.JoinHorizontal(10,
			channels,
			m.body.View(),
		),
		m.input.View(),
	)
}
