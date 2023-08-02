package main

import (
	"github.com/nbd-wtf/go-nostr"
	"mvdan.cc/xurls/v2"
)

var urlMatcher = xurls.Strict()

type Note struct {
	*nostr.Event
}
