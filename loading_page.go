package main

import (
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type loadingPage struct {
	sp spinner.Model
}

func newLoadingPage() *loadingPage {
	sp := spinner.New()
	sp.Style = lipgloss.NewStyle().
		Foreground(lipgloss.Color("69")).
		AlignHorizontal(lipgloss.Center).
		AlignVertical(lipgloss.Center)
	sp.Spinner = spinner.Points
	return &loadingPage{sp}
}

func (lp loadingPage) View(x, y int) string {
	lp.sp.Style = lp.sp.Style.
		Width(x).
		Height(y)
	return lp.sp.View()
}

func (lp loadingPage) Init() tea.Cmd {
	return lp.sp.Tick
}

func (lp *loadingPage) Update(msg tea.Msg) (page, tea.Cmd) {
	sp, cmd := lp.sp.Update(msg)
	lp.sp = sp
	return lp, cmd
}
