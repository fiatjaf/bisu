package main

import (
	"log"
	"strings"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nbd-wtf/go-nostr"
)

var mutex sync.Mutex

type SingletonCommand int

const (
	UpdateFollows SingletonCommand = iota
)

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case error:
		log.Printf("error msg: %s", msg.Error())
	case *nostr.Event:
		mutex.Lock()
		defer mutex.Unlock()
		newModel, cmd := m.homefeed.Update(msg)
		m.homefeed = newModel.(*feedPage)
		cmds = append(cmds, cmd)
	case Page:
		m.page = msg
	case SingletonCommand:
		switch msg {
		case UpdateFollows:
			m.sidebar.Update(msg)
		}
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "tab":
			switch m.active {
			case Input:
				m.active = Sidebar
				m.input.Blur()
				m.sidebar.table.Focus()
			case Sidebar:
				m.active = Screen
				m.sidebar.table.Blur()
				m.page.Focus()
			case Screen:
				m.active = Input
				m.page.Blur()
				m.input.Focus()
			}
		case "esc":
			m.active = Input
			m.input.Focus()
		default:
			switch m.active {
			case Input:
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
			case Sidebar:
				newModel, cmd := m.sidebar.table.Update(msg)
				m.sidebar.table = newModel
				cmds = append(cmds, cmd)

				switch msg.String() {
				case "enter":
					// select from list of channels
					m.sidebar.table.Blur()
					m.active = Screen
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
