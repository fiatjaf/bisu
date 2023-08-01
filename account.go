package main

import (
	"encoding/json"
	"net/http"
)

func verifyCredentialsHandler(w http.ResponseWriter, r *http.Request) {
	acct, err := profile.toAccount(r.Context(), &ToAccountOpts{WithSource: true})
	if err != nil {
		panic(err)
	}
	json.NewEncoder(w).Encode(acct)
}
