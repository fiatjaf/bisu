package main

import (
	"strconv"
)

func replaceableKey(k int, pk string) string { return "k" + strconv.Itoa(k) + ":" + short(pk) }

func shouldFetchNip05Key(pk string) string       { return "sfnip5:" + short(pk) }
func shouldFetchReplKey(k int, pk string) string { return "sfrepl" + strconv.Itoa(k) + ":" + short(pk) }
func shouldFetchEventKey(id string) string       { return "sfevt:" + short(id) }
func shouldFetchZapParamsKey(pk string) string   { return "sfzp:" + short(pk) }
