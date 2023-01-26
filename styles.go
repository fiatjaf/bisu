package main

import "github.com/charmbracelet/lipgloss"

type globalStyles struct {
	blue   lipgloss.Color
	red    lipgloss.Color
	yellow lipgloss.Color
	green  lipgloss.Color
	white  lipgloss.Color
	dimmed lipgloss.Color
}

var styles = globalStyles{
	blue:   lipgloss.Color("#634cc2"),
	red:    lipgloss.Color("#c24c70"),
	yellow: lipgloss.Color("#abc24c"),
	green:  lipgloss.Color("#4cc29e"),
	white:  lipgloss.Color("#ffffff"),
	dimmed: lipgloss.Color("#dddddd"),
}
