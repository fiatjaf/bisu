package main

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/truncate"
	"github.com/muesli/reflow/wordwrap"
	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip19"
)

type NoteItemDelegate struct {
	Normal   NoteItemDelegateStyles
	Selected NoteItemDelegateStyles
}

type NoteItemDelegateStyles struct {
	NameStyle lipgloss.Style
	DateStyle lipgloss.Style
	TextStyle lipgloss.Style
}

func newNoteItemDelegate() *NoteItemDelegate {
	nid := &NoteItemDelegate{
		Normal: NoteItemDelegateStyles{
			NameStyle: lipgloss.NewStyle().
				Width(65).
				PaddingLeft(2).
				Align(lipgloss.Left).
				Bold(true).
				Foreground(styles.yellow),
			DateStyle: lipgloss.NewStyle().
				PaddingRight(2).
				Align(lipgloss.Right).
				Foreground(styles.dimmed),
			TextStyle: lipgloss.NewStyle().
				Padding(0, 0, 0, 2),
		},
	}

	nid.Selected.NameStyle = nid.Normal.NameStyle.Copy().
		Foreground(styles.green).
		BorderLeft(true).
		BorderForeground(styles.green).
		PaddingLeft(1).
		BorderStyle(lipgloss.ThickBorder())
	nid.Selected.DateStyle = nid.Normal.DateStyle.Copy().
		PaddingRight(3). // because the left border is not accounted for in the width calculation below
		Foreground(styles.green)
	nid.Selected.TextStyle = nid.Normal.TextStyle.Copy().
		Foreground(styles.green).
		BorderLeft(true).
		BorderForeground(styles.green).
		PaddingLeft(1).
		BorderStyle(lipgloss.ThickBorder())

	return nid
}

// Render renders the item's view.
func (nid *NoteItemDelegate) Render(w io.Writer, m list.Model, index int, li list.Item) {
	if m.Width() <= 0 {
		// short-circuit
		return
	}

	event := li.(item).Event
	isSelected := index == m.Index()
	noteStyles := nid.Normal
	if isSelected {
		noteStyles = nid.Selected
	}

	var name string
	metadata := store.GetReplaceableEvent(context.Background(), event.PubKey, 0)
	if metadata != nil {
		profile, _ := nostr.ParseMetadata(*metadata)
		if profile != nil {
			name = profile.Name
		}
	}
	if name == "" {
		name, _ = nip19.EncodePublicKey(event.PubKey)
	}

	date := event.CreatedAt.Format("2006-01-02 15:04:05")

	fmt.Fprint(w,
		noteStyles.NameStyle.
			Render(truncate.StringWithTail(name, 64, "…")),
	)
	fmt.Fprint(w,
		noteStyles.DateStyle.
			Width(m.Width()-63-2-2).
			Render(date),
	)
	fmt.Fprint(w, "\n")

	lines := make([]string, 3) // ensure all notes show exactly 3 lines of text, even if they're smaller
	i := 0
	for _, line := range strings.Split(
		wordwrap.String(takeFirstString(event.Content, (m.Width()-2-3)*3), m.Width()-2-3),
		"\n",
	) {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		lines[i] = line
		i++
		if i >= 3 {
			break
		}
	}
	text := strings.Join(lines, "\n")

	fmt.Fprint(w,
		noteStyles.TextStyle.
			Render(text),
	)
}

// Height is the height of the list item.
func (_ *NoteItemDelegate) Height() int {
	return 4
}

// Spacing is the size of the horizontal gap between list items in cells.
func (_ *NoteItemDelegate) Spacing() int {
	return 1
}

// Update is the update loop for items. All messages in the list's update
// loop will pass through here except when the user is setting a filter.
// Use this method to perform item-level updates appropriate to this
// delegate.
func (_ *NoteItemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
}
