package main

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/arriqaaq/flashdb"
)

// shortUint64 is the same as short(), but returns the result as a uint64 number
func shortUint64(idOrPubkey string) uint64 {
	length := len(idOrPubkey)
	if length < 8 {
		panic("shortUint64 called with bad value")
		return 0
	}
	b, err := hex.DecodeString(idOrPubkey[length-8:])
	if err != nil {
		panic(err)
		return 0
	}
	return uint64(binary.BigEndian.Uint32(b))
}

// short takes just the last 8 characters of these strings
// (used just for redis keys).
// would have been better to take the first 8, but people are doing proof-of-work
// on pubkeys and event ids so that would break things.
func short(idOrPubkey string) string {
	// we must check the length here because in some cases we're passing
	// the keys already shortened to this so we can't ever be sure
	length := len(idOrPubkey)
	if length < 8 {
		panic(fmt.Errorf("short called with bad value: %s", idOrPubkey))
	}
	return idOrPubkey[length-8:]
}

func doneRecently(key string, howOften time.Duration) bool {
	var yes bool
	flash.Update(func(txn *flashdb.Tx) error {
		done, _ := txn.Get(key)
		if done == "1" {
			return nil
		} else {
			yes = false
			return txn.SetEx(key, "1", int64(howOften.Seconds()))
		}
	})
	return yes
}
