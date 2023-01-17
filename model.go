package main

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nbd-wtf/go-nostr"
)

type activeView uint

const (
	input activeView = iota
	sidebar
	screen
)

type Model struct {
	width  int
	height int

	active activeView
	page   page

	sidebar table.Model
	input   textarea.Model

	homefeed *feedPage

	screenSubject string
}

type page interface {
	Init() tea.Cmd
	Update(tea.Msg) (page, tea.Cmd)
	View(x, y int) string
	Focus()
	Blur()
}

func initialModel() Model {
	m := Model{
		active: input,
		page:   newLoadingPage(),

		sidebar: newList(),
		input:   newTextarea(),

		homefeed: newFeedPage("Home Feed", []*nostr.Event{}),
	}

	return m
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		tea.EnterAltScreen,
		initialFetch(m),
		m.page.Init(),
	)
}

func newList() table.Model {
	t := table.New(
		table.WithColumns([]table.Column{{Title: "channels", Width: listWidth}}),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	t.SetRows([]table.Row{})

	return t
}

func newTextarea() textarea.Model {
	t := textarea.New()
	t.Prompt = ""
	t.Placeholder = "Good afternoon."
	t.Focus()

	t.CharLimit = 680

	t.BlurredStyle.Base = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(inactiveBorderColor)
	t.FocusedStyle.Base = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(activeBorderColor)

	t.ShowLineNumbers = false

	t.KeyMap.InsertNewline = key.NewBinding(key.WithKeys("shift+enter", "ctrl+m"))

	return t
}
