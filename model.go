package main

import (
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type activeView uint

const (
	input activeView = iota
	list
	screen
)

type Model struct {
	width  int
	height int

	active activeView

	list  table.Model
	input textarea.Model

	screenSubject string
}

func initialModel() Model {
	m := Model{
		active: input,

		list:  newList(),
		input: newTextarea(),
	}

	return m
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		tea.EnterAltScreen,
		initialFetch,
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
	return t
}