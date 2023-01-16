package main

import (
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "tab":
			switch m.active {
			case input:
				m.active = list
				m.input.Blur()
				m.list.Focus()
			case list:
				m.active = input
				m.list.Blur()
				m.input.Focus()
			case screen:
				m.active = list
				m.list.Focus()
			}
		case "esc":
			m.active = input
			m.input.Focus()
		default:
			switch m.active {
			case input:
				newModel, cmd := m.input.Update(msg)
				m.input = newModel
				cmds = append(cmds, cmd)
			case list:
				newModel, cmd := m.list.Update(msg)

				switch msg.String() {
				case "enter":
					m.list.Blur()
					m.active = screen
					m.screenSubject = newModel.SelectedRow()[0]
				}

				m.list = newModel
				cmds = append(cmds, cmd)
			}
		}
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width
	}

	return m, tea.Batch(cmds...)
}
