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

// knownLocales são os idiomas aceitos (espelha os locales do frontend).
var knownLocales = map[string]bool{"pt-BR": true, "en": true}

// resolveLocale lê ?locale=; cai para o default quando ausente/desconhecido.
func resolveLocale(r *http.Request) string {
	if loc := r.URL.Query().Get("locale"); knownLocales[loc] {
		return loc
	}
	return blog.DefaultLocale
}

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

// ListPosts responde GET /posts com os posts publicados no locale pedido.
func ListPosts(store blog.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if store == nil {
			writeJSON(w, http.StatusServiceUnavailable, apiError{Status: "unavailable"})
			return
		}

		posts, err := store.ListPublished(r.Context(), resolveLocale(r), r.URL.Query().Get("tag"))
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

// GetPost responde GET /posts/{slug} com o post completo no locale pedido.
func GetPost(store blog.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if store == nil {
			writeJSON(w, http.StatusServiceUnavailable, apiError{Status: "unavailable"})
			return
		}

		post, err := store.GetPublishedBySlug(r.Context(), resolveLocale(r), r.PathValue("slug"))
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

type translationJSON struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

type tagInputJSON struct {
	Slug  string            `json:"slug"`
	Names map[string]string `json:"names"`
}

type createPostRequest struct {
	Slug         string                     `json:"slug"`
	Published    bool                       `json:"published"`
	Translations map[string]translationJSON `json:"translations"`
	Tags         []tagInputJSON             `json:"tags"`
}

const (
	maxTags     = 10
	maxTagRunes = 50
)

func validatePost(req createPostRequest) string {
	if !slugPattern.MatchString(req.Slug) {
		return "slug inválido (use letras minúsculas, números e hífens)"
	}
	def, ok := req.Translations[blog.DefaultLocale]
	if !ok || strings.TrimSpace(def.Title) == "" {
		return "informe o título no idioma padrão (" + blog.DefaultLocale + ")"
	}
	if strings.TrimSpace(def.Content) == "" {
		return "informe o conteúdo no idioma padrão (" + blog.DefaultLocale + ")"
	}
	for loc := range req.Translations {
		if !knownLocales[loc] {
			return "idioma desconhecido: " + loc
		}
	}
	if len(req.Tags) > maxTags {
		return "muitas tags (máximo 10)"
	}
	for _, tag := range req.Tags {
		if !slugPattern.MatchString(tag.Slug) {
			return "tag inválida"
		}
		for loc, name := range tag.Names {
			if !knownLocales[loc] || strings.TrimSpace(name) == "" ||
				utf8.RuneCountInString(name) > maxTagRunes {
				return "tag inválida"
			}
		}
	}
	return ""
}

// CreatePost responde POST /posts, protegido por bearer token. Aceita as
// traduções por idioma; a tradução do idioma padrão é obrigatória.
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

		input := blog.PostInput{Slug: req.Slug, Translations: map[string]blog.Translation{}}
		if req.Published {
			now := time.Now().UTC()
			input.PublishedAt = &now
		}
		for loc, tr := range req.Translations {
			input.Translations[loc] = blog.Translation{
				Title:   strings.TrimSpace(tr.Title),
				Content: tr.Content,
			}
		}
		for _, tag := range req.Tags {
			names := make(map[string]string, len(tag.Names))
			for loc, name := range tag.Names {
				names[loc] = strings.TrimSpace(name)
			}
			input.Tags = append(input.Tags, blog.TagInput{Slug: tag.Slug, Names: names})
		}

		err := store.Create(r.Context(), input)
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

		def := input.Translations[blog.DefaultLocale]
		respTags := make([]tagJSON, 0, len(input.Tags))
		for _, tag := range input.Tags {
			respTags = append(respTags, tagJSON{Slug: tag.Slug, Name: tag.Names[blog.DefaultLocale]})
		}
		writeJSON(w, http.StatusCreated, postDetail{
			Slug:        input.Slug,
			Title:       def.Title,
			Content:     def.Content,
			PublishedAt: input.PublishedAt,
			Tags:        respTags,
		})
	}
}
