package db

import (
	"context"
	"os"
	"testing"
)

// Teste de integração: roda apenas com TEST_DATABASE_URL apontando para um
// Postgres real (mesmo padrão de migrate_test.go).
func TestSaveContactMessage(t *testing.T) {
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
	store := ContactStore{DB: pool}

	if err := store.SaveContactMessage(ctx, "Ana", "ana@example.com", "Olá!"); err != nil {
		t.Fatalf("SaveContactMessage: %v", err)
	}

	var got string
	err = pool.QueryRowContext(ctx,
		`SELECT message FROM contact_messages WHERE email = 'ana@example.com'
		 ORDER BY id DESC LIMIT 1`).Scan(&got)
	if err != nil {
		t.Fatalf("SELECT: %v", err)
	}
	if got != "Olá!" {
		t.Errorf("message = %q, want %q", got, "Olá!")
	}
}
