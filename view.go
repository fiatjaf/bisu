package main

import "github.com/charmbracelet/lipgloss"

const (
	listWidth = 15

	activeBorderColor   = lipgloss.Color("255")
	inactiveBorderColor = lipgloss.Color("238")
)

func (m Model) View() string {
	listBorderColor := activeBorderColor
	if m.active != list {
		listBorderColor = inactiveBorderColor
	}
	screenBorderColor := activeBorderColor
	if m.active != screen {
		screenBorderColor = inactiveBorderColor
	}

	m.list.UpdateViewport()
	m.list.SetHeight((m.height * 80 / 100) - 7)
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
				Render(m.list.View()),
			lipgloss.NewStyle().
				Height(screenHeight).
				Width(screenWidth).
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(screenBorderColor).
				Render("loading "+m.screenSubject),
		),
		lipgloss.JoinHorizontal(lipgloss.Center,
			lipgloss.NewStyle().
				MarginLeft(2).
				Render(m.input.View()),
		),
	)
}
