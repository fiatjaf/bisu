package main

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nbd-wtf/go-nostr"
)

type ActiveView uint

const (
	Input = iota
	Sidebar
	Screen
)

type Model struct {
	width  int
	height int

	active ActiveView
	page   Page

	sidebar *SidebarModel
	input   textarea.Model

	homefeed *feedPage

	screenSubject string
}

type Page interface {
	Init() tea.Cmd
	Update(tea.Msg) (Page, tea.Cmd)
	View(x, y int) string
	Focus()
	Blur()
}

func initialModel() *Model {
	return &Model{
		active: Input,
		page:   newLoadingPage(),

		sidebar: newSidebar(),
		input:   newTextarea(),

		homefeed: newFeedPage("home", []*nostr.Event{}),
	}
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		tea.EnterAltScreen,
		initialFetch(m),
		m.page.Init(),
	)
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
