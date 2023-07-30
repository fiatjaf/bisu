package main

import (
	"testing"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/nbd-wtf/go-nostr"
)

func TestInsertIntoSortedList(t *testing.T) {
	var arr []list.Item
	arr = insertItemIntoDescendingList(arr, item{&nostr.Event{ID: "a", CreatedAt: time.Unix(0, 0)}})
	if len(arr) != 1 {
		t.Fatal("arr should have one element")
	}
	arr = insertItemIntoDescendingList(arr, item{&nostr.Event{ID: "b", CreatedAt: time.Unix(2, 0)}})
	if len(arr) != 2 {
		t.Fatal("arr should have two elements")
	}
	arr = insertItemIntoDescendingList(arr, item{&nostr.Event{ID: "d", CreatedAt: time.Unix(4, 0)}})
	if len(arr) != 3 {
		t.Fatal("arr should have three elements")
	}
	arr = insertItemIntoDescendingList(arr, item{&nostr.Event{ID: "b", CreatedAt: time.Unix(2, 0)}})
	if len(arr) != 3 {
		t.Fatal("arr should have three elements")
	}
	arr = insertItemIntoDescendingList(arr, item{&nostr.Event{ID: "g", CreatedAt: time.Unix(7, 0)}})
	if len(arr) != 4 {
		t.Fatal("arr should have four elements")
	}
	if arr[0].(item).ID != "g" || arr[2].(item).ID != "b" {
		t.Fatal("g should come first and b third: ", printableSortedArray(arr))
	}
	arr = insertItemIntoDescendingList(arr, item{&nostr.Event{ID: "f", CreatedAt: time.Unix(6, 0)}})
	if len(arr) != 5 {
		t.Fatal("arr should have five elements")
	}
	if arr[0].(item).ID != "g" || arr[1].(item).ID != "f" {
		t.Fatal("g should come first and f second: ", printableSortedArray(arr))
	}
	arr = insertItemIntoDescendingList(arr, item{&nostr.Event{ID: "_", CreatedAt: time.Unix(-1, 0)}})
	if len(arr) != 6 {
		t.Fatal("arr should have six elements")
	}
	if arr[0].(item).ID != "g" || arr[5].(item).ID != "_" {
		t.Fatal("g should come first and 0 last: ", printableSortedArray(arr))
	}
	arr = insertItemIntoDescendingList(arr, item{&nostr.Event{ID: "c", CreatedAt: time.Unix(3, 0)}})
	if len(arr) != 7 {
		t.Fatal("arr should have seven elements")
	}
	arr = insertItemIntoDescendingList(arr, item{&nostr.Event{ID: "b", CreatedAt: time.Unix(2, 0)}})
	if len(arr) != 7 {
		t.Fatal("arr should still have seven elements")
	}
	arr = insertItemIntoDescendingList(arr, item{&nostr.Event{ID: "b", CreatedAt: time.Unix(2, 0)}})
	if len(arr) != 7 {
		t.Fatal("arr should still have seven elements")
	}
	arr = insertItemIntoDescendingList(arr, item{&nostr.Event{ID: "a", CreatedAt: time.Unix(1, 0)}})
	if len(arr) != 7 {
		t.Fatal("arr should still have seven elements")
	}
	if arr[0].(item).ID != "g" || arr[6].(item).ID != "_" || arr[3].(item).ID != "c" {
		t.Fatal("g should come first, 0 last and c should be the fourth: ", printableSortedArray(arr))
	}
	arr = insertItemIntoDescendingList(arr, item{&nostr.Event{ID: "z", CreatedAt: time.Unix(7, 0)}})
	if len(arr) != 8 {
		t.Fatal("arr should have eight elements")
	}
	if arr[0].(item).ID != "z" {
		t.Fatal("z should come first: ", printableSortedArray(arr))
	}
}

func printableSortedArray(arr []list.Item) []string {
	p := make([]string, len(arr))
	for i, it := range arr {
		v := "nil"
		if it != nil {
			v = it.(item).ID[0:3]
		}
		p[i] = v
	}
	return p
}
