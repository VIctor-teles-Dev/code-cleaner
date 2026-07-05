package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/VIctor-teles-Dev/code-cleaner/apps/backend-api/internal/blog"
)

type fakePostStore struct {
	posts      []blog.Post
	err        error
	created    []blog.PostInput
	lastTag    string
	lastLocale string
}

func (f *fakePostStore) ListPublished(_ context.Context, locale, tagSlug string) ([]blog.Post, error) {
	f.lastLocale = locale
	f.lastTag = tagSlug
	if f.err != nil {
		return nil, f.err
	}
	if tagSlug == "" {
		return f.posts, nil
	}
	var filtered []blog.Post
	for _, p := range f.posts {
		for _, t := range p.Tags {
			if t.Slug == tagSlug {
				filtered = append(filtered, p)
				break
			}
		}
	}
	return filtered, nil
}

func (f *fakePostStore) GetPublishedBySlug(_ context.Context, locale, slug string) (blog.Post, error) {
	f.lastLocale = locale
	if f.err != nil {
		return blog.Post{}, f.err
	}
	for _, p := range f.posts {
		if p.Slug == slug {
			return p, nil
		}
	}
	return blog.Post{}, blog.ErrNotFound
}

func (f *fakePostStore) Create(_ context.Context, post blog.PostInput) error {
	if f.err != nil {
		return f.err
	}
	for _, p := range f.posts {
		if p.Slug == post.Slug {
			return blog.ErrDuplicateSlug
		}
	}
	for _, c := range f.created {
		if c.Slug == post.Slug {
			return blog.ErrDuplicateSlug
		}
	}
	f.created = append(f.created, post)
	return nil
}

func publishedPost() blog.Post {
	published := time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)
	return blog.Post{
		Slug:        "primeiro-post",
		Title:       "Primeiro post",
		Content:     "# Título\n\nUm parágrafo de introdução.\n\nMais texto.",
		PublishedAt: &published,
		Tags:        []blog.Tag{{Slug: "minhas-aplicacoes", Name: "minhas aplicações"}},
	}
}

func TestListPostsReturnsSummaries(t *testing.T) {
	store := &fakePostStore{posts: []blog.Post{publishedPost()}}
	req := httptest.NewRequest(http.MethodGet, "/posts", nil)
	rec := httptest.NewRecorder()

	ListPosts(store)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	if !strings.Contains(body, `"slug":"primeiro-post"`) {
		t.Errorf("body = %q, want the post slug", body)
	}
	if !strings.Contains(body, `"excerpt":"Um parágrafo de introdução."`) {
		t.Errorf("body = %q, want the excerpt (first text line, skipping headings)", body)
	}
	if strings.Contains(body, `"content"`) {
		t.Errorf("body = %q, listing must not include full content", body)
	}
	if !strings.Contains(body, `"tags":[{"slug":"minhas-aplicacoes","name":"minhas aplicações"}]`) {
		t.Errorf("body = %q, want the post tags", body)
	}
}

func TestListPostsPassesLocale(t *testing.T) {
	store := &fakePostStore{}
	req := httptest.NewRequest(http.MethodGet, "/posts?locale=en", nil)
	ListPosts(store)(httptest.NewRecorder(), req)
	if store.lastLocale != "en" {
		t.Errorf("store.lastLocale = %q, want en", store.lastLocale)
	}
}

func TestListPostsDefaultsLocale(t *testing.T) {
	store := &fakePostStore{}
	req := httptest.NewRequest(http.MethodGet, "/posts?locale=zz", nil) // desconhecido
	ListPosts(store)(httptest.NewRecorder(), req)
	if store.lastLocale != blog.DefaultLocale {
		t.Errorf("store.lastLocale = %q, want %q (fallback)", store.lastLocale, blog.DefaultLocale)
	}
}

func TestListPostsFiltersByTag(t *testing.T) {
	store := &fakePostStore{posts: []blog.Post{publishedPost()}}
	req := httptest.NewRequest(http.MethodGet, "/posts?tag=outra-tag", nil)
	rec := httptest.NewRecorder()

	ListPosts(store)(rec, req)

	if store.lastTag != "outra-tag" {
		t.Errorf("store.lastTag = %q, want the query param", store.lastTag)
	}
	if body := strings.TrimSpace(rec.Body.String()); body != "[]" {
		t.Errorf("body = %q, want no posts for another tag", body)
	}
}

func TestListPostsEmpty(t *testing.T) {
	rec := httptest.NewRecorder()
	ListPosts(&fakePostStore{})(rec, httptest.NewRequest(http.MethodGet, "/posts", nil))

	if body := strings.TrimSpace(rec.Body.String()); body != "[]" {
		t.Errorf("body = %q, want empty json array", body)
	}
}

func getPost(t *testing.T, store blog.Store, slug string) *httptest.ResponseRecorder {
	t.Helper()
	mux := http.NewServeMux()
	mux.Handle("GET /posts/{slug}", GetPost(store))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/posts/"+slug, nil))
	return rec
}

func TestGetPostBySlug(t *testing.T) {
	store := &fakePostStore{posts: []blog.Post{publishedPost()}}

	rec := getPost(t, store, "primeiro-post")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if body := rec.Body.String(); !strings.Contains(body, `"content"`) {
		t.Errorf("body = %q, want full content", body)
	}
}

func TestGetPostNotFound(t *testing.T) {
	rec := getPost(t, &fakePostStore{}, "nao-existe")

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func createPost(store blog.Store, token, auth, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/posts", strings.NewReader(body))
	if auth != "" {
		req.Header.Set("Authorization", auth)
	}
	rec := httptest.NewRecorder()
	CreatePost(store, token)(rec, req)
	return rec
}

const validPost = `{"slug":"novo-post","published":true,"translations":{"pt-BR":{"title":"Novo","content":"corpo"}}}`

func TestCreatePostRequiresToken(t *testing.T) {
	store := &fakePostStore{}

	cases := []struct {
		label string
		auth  string
		want  int
	}{
		{"no auth header", "", http.StatusUnauthorized},
		{"wrong token", "Bearer errado", http.StatusUnauthorized},
		{"right token", "Bearer segredo", http.StatusCreated},
	}
	for _, tc := range cases {
		t.Run(tc.label, func(t *testing.T) {
			rec := createPost(store, "segredo", tc.auth, validPost)
			if rec.Code != tc.want {
				t.Errorf("status = %d, want %d", rec.Code, tc.want)
			}
		})
	}
}

func TestCreatePostDisabledWithoutToken(t *testing.T) {
	rec := createPost(&fakePostStore{}, "", "Bearer qualquer", validPost)

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusServiceUnavailable)
	}
}

func TestCreatePostSetsPublishedAt(t *testing.T) {
	store := &fakePostStore{}

	createPost(store, "segredo", "Bearer segredo", validPost)

	if len(store.created) != 1 {
		t.Fatalf("created = %d posts, want 1", len(store.created))
	}
	if store.created[0].PublishedAt == nil {
		t.Error("PublishedAt = nil, want set when published=true")
	}
}

func TestCreatePostDraftHasNoPublishedAt(t *testing.T) {
	store := &fakePostStore{}

	createPost(store, "segredo", "Bearer segredo",
		`{"slug":"rascunho","published":false,"translations":{"pt-BR":{"title":"Rascunho","content":"wip"}}}`)

	if len(store.created) != 1 || store.created[0].PublishedAt != nil {
		t.Errorf("created = %+v, want one draft without PublishedAt", store.created)
	}
}

func TestCreatePostStoresTranslations(t *testing.T) {
	store := &fakePostStore{}

	createPost(store, "segredo", "Bearer segredo",
		`{"slug":"bi","translations":{"pt-BR":{"title":"Olá","content":"c"},"en":{"title":"Hi","content":"c"}}}`)

	if len(store.created) != 1 {
		t.Fatalf("created = %d, want 1", len(store.created))
	}
	tr := store.created[0].Translations
	if tr["pt-BR"].Title != "Olá" || tr["en"].Title != "Hi" {
		t.Errorf("translations = %+v, want pt-BR e en", tr)
	}
}

func TestCreatePostValidation(t *testing.T) {
	cases := []struct {
		label string
		body  string
	}{
		{"bad slug", `{"slug":"Com Espaço","translations":{"pt-BR":{"title":"T","content":"c"}}}`},
		{"missing default title", `{"slug":"ok","translations":{"pt-BR":{"content":"c"}}}`},
		{"missing default content", `{"slug":"ok","translations":{"pt-BR":{"title":"T"}}}`},
		{"missing default locale", `{"slug":"ok","translations":{"en":{"title":"T","content":"c"}}}`},
		{"unknown locale", `{"slug":"ok","translations":{"pt-BR":{"title":"T","content":"c"},"zz":{"title":"x","content":"y"}}}`},
		{"malformed json", `{`},
	}
	for _, tc := range cases {
		t.Run(tc.label, func(t *testing.T) {
			rec := createPost(&fakePostStore{}, "segredo", "Bearer segredo", tc.body)
			if rec.Code != http.StatusBadRequest {
				t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
			}
		})
	}
}

func TestCreatePostWithTags(t *testing.T) {
	store := &fakePostStore{}

	rec := createPost(store, "segredo", "Bearer segredo",
		`{"slug":"com-tags","translations":{"pt-BR":{"title":"T","content":"c"}},`+
			`"tags":[{"slug":"arquitetura","names":{"pt-BR":"Arquitetura","en":"Architecture"}},`+
			`{"slug":"go","names":{"pt-BR":"Go","en":"Go"}}]}`)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusCreated, rec.Body.String())
	}
	if len(store.created) != 1 || len(store.created[0].Tags) != 2 {
		t.Fatalf("created = %+v, want one post with 2 tags", store.created)
	}
	got := store.created[0].Tags[0]
	if got.Slug != "arquitetura" || got.Names["pt-BR"] != "Arquitetura" || got.Names["en"] != "Architecture" {
		t.Errorf("tag = %+v, want slug + nomes por idioma", got)
	}
}

func TestCreatePostRejectsInvalidTags(t *testing.T) {
	cases := []struct {
		label string
		body  string
	}{
		{"bad tag slug", `{"slug":"ok","translations":{"pt-BR":{"title":"T","content":"c"}},"tags":[{"slug":"Com Espaço","names":{"pt-BR":"x"}}]}`},
		{"blank tag name", `{"slug":"ok","translations":{"pt-BR":{"title":"T","content":"c"}},"tags":[{"slug":"go","names":{"pt-BR":"  "}}]}`},
	}
	for _, tc := range cases {
		t.Run(tc.label, func(t *testing.T) {
			rec := createPost(&fakePostStore{}, "segredo", "Bearer segredo", tc.body)
			if rec.Code != http.StatusBadRequest {
				t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
			}
		})
	}
}

func TestCreatePostDuplicateSlug(t *testing.T) {
	store := &fakePostStore{posts: []blog.Post{publishedPost()}}

	rec := createPost(store, "segredo", "Bearer segredo",
		`{"slug":"primeiro-post","translations":{"pt-BR":{"title":"Duplicado","content":"c"}}}`)

	if rec.Code != http.StatusConflict {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusConflict)
	}
}
