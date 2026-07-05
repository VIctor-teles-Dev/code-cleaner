package db

import (
	"context"
	"database/sql"
	"os"
	"testing"
)

// testPool abre o pool de teste e aplica as migrations. Pula (skip) quando
// TEST_DATABASE_URL não está definido — mesmo padrão do backend-api.
func testPool(t *testing.T) *sql.DB {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	pool, err := Open(url)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = pool.Close() })
	if err := Migrate(pool); err != nil {
		t.Fatalf("Migrate: %v", err)
	}
	return pool
}

func TestMigrateIdempotent(t *testing.T) {
	pool := testPool(t)

	// Rodar de novo é no-op (advisory lock + ErrNoChange).
	if err := Migrate(pool); err != nil {
		t.Fatalf("Migrate (2ª vez): %v", err)
	}

	for _, table := range []string{"links", "clicks"} {
		var exists bool
		err := pool.QueryRowContext(context.Background(),
			`SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = $1)`,
			table).Scan(&exists)
		if err != nil {
			t.Fatalf("checando tabela %s: %v", table, err)
		}
		if !exists {
			t.Errorf("tabela %s não existe após Migrate", table)
		}
	}
}
