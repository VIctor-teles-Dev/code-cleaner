package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/VIctor-teles-Dev/code-cleaner/apps/url-shortener/internal/cache"
	"github.com/VIctor-teles-Dev/code-cleaner/apps/url-shortener/internal/shortener"
)

type fakeEnqueuer struct{ events []shortener.ClickEvent }

func (f *fakeEnqueuer) Enqueue(e shortener.ClickEvent) { f.events = append(f.events, e) }

// countingStore conta as chamadas a Resolve para provar o cache hit.
type countingStore struct {
	*fakeStore
	resolveN int
}

func (c *countingStore) Resolve(ctx context.Context, slug string) (shortener.Link, error) {
	c.resolveN++
	return c.fakeStore.Resolve(ctx, slug)
}

func serveRedirect(h http.Handler, slug string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/"+slug, nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

// redirectMux registra o padrão real para que r.PathValue("slug") funcione.
func redirectMux(store shortener.Store, c *cache.Cache, w clickEnqueuer) *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("GET /{slug}", Redirect(store, c, w))
	return mux
}

func freshCache() *cache.Cache { return cache.New(10, time.Minute, time.Minute) }

func TestRedirectFound(t *testing.T) {
	store := newFakeStore()
	store.links["abc"] = shortener.Link{Slug: "abc", OriginalURL: "https://go.dev"}
	enq := &fakeEnqueuer{}

	rec := serveRedirect(redirectMux(store, freshCache(), enq), "abc")
	if rec.Code != http.StatusFound { // 302, nunca 301
		t.Fatalf("status = %d, want 302", rec.Code)
	}
	if loc := rec.Header().Get("Location"); loc != "https://go.dev" {
		t.Errorf("Location = %q, want https://go.dev", loc)
	}
	if len(enq.events) != 1 {
		t.Errorf("esperava 1 clique enfileirado, tem %d", len(enq.events))
	}
}

func TestRedirectNotFound(t *testing.T) {
	rec := serveRedirect(redirectMux(newFakeStore(), freshCache(), &fakeEnqueuer{}), "missing")
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}
}

func TestRedirectExpired(t *testing.T) {
	store := newFakeStore()
	past := time.Now().Add(-time.Hour)
	store.links["old"] = shortener.Link{Slug: "old", OriginalURL: "https://go.dev", ExpiresAt: &past}
	rec := serveRedirect(redirectMux(store, freshCache(), &fakeEnqueuer{}), "old")
	if rec.Code != http.StatusGone {
		t.Errorf("status = %d, want 410", rec.Code)
	}
}

func TestRedirectCacheHitAvoidsStore(t *testing.T) {
	base := newFakeStore()
	base.links["abc"] = shortener.Link{Slug: "abc", OriginalURL: "https://go.dev"}
	store := &countingStore{fakeStore: base}
	mux := redirectMux(store, freshCache(), &fakeEnqueuer{})

	serveRedirect(mux, "abc")
	serveRedirect(mux, "abc")
	if store.resolveN != 1 {
		t.Errorf("Resolve chamado %d vezes, want 1 (a 2ª deve vir do cache)", store.resolveN)
	}
}

func TestRedirectServiceUnavailable(t *testing.T) {
	rec := serveRedirect(redirectMux(nil, freshCache(), &fakeEnqueuer{}), "abc")
	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want 503", rec.Code)
	}
}
