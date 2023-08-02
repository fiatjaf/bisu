package main

import (
	"encoding/json"
	"net/http"
)

func verifyCredentialsHandler(w http.ResponseWriter, r *http.Request) {
	acct := profile.toAccount(r.Context(), &ToAccountOpts{WithSource: true})
	json.NewEncoder(w).Encode(acct)
}
