package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/VIctor-teles-Dev/code-cleaner/apps/url-shortener/internal/cache"
	"github.com/VIctor-teles-Dev/code-cleaner/apps/url-shortener/internal/shortener"
)

const adminToken = "segredo"

// fakeStore é um shortener.Store em memória, compartilhado pelos testes de
// handler (urls, redirect e analytics).
type fakeStore struct {
	links     map[string]shortener.Link
	err       error
	alwaysDup bool
	createN   int
}

func newFakeStore() *fakeStore { return &fakeStore{links: map[string]shortener.Link{}} }

func (f *fakeStore) Create(_ context.Context, link shortener.Link) error {
	f.createN++
	if f.err != nil {
		return f.err
	}
	if f.alwaysDup {
		return shortener.ErrDuplicateSlug
	}
	if _, ok := f.links[link.Slug]; ok {
		return shortener.ErrDuplicateSlug
	}
	f.links[link.Slug] = link
	return nil
}

func (f *fakeStore) Resolve(_ context.Context, slug string) (shortener.Link, error) {
	if f.err != nil {
		return shortener.Link{}, f.err
	}
	link, ok := f.links[slug]
	if !ok {
		return shortener.Link{}, shortener.ErrNotFound
	}
	return link, nil
}

func doCreate(h http.HandlerFunc, token, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/urls", strings.NewReader(body))
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h(rec, req)
	return rec
}

func newCreateHandler(store shortener.Store) http.HandlerFunc {
	return CreateURL(store, cache.New(10, time.Minute, time.Minute), adminToken, "https://curto.ccl.app.br")
}

func TestCreateURLAuth(t *testing.T) {
	h := newCreateHandler(newFakeStore())
	body := `{"original_url":"https://go.dev"}`
	if rec := doCreate(h, "", body); rec.Code != http.StatusUnauthorized {
		t.Errorf("sem token: status = %d, want 401", rec.Code)
	}
	if rec := doCreate(h, "errado", body); rec.Code != http.StatusUnauthorized {
		t.Errorf("token errado: status = %d, want 401", rec.Code)
	}
}

func TestCreateURLDisabled(t *testing.T) {
	// token vazio desabilita o endpoint
	if rec := doCreate(CreateURL(newFakeStore(), nil, "", ""), "", `{}`); rec.Code != http.StatusServiceUnavailable {
		t.Errorf("token vazio: status = %d, want 503", rec.Code)
	}
	// store nil também
	if rec := doCreate(CreateURL(nil, nil, adminToken, ""), adminToken, `{}`); rec.Code != http.StatusServiceUnavailable {
		t.Errorf("store nil: status = %d, want 503", rec.Code)
	}
}

func TestCreateURLCustomAlias(t *testing.T) {
	store := newFakeStore()
	rec := doCreate(newCreateHandler(store), adminToken,
		`{"original_url":"https://go.dev","custom_alias":"minha-marca"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201 (body=%s)", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	if !strings.Contains(body, `"slug":"minha-marca"`) {
		t.Errorf("body = %s, want slug minha-marca", body)
	}
	if !strings.Contains(body, `"short_url":"https://curto.ccl.app.br/minha-marca"`) {
		t.Errorf("body = %s, want short_url composta com BASE_URL", body)
	}
}

func TestCreateURLDuplicateAlias(t *testing.T) {
	store := newFakeStore()
	store.links["taken"] = shortener.Link{Slug: "taken"}
	rec := doCreate(newCreateHandler(store), adminToken,
		`{"original_url":"https://go.dev","custom_alias":"taken"}`)
	if rec.Code != http.StatusConflict {
		t.Errorf("status = %d, want 409", rec.Code)
	}
}

func TestCreateURLValidation(t *testing.T) {
	cases := []struct {
		name string
		body string
	}{
		{"url vazia", `{"original_url":""}`},
		{"scheme inválido", `{"original_url":"ftp://x.com"}`},
		{"alias reservado", `{"original_url":"https://go.dev","custom_alias":"api"}`},
		{"alias inválido", `{"original_url":"https://go.dev","custom_alias":"com espaço"}`},
		{"expiração no passado", `{"original_url":"https://go.dev","expire_at":"2000-01-01T00:00:00Z"}`},
		{"json malformado", `{`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := doCreate(newCreateHandler(newFakeStore()), adminToken, tc.body)
			if rec.Code != http.StatusBadRequest {
				t.Errorf("status = %d, want 400", rec.Code)
			}
		})
	}
}

func TestCreateURLRandomSlug(t *testing.T) {
	store := newFakeStore()
	rec := doCreate(newCreateHandler(store), adminToken, `{"original_url":"https://go.dev"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201", rec.Code)
	}
	if len(store.links) != 1 {
		t.Fatalf("esperava 1 link criado, tem %d", len(store.links))
	}
	for slug := range store.links {
		if len(slug) != 7 {
			t.Errorf("slug gerado = %q (len %d), want 7 chars", slug, len(slug))
		}
	}
}

func TestCreateURLSlugExhaustion(t *testing.T) {
	store := newFakeStore()
	store.alwaysDup = true
	rec := doCreate(newCreateHandler(store), adminToken, `{"original_url":"https://go.dev"}`)
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", rec.Code)
	}
	if store.createN != maxSlugAttempts {
		t.Errorf("tentativas = %d, want %d", store.createN, maxSlugAttempts)
	}
}
