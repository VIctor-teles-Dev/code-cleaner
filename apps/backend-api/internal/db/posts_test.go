package db

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/VIctor-teles-Dev/code-cleaner/apps/backend-api/internal/blog"
)

// Teste de integração: roda apenas com TEST_DATABASE_URL (mesmo padrão dos demais).
func TestPostStore(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}

	pool, err := Open(url)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer pool.Close()

	if err := Migrate(pool); err != nil {
		t.Fatalf("Migrate: %v", err)
	}

	ctx := context.Background()
	// CASCADE limpa também post_translations, tag_translations e post_tags.
	if _, err := pool.ExecContext(ctx, "TRUNCATE posts, tags CASCADE"); err != nil {
		t.Fatalf("TRUNCATE: %v", err)
	}

	store := PostStore{DB: pool}
	now := time.Now().UTC()

	arquitetura := blog.TagInput{
		Slug:  "arquitetura",
		Names: map[string]string{"pt-BR": "Arquitetura", "en": "Architecture"},
	}
	published := blog.PostInput{
		Slug:        "publicado",
		PublishedAt: &now,
		Translations: map[string]blog.Translation{
			"pt-BR": {Title: "Publicado", Content: "corpo pt"},
			"en":    {Title: "Published", Content: "body en"},
		},
		Tags: []blog.TagInput{arquitetura},
	}
	draft := blog.PostInput{
		Slug:         "rascunho",
		Translations: map[string]blog.Translation{"pt-BR": {Title: "Rascunho", Content: "wip"}},
	}

	for _, p := range []blog.PostInput{published, draft} {
		if err := store.Create(ctx, p); err != nil {
			t.Fatalf("Create(%s): %v", p.Slug, err)
		}
	}
	if err := store.Create(ctx, published); !errors.Is(err, blog.ErrDuplicateSlug) {
		t.Errorf("Create duplicado = %v, want ErrDuplicateSlug", err)
	}

	// Locale pt-BR: só o publicado, com título e tag em português.
	ptPosts, err := store.ListPublished(ctx, "pt-BR", "")
	if err != nil {
		t.Fatalf("ListPublished pt: %v", err)
	}
	if len(ptPosts) != 1 || ptPosts[0].Slug != "publicado" || ptPosts[0].Title != "Publicado" {
		t.Fatalf("ListPublished pt = %+v, want só o publicado com título pt", ptPosts)
	}
	if len(ptPosts[0].Tags) != 1 || ptPosts[0].Tags[0].Name != "Arquitetura" {
		t.Errorf("tags pt = %+v, want Arquitetura", ptPosts[0].Tags)
	}

	// Locale en: título e tag em inglês.
	enPosts, err := store.ListPublished(ctx, "en", "")
	if err != nil {
		t.Fatalf("ListPublished en: %v", err)
	}
	if len(enPosts) != 1 || enPosts[0].Title != "Published" {
		t.Fatalf("ListPublished en = %+v, want título en", enPosts)
	}
	if len(enPosts[0].Tags) != 1 || enPosts[0].Tags[0].Name != "Architecture" {
		t.Errorf("tags en = %+v, want Architecture", enPosts[0].Tags)
	}

	// Filtro por tag (slug é locale-agnóstico).
	filtered, err := store.ListPublished(ctx, "en", "arquitetura")
	if err != nil || len(filtered) != 1 {
		t.Errorf("ListPublished(tag) = %+v, %v; want o post da tag", filtered, err)
	}
	none, err := store.ListPublished(ctx, "en", "sem-posts")
	if err != nil || len(none) != 0 {
		t.Errorf("ListPublished(tag inexistente) = %+v, %v; want vazio", none, err)
	}

	// Detalhe em en; rascunho não publicado retorna ErrNotFound.
	gotEn, err := store.GetPublishedBySlug(ctx, "en", "publicado")
	if err != nil {
		t.Fatalf("GetPublishedBySlug en: %v", err)
	}
	if gotEn.Title != "Published" || gotEn.Content != "body en" {
		t.Errorf("get en = %+v, want conteúdo en", gotEn)
	}
	if _, err := store.GetPublishedBySlug(ctx, "pt-BR", "rascunho"); !errors.Is(err, blog.ErrNotFound) {
		t.Errorf("GetPublishedBySlug(rascunho) = %v, want ErrNotFound (draft)", err)
	}

	// Fallback: post só com pt-BR, pedido em en, cai para o pt-BR.
	onlyPt := blog.PostInput{
		Slug:         "so-pt",
		PublishedAt:  &now,
		Translations: map[string]blog.Translation{"pt-BR": {Title: "Só PT", Content: "só pt"}},
	}
	if err := store.Create(ctx, onlyPt); err != nil {
		t.Fatalf("Create(so-pt): %v", err)
	}
	fb, err := store.GetPublishedBySlug(ctx, "en", "so-pt")
	if err != nil {
		t.Fatalf("get so-pt en: %v", err)
	}
	if fb.Title != "Só PT" {
		t.Errorf("fallback = %q, want o título pt-BR 'Só PT'", fb.Title)
	}

	// Upsert de tag: reusar a mesma tag não duplica.
	second := blog.PostInput{
		Slug:         "segundo",
		PublishedAt:  &now,
		Translations: map[string]blog.Translation{"pt-BR": {Title: "Segundo", Content: "c"}},
		Tags:         []blog.TagInput{arquitetura},
	}
	if err := store.Create(ctx, second); err != nil {
		t.Fatalf("Create(segundo): %v", err)
	}
	var tagCount int
	if err := pool.QueryRowContext(ctx,
		`SELECT count(*) FROM tags WHERE slug = 'arquitetura'`).Scan(&tagCount); err != nil {
		t.Fatalf("count tags: %v", err)
	}
	if tagCount != 1 {
		t.Errorf("tags com mesmo slug = %d, want 1 (upsert)", tagCount)
	}
}
