package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/VIctor-teles-Dev/code-cleaner/apps/url-shortener/internal/cache"
	"github.com/VIctor-teles-Dev/code-cleaner/apps/url-shortener/internal/shortener"
)

const (
	maxCreateBody   = 1 << 20 // 1 MB
	maxSlugAttempts = 5
)

type createURLRequest struct {
	OriginalURL string `json:"original_url"`
	CustomAlias string `json:"custom_alias"`
	ExpireAt    string `json:"expire_at"`
}

type createURLResponse struct {
	Slug        string     `json:"slug"`
	ShortURL    string     `json:"short_url"`
	OriginalURL string     `json:"original_url"`
	ExpiresAt   *time.Time `json:"expires_at"`
}

// CreateURL responde POST /api/v1/urls, protegido por bearer token. Sem custom
// alias, gera um slug base62 e reinsere em caso de colisão (23505).
func CreateURL(store shortener.Store, linkCache *cache.Cache, token, baseURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if store == nil || token == "" {
			writeJSON(w, http.StatusServiceUnavailable, apiError{Status: "unavailable"})
			return
		}
		if !authorized(r, token) {
			writeJSON(w, http.StatusUnauthorized, apiError{Status: "unauthorized"})
			return
		}

		var req createURLRequest
		r.Body = http.MaxBytesReader(w, r.Body, maxCreateBody)
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest,
				apiError{Status: "invalid", Error: "corpo da requisição inválido"})
			return
		}

		original := strings.TrimSpace(req.OriginalURL)
		if msg := shortener.ValidateURL(original); msg != "" {
			writeJSON(w, http.StatusBadRequest, apiError{Status: "invalid", Error: msg})
			return
		}
		alias := strings.TrimSpace(req.CustomAlias)
		if msg := shortener.ValidateAlias(alias); msg != "" {
			writeJSON(w, http.StatusBadRequest, apiError{Status: "invalid", Error: msg})
			return
		}

		expiresAt, msg := parseExpireAt(req.ExpireAt)
		if msg != "" {
			writeJSON(w, http.StatusBadRequest, apiError{Status: "invalid", Error: msg})
			return
		}

		link := shortener.Link{OriginalURL: original, ExpiresAt: expiresAt}
		status, err := createLink(r, store, linkCache, alias, &link)
		if err != nil {
			writeJSON(w, status, apiError{Status: statusText(status)})
			return
		}
		if status == http.StatusConflict {
			writeJSON(w, status, apiError{Status: "conflict", Error: "esse alias já está em uso"})
			return
		}

		writeJSON(w, http.StatusCreated, createURLResponse{
			Slug:        link.Slug,
			ShortURL:    shortURL(baseURL, link.Slug),
			OriginalURL: link.OriginalURL,
			ExpiresAt:   link.ExpiresAt,
		})
	}
}

// createLink persiste o link (alias custom ou slug gerado). Retorna
// (StatusConflict, nil) para alias duplicado e (5xx, err) para falhas reais;
// (0, nil) em sucesso.
func createLink(r *http.Request, store shortener.Store, linkCache *cache.Cache, alias string, link *shortener.Link) (int, error) {
	if alias != "" {
		link.Slug = alias
		err := store.Create(r.Context(), *link)
		switch {
		case errors.Is(err, shortener.ErrDuplicateSlug):
			return http.StatusConflict, nil
		case err != nil:
			log.Printf("urls: create (alias) failed: %v", err)
			return http.StatusInternalServerError, err
		}
		invalidate(linkCache, link.Slug)
		return 0, nil
	}

	for attempt := 0; attempt < maxSlugAttempts; attempt++ {
		slug, err := shortener.GenerateSlug()
		if err != nil {
			log.Printf("urls: slug gen failed: %v", err)
			return http.StatusInternalServerError, err
		}
		link.Slug = slug
		err = store.Create(r.Context(), *link)
		if errors.Is(err, shortener.ErrDuplicateSlug) {
			continue // colisão rara: tenta outro slug
		}
		if err != nil {
			log.Printf("urls: create failed: %v", err)
			return http.StatusInternalServerError, err
		}
		invalidate(linkCache, link.Slug)
		return 0, nil
	}

	log.Printf("urls: esgotou %d tentativas de slug", maxSlugAttempts)
	return http.StatusInternalServerError, errors.New("slug attempts exhausted")
}

func invalidate(linkCache *cache.Cache, slug string) {
	if linkCache != nil {
		linkCache.Delete(slug) // limpa eventual entrada de negative cache
	}
}

func parseExpireAt(raw string) (*time.Time, string) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, ""
	}
	t, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return nil, "expire_at deve ser uma data ISO 8601 (RFC 3339)"
	}
	if !t.After(time.Now()) {
		return nil, "expire_at deve estar no futuro"
	}
	utc := t.UTC()
	return &utc, ""
}

func shortURL(baseURL, slug string) string {
	base := strings.TrimRight(baseURL, "/")
	if base == "" {
		return "/" + slug
	}
	return base + "/" + slug
}

func statusText(status int) string {
	if status >= 500 {
		return "error"
	}
	return "invalid"
}
