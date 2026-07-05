package handler

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/VIctor-teles-Dev/code-cleaner/apps/url-shortener/internal/cache"
	"github.com/VIctor-teles-Dev/code-cleaner/apps/url-shortener/internal/shortener"
)

// clickEnqueuer enfileira um evento de clique (satisfeito por *analytics.Worker).
type clickEnqueuer interface {
	Enqueue(shortener.ClickEvent)
}

// Redirect responde GET /{slug} com um 302 para a URL original. É o caminho
// quente: tenta o cache antes do banco e enfileira o clique de forma assíncrona.
func Redirect(store shortener.Store, linkCache *cache.Cache, worker clickEnqueuer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if store == nil {
			http.Error(w, "serviço indisponível", http.StatusServiceUnavailable)
			return
		}

		slug := r.PathValue("slug")
		res, err := resolve(r.Context(), store, linkCache, slug)
		if err != nil {
			log.Printf("redirect: resolve failed: %v", err)
			http.Error(w, "erro interno", http.StatusInternalServerError)
			return
		}
		if !res.Found {
			http.NotFound(w, r)
			return
		}
		if res.Expired(time.Now()) {
			w.Header().Set("Cache-Control", "no-store")
			http.Error(w, "link expirado", http.StatusGone)
			return
		}

		if worker != nil {
			worker.Enqueue(shortener.ClickEvent{
				Slug:      slug,
				ClickedAt: time.Now().UTC(),
				IP:        clientIP(r),
				UserAgent: r.UserAgent(),
				Referrer:  r.Referer(),
			})
		}

		// 302 (Found), nunca 301: um 301 é cacheado pelo browser e os cliques
		// seguintes não passariam pelo servidor, destruindo a analytics.
		w.Header().Set("Cache-Control", "no-store")
		http.Redirect(w, r, res.OriginalURL, http.StatusFound)
	}
}

// resolve busca o slug no cache e, no miss, no banco — populando o cache
// (inclusive negative cache para 404). Erro real de banco vira err (-> 500);
// "não existe" vira Resolution{Found: false}.
func resolve(ctx context.Context, store shortener.Store, linkCache *cache.Cache, slug string) (shortener.Resolution, error) {
	if linkCache != nil {
		if res, ok := linkCache.Get(slug); ok {
			return res, nil
		}
	}

	link, err := store.Resolve(ctx, slug)
	var res shortener.Resolution
	switch {
	case errors.Is(err, shortener.ErrNotFound):
		res = shortener.Resolution{Found: false}
	case err != nil:
		return shortener.Resolution{}, err
	default:
		res = shortener.Resolution{
			OriginalURL: link.OriginalURL,
			ExpiresAt:   link.ExpiresAt,
			Found:       true,
		}
	}

	if linkCache != nil {
		linkCache.Put(slug, res)
	}
	return res, nil
}

// clientIP extrai o IP do cliente, preferindo o primeiro salto de
// X-Forwarded-For (setado pelo ingress-nginx) e caindo para RemoteAddr.
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if i := strings.IndexByte(xff, ','); i >= 0 {
			return strings.TrimSpace(xff[:i])
		}
		return strings.TrimSpace(xff)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
