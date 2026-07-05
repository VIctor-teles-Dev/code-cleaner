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
	if _, err := pool.ExecContext(ctx, "TRUNCATE posts, tags CASCADE"); err != nil {
		t.Fatalf("TRUNCATE: %v", err)
	}

	store := PostStore{DB: pool}
	now := time.Now().UTC()

	tag := blog.Tag{Slug: "minhas-aplicacoes", Name: "minhas aplicações"}
	published := blog.Post{Slug: "publicado", Title: "Publicado", Content: "corpo",
		PublishedAt: &now, Tags: []blog.Tag{tag}}
	draft := blog.Post{Slug: "rascunho", Title: "Rascunho", Content: "wip"}

	for _, p := range []blog.Post{published, draft} {
		if err := store.Create(ctx, p); err != nil {
			t.Fatalf("Create(%s): %v", p.Slug, err)
		}
	}

	if err := store.Create(ctx, published); !errors.Is(err, blog.ErrDuplicateSlug) {
		t.Errorf("Create duplicado = %v, want ErrDuplicateSlug", err)
	}

	posts, err := store.ListPublished(ctx, "")
	if err != nil {
		t.Fatalf("ListPublished: %v", err)
	}
	if len(posts) != 1 || posts[0].Slug != "publicado" {
		t.Fatalf("ListPublished = %+v, want only the published post", posts)
	}
	if len(posts[0].Tags) != 1 || posts[0].Tags[0] != tag {
		t.Errorf("ListPublished tags = %+v, want %+v", posts[0].Tags, tag)
	}

	filtered, err := store.ListPublished(ctx, "minhas-aplicacoes")
	if err != nil || len(filtered) != 1 {
		t.Errorf("ListPublished(tag) = %+v, %v; want the tagged post", filtered, err)
	}
	none, err := store.ListPublished(ctx, "tag-sem-posts")
	if err != nil || len(none) != 0 {
		t.Errorf("ListPublished(tag inexistente) = %+v, %v; want empty", none, err)
	}

	got, err := store.GetPublishedBySlug(ctx, "publicado")
	if err != nil {
		t.Errorf("GetPublishedBySlug(publicado) = %v, want nil", err)
	}
	if len(got.Tags) != 1 || got.Tags[0] != tag {
		t.Errorf("GetPublishedBySlug tags = %+v, want %+v", got.Tags, tag)
	}
	if _, err := store.GetPublishedBySlug(ctx, "rascunho"); !errors.Is(err, blog.ErrNotFound) {
		t.Errorf("GetPublishedBySlug(rascunho) = %v, want ErrNotFound (draft)", err)
	}

	// Reuso de tag existente: outro post com a mesma tag não duplica
	second := blog.Post{Slug: "segundo", Title: "Segundo", Content: "c",
		PublishedAt: &now, Tags: []blog.Tag{tag}}
	if err := store.Create(ctx, second); err != nil {
		t.Fatalf("Create(segundo): %v", err)
	}
	var tagCount int
	if err := pool.QueryRowContext(ctx,
		`SELECT count(*) FROM tags WHERE slug = 'minhas-aplicacoes'`).Scan(&tagCount); err != nil {
		t.Fatalf("count tags: %v", err)
	}
	if tagCount != 1 {
		t.Errorf("tags com mesmo slug = %d, want 1 (upsert)", tagCount)
	}
}
