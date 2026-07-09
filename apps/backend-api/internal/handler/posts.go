package handler

import (
	"crypto/subtle"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/VIctor-teles-Dev/code-cleaner/apps/backend-api/internal/blog"
)

type tagJSON struct {
	Slug string `json:"slug"`
	Name string `json:"name"`
}

func toTagJSON(tags []blog.Tag) []tagJSON {
	out := make([]tagJSON, 0, len(tags))
	for _, t := range tags {
		out = append(out, tagJSON{Slug: t.Slug, Name: t.Name})
	}
	return out
}

type postSummary struct {
	Slug        string     `json:"slug"`
	Title       string     `json:"title"`
	Excerpt     string     `json:"excerpt"`
	PublishedAt *time.Time `json:"published_at"`
	Tags        []tagJSON  `json:"tags"`
}

type postDetail struct {
	Slug        string     `json:"slug"`
	Title       string     `json:"title"`
	Content     string     `json:"content"`
	PublishedAt *time.Time `json:"published_at"`
	Tags        []tagJSON  `json:"tags"`
}

const excerptMaxRunes = 180

// excerpt extrai a primeira linha de texto do markdown (pulando títulos)
// e trunca no limite, para a listagem não carregar o conteúdo inteiro.
func excerpt(content string) string {
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if utf8.RuneCountInString(line) <= excerptMaxRunes {
			return line
		}
		runes := []rune(line)
		return strings.TrimSpace(string(runes[:excerptMaxRunes])) + "…"
	}
	return ""
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

type apiError struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

func toSummaries(posts []blog.Post) []postSummary {
	summaries := make([]postSummary, 0, len(posts))
	for _, p := range posts {
		summaries = append(summaries, postSummary{
			Slug:        p.Slug,
			Title:       p.Title,
			Excerpt:     excerpt(p.Content),
			PublishedAt: p.PublishedAt,
			Tags:        toTagJSON(p.Tags),
		})
	}
	return summaries
}

func toPostDetail(p blog.Post) postDetail {
	return postDetail{
		Slug:        p.Slug,
		Title:       p.Title,
		Content:     p.Content,
		PublishedAt: p.PublishedAt,
		Tags:        toTagJSON(p.Tags),
	}
}

// ListPosts responde GET /posts com os posts publicados.
func ListPosts(store blog.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if store == nil {
			writeJSON(w, http.StatusServiceUnavailable, apiError{Status: "unavailable"})
			return
		}

		posts, err := store.ListPublished(r.Context(), r.URL.Query().Get("tag"))
		if err != nil {
			log.Printf("posts: list failed: %v", err)
			writeJSON(w, http.StatusInternalServerError, apiError{Status: "error"})
			return
		}

		writeJSON(w, http.StatusOK, toSummaries(posts))
	}
}

// ListAllPosts responde GET /admin/posts com todos os posts, inclusive
// rascunhos. Protegido por bearer token (uso do admin).
func ListAllPosts(store blog.Store, token string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !requireAuth(w, r, store, token) {
			return
		}

		posts, err := store.ListAll(r.Context())
		if err != nil {
			log.Printf("posts: admin list failed: %v", err)
			writeJSON(w, http.StatusInternalServerError, apiError{Status: "error"})
			return
		}

		writeJSON(w, http.StatusOK, toSummaries(posts))
	}
}

// GetPost responde GET /posts/{slug} com o post completo.
func GetPost(store blog.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if store == nil {
			writeJSON(w, http.StatusServiceUnavailable, apiError{Status: "unavailable"})
			return
		}

		post, err := store.GetPublishedBySlug(r.Context(), r.PathValue("slug"))
		if errors.Is(err, blog.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, apiError{Status: "not_found"})
			return
		}
		if err != nil {
			log.Printf("posts: get failed: %v", err)
			writeJSON(w, http.StatusInternalServerError, apiError{Status: "error"})
			return
		}

		writeJSON(w, http.StatusOK, toPostDetail(post))
	}
}

// GetAnyPost responde GET /admin/posts/{slug} com o post completo,
// publicado ou rascunho. Protegido por bearer token (edição no admin).
func GetAnyPost(store blog.Store, token string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !requireAuth(w, r, store, token) {
			return
		}

		post, err := store.GetBySlug(r.Context(), r.PathValue("slug"))
		if errors.Is(err, blog.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, apiError{Status: "not_found"})
			return
		}
		if err != nil {
			log.Printf("posts: admin get failed: %v", err)
			writeJSON(w, http.StatusInternalServerError, apiError{Status: "error"})
			return
		}

		writeJSON(w, http.StatusOK, toPostDetail(post))
	}
}

var slugPattern = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

type createPostRequest struct {
	Slug      string   `json:"slug"`
	Title     string   `json:"title"`
	Content   string   `json:"content"`
	Published bool     `json:"published"`
	Tags      []string `json:"tags"`
}

// updatePostRequest é o corpo do PUT /posts/{slug}: sem slug, que vem do path
// e é imutável.
type updatePostRequest struct {
	Title     string   `json:"title"`
	Content   string   `json:"content"`
	Published bool     `json:"published"`
	Tags      []string `json:"tags"`
}

const (
	maxTags     = 10
	maxTagRunes = 50
)

// validateFields valida o que create e update têm em comum (título, conteúdo
// e tags). Retorna string vazia quando válido.
func validateFields(title, content string, tags []string) string {
	switch {
	case strings.TrimSpace(title) == "":
		return "informe o título"
	case strings.TrimSpace(content) == "":
		return "informe o conteúdo"
	case len(tags) > maxTags:
		return "muitas tags (máximo 10)"
	}
	for _, tag := range tags {
		trimmed := strings.TrimSpace(tag)
		if trimmed == "" || utf8.RuneCountInString(trimmed) > maxTagRunes ||
			blog.Slugify(trimmed) == "" {
			return "tag inválida"
		}
	}
	return ""
}

func validatePost(req createPostRequest) string {
	if !slugPattern.MatchString(req.Slug) {
		return "slug inválido (use letras minúsculas, números e hífens)"
	}
	return validateFields(req.Title, req.Content, req.Tags)
}

// toTags normaliza nomes de tags (slug + nome original) para persistência.
func toTags(names []string) []blog.Tag {
	tags := make([]blog.Tag, 0, len(names))
	for _, name := range names {
		name = strings.TrimSpace(name)
		tags = append(tags, blog.Tag{Slug: blog.Slugify(name), Name: name})
	}
	return tags
}

// publishedAt devolve o carimbo de publicação: agora (UTC) quando published,
// nil para rascunho. Em updates, nil despublica e não-nil preserva a data
// original (ver PostStore.Update).
func publishedAt(published bool) *time.Time {
	if !published {
		return nil
	}
	now := time.Now().UTC()
	return &now
}

// authorized confere o header Authorization: Bearer <token> em tempo constante.
func authorized(r *http.Request, token string) bool {
	got, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
	return ok && subtle.ConstantTimeCompare([]byte(got), []byte(token)) == 1
}

// requireAuth garante store configurado e token válido, escrevendo a resposta
// de erro (503 sem configuração, 401 sem/errado token) e devolvendo false
// quando a requisição não deve prosseguir.
func requireAuth(w http.ResponseWriter, r *http.Request, store blog.Store, token string) bool {
	if store == nil || token == "" {
		writeJSON(w, http.StatusServiceUnavailable, apiError{Status: "unavailable"})
		return false
	}
	if !authorized(r, token) {
		writeJSON(w, http.StatusUnauthorized, apiError{Status: "unauthorized"})
		return false
	}
	return true
}

// CreatePost responde POST /posts, protegido por bearer token. Token vazio
// desabilita o endpoint (503) — escrita só é possível quando configurada.
func CreatePost(store blog.Store, token string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !requireAuth(w, r, store, token) {
			return
		}

		var req createPostRequest
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest,
				apiError{Status: "invalid", Error: "corpo da requisição inválido"})
			return
		}
		if msg := validatePost(req); msg != "" {
			writeJSON(w, http.StatusBadRequest, apiError{Status: "invalid", Error: msg})
			return
		}

		post := blog.Post{
			Slug:        req.Slug,
			Title:       strings.TrimSpace(req.Title),
			Content:     req.Content,
			PublishedAt: publishedAt(req.Published),
			Tags:        toTags(req.Tags),
		}

		err := store.Create(r.Context(), post)
		if errors.Is(err, blog.ErrDuplicateSlug) {
			writeJSON(w, http.StatusConflict,
				apiError{Status: "conflict", Error: "já existe um post com esse slug"})
			return
		}
		if err != nil {
			log.Printf("posts: create failed: %v", err)
			writeJSON(w, http.StatusInternalServerError, apiError{Status: "error"})
			return
		}

		writeJSON(w, http.StatusCreated, toPostDetail(post))
	}
}

// UpdatePost responde PUT /posts/{slug}: edita um post existente (o slug é
// imutável). Protegido por bearer token.
func UpdatePost(store blog.Store, token string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !requireAuth(w, r, store, token) {
			return
		}

		var req updatePostRequest
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest,
				apiError{Status: "invalid", Error: "corpo da requisição inválido"})
			return
		}
		if msg := validateFields(req.Title, req.Content, req.Tags); msg != "" {
			writeJSON(w, http.StatusBadRequest, apiError{Status: "invalid", Error: msg})
			return
		}

		post := blog.Post{
			Slug:        r.PathValue("slug"),
			Title:       strings.TrimSpace(req.Title),
			Content:     req.Content,
			PublishedAt: publishedAt(req.Published),
			Tags:        toTags(req.Tags),
		}

		err := store.Update(r.Context(), post)
		if errors.Is(err, blog.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, apiError{Status: "not_found"})
			return
		}
		if err != nil {
			log.Printf("posts: update failed: %v", err)
			writeJSON(w, http.StatusInternalServerError, apiError{Status: "error"})
			return
		}

		writeJSON(w, http.StatusOK, toPostDetail(post))
	}
}

// DeletePost responde DELETE /posts/{slug}. Protegido por bearer token.
func DeletePost(store blog.Store, token string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !requireAuth(w, r, store, token) {
			return
		}

		err := store.Delete(r.Context(), r.PathValue("slug"))
		if errors.Is(err, blog.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, apiError{Status: "not_found"})
			return
		}
		if err != nil {
			log.Printf("posts: delete failed: %v", err)
			writeJSON(w, http.StatusInternalServerError, apiError{Status: "error"})
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
