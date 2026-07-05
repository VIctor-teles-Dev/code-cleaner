package handler

import (
	"errors"
	"log"
	"net/http"

	"github.com/VIctor-teles-Dev/code-cleaner/apps/url-shortener/internal/shortener"
)

// GetAnalytics responde GET /api/v1/analytics/{slug} com as estatísticas de
// clique. Protegido pelo mesmo bearer token da criação. Retorna 404 quando o
// slug não existe (distinto de um link real com zero cliques).
func GetAnalytics(store shortener.Store, stats shortener.AnalyticsStore, token string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if store == nil || stats == nil || token == "" {
			writeJSON(w, http.StatusServiceUnavailable, apiError{Status: "unavailable"})
			return
		}
		if !authorized(r, token) {
			writeJSON(w, http.StatusUnauthorized, apiError{Status: "unauthorized"})
			return
		}

		slug := r.PathValue("slug")
		if _, err := store.Resolve(r.Context(), slug); errors.Is(err, shortener.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, apiError{Status: "not_found"})
			return
		} else if err != nil {
			log.Printf("analytics: resolve failed: %v", err)
			writeJSON(w, http.StatusInternalServerError, apiError{Status: "error"})
			return
		}

		result, err := stats.Stats(r.Context(), slug)
		if err != nil {
			log.Printf("analytics: stats failed: %v", err)
			writeJSON(w, http.StatusInternalServerError, apiError{Status: "error"})
			return
		}

		writeJSON(w, http.StatusOK, result)
	}
}
