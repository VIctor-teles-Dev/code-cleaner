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
		writeJSON(w, http.StatusOK, summaries)
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

		writeJSON(w, http.StatusOK, postDetail{
			Slug:        post.Slug,
			Title:       post.Title,
			Content:     post.Content,
			PublishedAt: post.PublishedAt,
			Tags:        toTagJSON(post.Tags),
		})
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

const (
	maxTags     = 10
	maxTagRunes = 50
)

func validatePost(req createPostRequest) string {
	switch {
	case !slugPattern.MatchString(req.Slug):
		return "slug inválido (use letras minúsculas, números e hífens)"
	case strings.TrimSpace(req.Title) == "":
		return "informe o título"
	case strings.TrimSpace(req.Content) == "":
		return "informe o conteúdo"
	case len(req.Tags) > maxTags:
		return "muitas tags (máximo 10)"
	}
	for _, tag := range req.Tags {
		trimmed := strings.TrimSpace(tag)
		if trimmed == "" || utf8.RuneCountInString(trimmed) > maxTagRunes ||
			blog.Slugify(trimmed) == "" {
			return "tag inválida"
		}
	}
	return ""
}

// CreatePost responde POST /posts, protegido por bearer token. Token vazio
// desabilita o endpoint (503) — escrita só é possível quando configurada.
func CreatePost(store blog.Store, token string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if store == nil || token == "" {
			writeJSON(w, http.StatusServiceUnavailable, apiError{Status: "unavailable"})
			return
		}

		got, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
		if !ok || subtle.ConstantTimeCompare([]byte(got), []byte(token)) != 1 {
			writeJSON(w, http.StatusUnauthorized, apiError{Status: "unauthorized"})
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
			Slug:    req.Slug,
			Title:   strings.TrimSpace(req.Title),
			Content: req.Content,
		}
		if req.Published {
			now := time.Now().UTC()
			post.PublishedAt = &now
		}
		for _, tag := range req.Tags {
			name := strings.TrimSpace(tag)
			post.Tags = append(post.Tags, blog.Tag{Slug: blog.Slugify(name), Name: name})
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

		writeJSON(w, http.StatusCreated, postDetail{
			Slug:        post.Slug,
			Title:       post.Title,
			Content:     post.Content,
			PublishedAt: post.PublishedAt,
			Tags:        toTagJSON(post.Tags),
		})
	}
}
