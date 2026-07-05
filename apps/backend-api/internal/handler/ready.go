package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

// Pinger is the dependency-check contract, satisfied by *sql.DB.
type Pinger interface {
	PingContext(ctx context.Context) error
}

type readyResponse struct {
	Status string `json:"status"`
}

// Ready reports whether the API can serve traffic, checking its dependencies.
// A nil pinger means there is no dependency to check (e.g. dev without database).
func Ready(db Pinger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if db != nil {
			ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
			defer cancel()
			if err := db.PingContext(ctx); err != nil {
				w.WriteHeader(http.StatusServiceUnavailable)
				_ = json.NewEncoder(w).Encode(readyResponse{Status: "unavailable"})
				return
			}
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(readyResponse{Status: "ready"})
	}
}
