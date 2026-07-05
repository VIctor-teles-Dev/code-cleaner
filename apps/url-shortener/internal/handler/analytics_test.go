package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/VIctor-teles-Dev/code-cleaner/apps/url-shortener/internal/shortener"
)

type fakeStats struct {
	result shortener.Analytics
	err    error
}

func (f fakeStats) Stats(_ context.Context, _ string) (shortener.Analytics, error) {
	return f.result, f.err
}

func serveAnalytics(h http.Handler, slug, token string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/analytics/"+slug, nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func analyticsMux(store shortener.Store, stats shortener.AnalyticsStore, token string) *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("GET /api/v1/analytics/{slug}", GetAnalytics(store, stats, token))
	return mux
}

func TestGetAnalyticsOK(t *testing.T) {
	store := newFakeStore()
	store.links["abc"] = shortener.Link{Slug: "abc", OriginalURL: "https://go.dev"}
	stats := fakeStats{result: shortener.Analytics{Slug: "abc", TotalClicks: 42}}

	rec := serveAnalytics(analyticsMux(store, stats, adminToken), "abc", adminToken)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"total_clicks":42`) {
		t.Errorf("body = %s, want total_clicks 42", rec.Body.String())
	}
}

func TestGetAnalyticsNotFound(t *testing.T) {
	rec := serveAnalytics(analyticsMux(newFakeStore(), fakeStats{}, adminToken), "missing", adminToken)
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}
}

func TestGetAnalyticsAuth(t *testing.T) {
	store := newFakeStore()
	store.links["abc"] = shortener.Link{Slug: "abc"}
	rec := serveAnalytics(analyticsMux(store, fakeStats{}, adminToken), "abc", "")
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("sem token: status = %d, want 401", rec.Code)
	}
}

func TestGetAnalyticsDisabled(t *testing.T) {
	rec := serveAnalytics(analyticsMux(newFakeStore(), fakeStats{}, ""), "abc", "")
	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want 503", rec.Code)
	}
}
