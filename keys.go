package main

import (
	"strconv"
)

func shouldFetchNip05Key(pk string) string       { return "sfnip5:" + short(pk) }
func shouldFetchReplKey(k int, pk string) string { return "sfrepl" + strconv.Itoa(k) + ":" + short(pk) }
func shouldFetchEventKey(id string) string       { return "sfevt:" + short(id) }
func shouldFetchZapParamsKey(pk string) string   { return "sfzp:" + short(pk) }
