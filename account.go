package main

import (
	"encoding/json"
	"net/http"
)

func verifyCredentialsHandler(w http.ResponseWriter, r *http.Request) {
	acct := toAccount(r.Context(), profile, &ToAccountOpts{WithSource: true})
	json.NewEncoder(w).Encode(acct)
}
