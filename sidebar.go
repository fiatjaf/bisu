package main

import (
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type SidebarModel struct {
	table table.Model
}

func newSidebar() *SidebarModel {
	t := table.New(
		table.WithColumns([]table.Column{{Title: "channels", Width: listWidth}}),
	)

	st := table.DefaultStyles()
	st.Header = st.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(styles.dimmed).
		Background(styles.blue).
		BorderBottom(true).
		Bold(false)
	st.Selected = st.Selected.
		Foreground(styles.white).
		Background(styles.blue).
		Bold(false)
	t.SetStyles(st)

	sidebar := &SidebarModel{table: t}
	sidebar.Update(UpdateFollows)

	return sidebar
}

func (s *SidebarModel) Update(msg tea.Msg) {
	switch msg {
	case UpdateFollows:
		follows := store.GetFollowedKeys(config.PublicKey)
		rows := make([]table.Row, len(follows))
		for i, f := range follows {
			rows[i] = table.Row{f}
		}

		s.table.SetRows(rows)
	}
}
