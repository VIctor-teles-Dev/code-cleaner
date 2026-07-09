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

type Tag struct {
	Slug string
	Name string
}

type Post struct {
	Slug        string
	Title       string
	Content     string
	PublishedAt *time.Time
	Tags        []Tag
}

var (
	ErrNotFound      = errors.New("post not found")
	ErrDuplicateSlug = errors.New("slug already exists")
)

type Store interface {
	// ListPublished retorna os posts publicados, do mais recente ao mais
	// antigo. tagSlug vazio lista todos; preenchido, filtra pela tag.
	ListPublished(ctx context.Context, tagSlug string) ([]Post, error)
	// ListAll retorna todos os posts (inclusive rascunhos), do mais recente
	// ao mais antigo. Uso exclusivo do admin.
	ListAll(ctx context.Context) ([]Post, error)
	// GetPublishedBySlug retorna ErrNotFound quando o post não existe
	// ou ainda não foi publicado.
	GetPublishedBySlug(ctx context.Context, slug string) (Post, error)
	// GetBySlug retorna o post por slug independente de publicação
	// (para edição no admin). ErrNotFound quando não existe.
	GetBySlug(ctx context.Context, slug string) (Post, error)
	// Create persiste o post e faz upsert das tags pelo slug.
	// Retorna ErrDuplicateSlug quando o slug do post já está em uso.
	Create(ctx context.Context, post Post) error
	// Update altera título, conteúdo, publicação e tags do post identificado
	// por post.Slug (o slug em si é imutável). PublishedAt nil despublica;
	// não-nil publica preservando a data original se já era publicado.
	// Retorna ErrNotFound quando o slug não existe.
	Update(ctx context.Context, post Post) error
	// Delete remove o post (e seus vínculos de tag) pelo slug.
	// Retorna ErrNotFound quando o slug não existe.
	Delete(ctx context.Context, slug string) error
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
