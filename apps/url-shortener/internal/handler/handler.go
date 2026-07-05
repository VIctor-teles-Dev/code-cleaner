// Package handler expõe os handlers HTTP do encurtador. Cada handler é uma
// factory que fecha sobre suas dependências, com degradação graciosa (store
// nil -> 503), no mesmo estilo do backend-api.
package handler

import (
	"crypto/subtle"
	"encoding/json"
	"net/http"
	"strings"
)

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

type apiError struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

// authorized valida o bearer token em tempo constante.
func authorized(r *http.Request, token string) bool {
	got, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
	return ok && subtle.ConstantTimeCompare([]byte(got), []byte(token)) == 1
}
