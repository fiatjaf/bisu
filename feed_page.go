package main

import (
	"log"

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

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.DisableQuitKeybindings()
	l.Title = title
	log.Print(l.VisibleItems())

	return &feedPage{
		l: l,
	}
}

func (fp feedPage) View(x, y int) string {
	fp.l.SetWidth(x)
	fp.l.SetHeight(y)

	return fp.l.View()
}

func (fp feedPage) Init() tea.Cmd {
	return nil
}

func (fp *feedPage) Update(msg tea.Msg) (page, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case []*nostr.Event:
		items := fp.l.Items()
		for _, event := range msg {
			items = insertItemIntoDescendingList(items, item{event})
		}
		cmd = fp.l.SetItems(items)
		log.Print(cmd)
		log.Print(items)
	case *nostr.Event:
		cmd = fp.l.SetItems(insertItemIntoDescendingList(fp.l.Items(), item{msg}))
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
