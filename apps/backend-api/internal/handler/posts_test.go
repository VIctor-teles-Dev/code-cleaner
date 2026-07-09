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
	posts   []blog.Post
	err     error
	created []blog.Post
	updated []blog.Post
	deleted []string
	lastTag string
}

func (f *fakePostStore) ListPublished(_ context.Context, tagSlug string) ([]blog.Post, error) {
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

func (f *fakePostStore) ListAll(_ context.Context) ([]blog.Post, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.posts, nil
}

func (f *fakePostStore) GetPublishedBySlug(_ context.Context, slug string) (blog.Post, error) {
	if f.err != nil {
		return blog.Post{}, f.err
	}
	for _, p := range f.posts {
		if p.Slug == slug && p.PublishedAt != nil {
			return p, nil
		}
	}
	return blog.Post{}, blog.ErrNotFound
}

func (f *fakePostStore) GetBySlug(_ context.Context, slug string) (blog.Post, error) {
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

func (f *fakePostStore) Create(_ context.Context, post blog.Post) error {
	if f.err != nil {
		return f.err
	}
	for _, p := range f.posts {
		if p.Slug == post.Slug {
			return blog.ErrDuplicateSlug
		}
	}
	f.created = append(f.created, post)
	return nil
}

func (f *fakePostStore) Update(_ context.Context, post blog.Post) error {
	if f.err != nil {
		return f.err
	}
	for i, p := range f.posts {
		if p.Slug == post.Slug {
			f.posts[i] = post
			f.updated = append(f.updated, post)
			return nil
		}
	}
	return blog.ErrNotFound
}

func (f *fakePostStore) Delete(_ context.Context, slug string) error {
	if f.err != nil {
		return f.err
	}
	for i, p := range f.posts {
		if p.Slug == slug {
			f.posts = append(f.posts[:i], f.posts[i+1:]...)
			f.deleted = append(f.deleted, slug)
			return nil
		}
	}
	return blog.ErrNotFound
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

const validPost = `{"slug":"novo-post","title":"Novo","content":"corpo","published":true}`

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
		`{"slug":"rascunho","title":"Rascunho","content":"wip","published":false}`)

	if len(store.created) != 1 || store.created[0].PublishedAt != nil {
		t.Errorf("created = %+v, want one draft without PublishedAt", store.created)
	}
}

func TestCreatePostValidation(t *testing.T) {
	cases := []struct {
		label string
		body  string
	}{
		{"bad slug", `{"slug":"Com Espaço","title":"T","content":"c"}`},
		{"missing title", `{"slug":"ok","content":"c"}`},
		{"missing content", `{"slug":"ok","title":"T"}`},
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
		`{"slug":"com-tags","title":"T","content":"c","tags":["minhas aplicações","Go"]}`)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusCreated)
	}
	if len(store.created) != 1 || len(store.created[0].Tags) != 2 {
		t.Fatalf("created = %+v, want one post with 2 tags", store.created)
	}
	got := store.created[0].Tags[0]
	if got.Slug != "minhas-aplicacoes" || got.Name != "minhas aplicações" {
		t.Errorf("tag = %+v, want slugified slug and original name", got)
	}
}

func TestCreatePostRejectsInvalidTags(t *testing.T) {
	cases := []struct {
		label string
		body  string
	}{
		{"blank tag", `{"slug":"ok","title":"T","content":"c","tags":["  "]}`},
		{"symbol-only tag", `{"slug":"ok","title":"T","content":"c","tags":["!!!"]}`},
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
		`{"slug":"primeiro-post","title":"Duplicado","content":"c"}`)

	if rec.Code != http.StatusConflict {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusConflict)
	}
}

func draftPost() blog.Post {
	return blog.Post{Slug: "rascunho", Title: "Rascunho", Content: "wip"}
}

func adminListPosts(store blog.Store, token, auth string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/admin/posts", nil)
	if auth != "" {
		req.Header.Set("Authorization", auth)
	}
	rec := httptest.NewRecorder()
	ListAllPosts(store, token)(rec, req)
	return rec
}

func TestListAllPostsRequiresToken(t *testing.T) {
	store := &fakePostStore{posts: []blog.Post{publishedPost(), draftPost()}}

	if rec := adminListPosts(store, "segredo", "Bearer errado"); rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestListAllPostsIncludesDrafts(t *testing.T) {
	store := &fakePostStore{posts: []blog.Post{publishedPost(), draftPost()}}

	rec := adminListPosts(store, "segredo", "Bearer segredo")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	if !strings.Contains(body, `"slug":"rascunho"`) {
		t.Errorf("body = %q, want the draft slug (admin list includes drafts)", body)
	}
	if strings.Contains(body, `"content"`) {
		t.Errorf("body = %q, admin listing must not include full content", body)
	}
}

func adminGetPost(store blog.Store, token, auth, slug string) *httptest.ResponseRecorder {
	mux := http.NewServeMux()
	mux.Handle("GET /admin/posts/{slug}", GetAnyPost(store, token))
	req := httptest.NewRequest(http.MethodGet, "/admin/posts/"+slug, nil)
	if auth != "" {
		req.Header.Set("Authorization", auth)
	}
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	return rec
}

func TestGetAnyPostReturnsDraft(t *testing.T) {
	store := &fakePostStore{posts: []blog.Post{draftPost()}}

	rec := adminGetPost(store, "segredo", "Bearer segredo", "rascunho")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if body := rec.Body.String(); !strings.Contains(body, `"content":"wip"`) {
		t.Errorf("body = %q, want the draft content", body)
	}
}

func updatePost(store blog.Store, token, auth, slug, body string) *httptest.ResponseRecorder {
	mux := http.NewServeMux()
	mux.Handle("PUT /posts/{slug}", UpdatePost(store, token))
	req := httptest.NewRequest(http.MethodPut, "/posts/"+slug, strings.NewReader(body))
	if auth != "" {
		req.Header.Set("Authorization", auth)
	}
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	return rec
}

func TestUpdatePostRequiresToken(t *testing.T) {
	store := &fakePostStore{posts: []blog.Post{publishedPost()}}

	rec := updatePost(store, "segredo", "Bearer errado", "primeiro-post",
		`{"title":"Novo","content":"c"}`)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestUpdatePostEditsExisting(t *testing.T) {
	store := &fakePostStore{posts: []blog.Post{publishedPost()}}

	rec := updatePost(store, "segredo", "Bearer segredo", "primeiro-post",
		`{"title":"Título novo","content":"corpo novo","published":true,"tags":["Go"]}`)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if len(store.updated) != 1 {
		t.Fatalf("updated = %d posts, want 1", len(store.updated))
	}
	got := store.updated[0]
	if got.Slug != "primeiro-post" || got.Title != "Título novo" {
		t.Errorf("updated = %+v, want slug from path and new title", got)
	}
	if got.PublishedAt == nil {
		t.Error("PublishedAt = nil, want set when published=true")
	}
	if len(got.Tags) != 1 || got.Tags[0].Slug != "go" {
		t.Errorf("tags = %+v, want the slugified Go tag", got.Tags)
	}
}

func TestUpdatePostUnpublish(t *testing.T) {
	store := &fakePostStore{posts: []blog.Post{publishedPost()}}

	updatePost(store, "segredo", "Bearer segredo", "primeiro-post",
		`{"title":"T","content":"c","published":false}`)

	if len(store.updated) != 1 || store.updated[0].PublishedAt != nil {
		t.Errorf("updated = %+v, want PublishedAt nil when published=false", store.updated)
	}
}

func TestUpdatePostNotFound(t *testing.T) {
	rec := updatePost(&fakePostStore{}, "segredo", "Bearer segredo", "nao-existe",
		`{"title":"T","content":"c"}`)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestUpdatePostValidation(t *testing.T) {
	store := &fakePostStore{posts: []blog.Post{publishedPost()}}

	rec := updatePost(store, "segredo", "Bearer segredo", "primeiro-post",
		`{"title":"","content":"c"}`)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func deletePost(store blog.Store, token, auth, slug string) *httptest.ResponseRecorder {
	mux := http.NewServeMux()
	mux.Handle("DELETE /posts/{slug}", DeletePost(store, token))
	req := httptest.NewRequest(http.MethodDelete, "/posts/"+slug, nil)
	if auth != "" {
		req.Header.Set("Authorization", auth)
	}
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	return rec
}

func TestDeletePostRequiresToken(t *testing.T) {
	store := &fakePostStore{posts: []blog.Post{publishedPost()}}

	if rec := deletePost(store, "segredo", "", "primeiro-post"); rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
	if len(store.deleted) != 0 {
		t.Errorf("deleted = %v, want nothing removed without auth", store.deleted)
	}
}

func TestDeletePostRemoves(t *testing.T) {
	store := &fakePostStore{posts: []blog.Post{publishedPost()}}

	rec := deletePost(store, "segredo", "Bearer segredo", "primeiro-post")

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
	if len(store.deleted) != 1 || store.deleted[0] != "primeiro-post" {
		t.Errorf("deleted = %v, want the post slug", store.deleted)
	}
}

func TestDeletePostNotFound(t *testing.T) {
	rec := deletePost(&fakePostStore{}, "segredo", "Bearer segredo", "nao-existe")

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}
