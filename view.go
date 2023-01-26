package main

import (
	"github.com/charmbracelet/lipgloss"
)

const (
	listWidth = 15

	activeBorderColor   = lipgloss.Color("255")
	inactiveBorderColor = lipgloss.Color("238")
)

func (m *Model) View() string {
	listBorderColor := activeBorderColor
	if m.active != Sidebar {
		listBorderColor = inactiveBorderColor
	}
	screenBorderColor := activeBorderColor
	if m.active != Screen {
		screenBorderColor = inactiveBorderColor
	}

	m.sidebar.table.UpdateViewport()
	m.sidebar.table.SetHeight((m.height * 80 / 100) - 7)
	screenHeight := (m.height * 80 / 100) - 5
	screenWidth := (m.width * 100 / 100) - 10 - listWidth
	m.input.SetHeight(m.height * 20 / 100)
	m.input.SetWidth((m.width * 100 / 100) - 4)

	return lipgloss.JoinVertical(lipgloss.Center,
		lipgloss.JoinHorizontal(lipgloss.Left,
			lipgloss.NewStyle().
				MarginLeft(2).
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(listBorderColor).
				Render(m.sidebar.table.View()),
			lipgloss.NewStyle().
				Height(screenHeight).
				Width(screenWidth).
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(screenBorderColor).
				Render(m.page.View(screenWidth, screenHeight)),
		),
		lipgloss.JoinHorizontal(lipgloss.Center,
			lipgloss.NewStyle().
				MarginLeft(2).
				Render(m.input.View()),
		),
	)
}
