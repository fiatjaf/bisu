package main

import tea "github.com/charmbracelet/bubbletea"

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		case "up", "k":
			if m.selected > 0 {
				m.selected--
			}
		case "down", "j":
			if m.selected < len(m.channels)-1 {
				m.selected++
			}
		}
	}

	if updated, c := m.input.Update(msg); true {
		m.input = updated
		cmd = tea.Batch(cmd, c)
	}

	if updated, c := m.body.Update(msg); true {
		m.body = updated
		cmd = tea.Batch(cmd, c)
	}

	return m, cmd
}
