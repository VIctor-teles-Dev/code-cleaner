package db

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/VIctor-teles-Dev/code-cleaner/apps/url-shortener/internal/shortener"
)

func TestLinkStore(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	if _, err := pool.ExecContext(ctx, "TRUNCATE clicks, links RESTART IDENTITY CASCADE"); err != nil {
		t.Fatalf("TRUNCATE: %v", err)
	}
	store := LinkStore{DB: pool}

	link := shortener.Link{Slug: "abc123", OriginalURL: "https://go.dev"}
	if err := store.Create(ctx, link); err != nil {
		t.Fatalf("Create: %v", err)
	}

	// Slug duplicado -> ErrDuplicateSlug (via 23505).
	if err := store.Create(ctx, link); !errors.Is(err, shortener.ErrDuplicateSlug) {
		t.Errorf("Create duplicado = %v, want ErrDuplicateSlug", err)
	}

	got, err := store.Resolve(ctx, "abc123")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if got.OriginalURL != "https://go.dev" {
		t.Errorf("OriginalURL = %q, want https://go.dev", got.OriginalURL)
	}

	if _, err := store.Resolve(ctx, "nao-existe"); !errors.Is(err, shortener.ErrNotFound) {
		t.Errorf("Resolve inexistente = %v, want ErrNotFound", err)
	}

	// Link expirado é retornado bruto (Resolve não filtra expiração no SQL).
	past := time.Now().Add(-time.Hour).UTC()
	if err := store.Create(ctx, shortener.Link{Slug: "old", OriginalURL: "https://x.com", ExpiresAt: &past}); err != nil {
		t.Fatalf("Create expirado: %v", err)
	}
	expired, err := store.Resolve(ctx, "old")
	if err != nil {
		t.Fatalf("Resolve expirado: %v", err)
	}
	if expired.ExpiresAt == nil {
		t.Error("Resolve deveria retornar expires_at do link expirado")
	}
}
