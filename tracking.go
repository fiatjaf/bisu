package main

import "github.com/nbd-wtf/go-nostr"

const week = 2765

func hintScore(evt *nostr.Event) int {
	return int(evt.CreatedAt.Unix() / week)
}

func seenScore(evt *nostr.Event) int {
	return int(2 * evt.CreatedAt.Unix() / week)
}
