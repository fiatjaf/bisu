package main

import (
	"log"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case error:
		log.Print(msg.Error())
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

				switch msg.String() {
				case "enter":
					// enter text or command
					text := strings.TrimSpace(m.input.Value())
					if len(text) > 0 {
						if text[0] == '/' {
							cmd, err := handleCommand(text)
							if err == nil {
								cmds = append(cmds, cmd)
								m.input.SetValue("")
							}
						} else {
							cmds = append(cmds, publishNote(text))
							m.input.SetValue("")
						}
					}
				}
			case list:
				newModel, cmd := m.list.Update(msg)
				m.list = newModel
				cmds = append(cmds, cmd)

				switch msg.String() {
				case "enter":
					// select from list of channels
					m.list.Blur()
					m.active = screen
					m.screenSubject = newModel.SelectedRow()[0]
				}
			}
		}
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width
	}

	// always update the main screen
	newModel, cmd := m.page.Update(msg)
	m.page = newModel
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}
