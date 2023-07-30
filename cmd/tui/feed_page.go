package main

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip19"
)

type feedPage struct {
	l       list.Model
	focused bool
}

type item struct{ *nostr.Event }

func (i item) Title() string {
	note, _ := nip19.EncodeNote(i.ID)
	return note
}

func (i item) Description() string {
	npub, _ := nip19.EncodePublicKey(i.PubKey)
	return npub + ": " + i.Content
}

func (i item) FilterValue() string { return i.Content }

func newFeedPage(title string, events []*nostr.Event) *feedPage {
	items := make([]list.Item, len(events))
	for i, event := range events {
		items[i] = item{event}
	}

	l := list.New(items, newNoteItemDelegate(), 0, 0)
	l.DisableQuitKeybindings()
	l.Styles.Title = l.Styles.Title.Background(styles.blue)
	l.Title = title

	return &feedPage{
		l: l,
	}
}

func (fp *feedPage) View(x, y int) string {
	fp.l.SetWidth(x)
	fp.l.SetHeight(y)

	return fp.l.View()
}

func (fp *feedPage) Init() tea.Cmd {
	return nil
}

func (fp *feedPage) Update(msg tea.Msg) (Page, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case []*nostr.Event:
		items := make([]list.Item, len(msg))
		for i, evt := range msg {
			items[i] = item{evt}
		}
		cmd = fp.l.SetItems(items)
	case *nostr.Event:
		before := fp.l.Items()
		after := insertItemIntoDescendingList(before, item{msg})
		cmd = fp.l.SetItems(after)
	}

	if fp.focused {
		l, updateCmd := fp.l.Update(msg)
		fp.l = l
		cmd = tea.Batch(cmd, updateCmd)
	}

	return fp, cmd
}

func (fp *feedPage) Focus() { fp.focused = true }
func (fp *feedPage) Blur()  { fp.focused = false }
