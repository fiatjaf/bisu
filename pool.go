package main

import (
	"context"

	"github.com/nbd-wtf/go-nostr"
)

var pool = nostr.NewSimplePool(context.Background())
