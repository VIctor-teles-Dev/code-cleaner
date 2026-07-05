// Package blog define o domínio do blog: os tipos Post e Tag, os erros de
// negócio e o contrato de persistência compartilhado entre handler e db.
package blog

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"
)

// DefaultLocale é o idioma de fallback quando uma tradução não existe.
const DefaultLocale = "pt-BR"

// Tag é uma tag já resolvida para um idioma (Name no locale pedido).
type Tag struct {
	Slug string
	Name string
}

// Post é um post já resolvido para um idioma (Title/Content no locale pedido).
type Post struct {
	Slug        string
	Title       string
	Content     string
	PublishedAt *time.Time
	Tags        []Tag
}

// Translation é o título/conteúdo de um post num idioma.
type Translation struct {
	Title   string
	Content string
}

// TagInput é uma tag na criação: slug estável + nome por idioma.
type TagInput struct {
	Slug  string
	Names map[string]string // locale -> nome
}

// PostInput é o payload de criação de um post multilíngue.
type PostInput struct {
	Slug         string
	PublishedAt  *time.Time
	Translations map[string]Translation // locale -> {title, content}
	Tags         []TagInput
}

var (
	ErrNotFound      = errors.New("post not found")
	ErrDuplicateSlug = errors.New("slug already exists")
)

type Store interface {
	// ListPublished retorna os posts publicados no locale pedido (com fallback
	// para DefaultLocale), do mais recente ao mais antigo. tagSlug vazio lista
	// todos; preenchido, filtra pela tag.
	ListPublished(ctx context.Context, locale, tagSlug string) ([]Post, error)
	// GetPublishedBySlug retorna o post no locale pedido (com fallback);
	// ErrNotFound quando o post não existe ou ainda não foi publicado.
	GetPublishedBySlug(ctx context.Context, locale, slug string) (Post, error)
	// Create persiste o post com suas traduções e faz upsert das tags (por
	// slug + nome por idioma). Retorna ErrDuplicateSlug se o slug já existe.
	Create(ctx context.Context, post PostInput) error
}

var (
	accentReplacer = strings.NewReplacer(
		"á", "a", "à", "a", "â", "a", "ã", "a", "ä", "a",
		"é", "e", "è", "e", "ê", "e", "ë", "e",
		"í", "i", "ì", "i", "î", "i", "ï", "i",
		"ó", "o", "ò", "o", "ô", "o", "õ", "o", "ö", "o",
		"ú", "u", "ù", "u", "û", "u", "ü", "u",
		"ç", "c", "ñ", "n",
	)
	nonSlugChars = regexp.MustCompile(`[^a-z0-9]+`)
)

// Slugify normaliza um nome (pt-BR) para slug: minúsculas, sem acentos,
// hífens no lugar de qualquer outra coisa.
func Slugify(name string) string {
	s := accentReplacer.Replace(strings.ToLower(strings.TrimSpace(name)))
	s = nonSlugChars.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}
