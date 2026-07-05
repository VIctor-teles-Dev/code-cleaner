package db

import (
	"os"
	"testing"
)

func TestMigrateWithUnreachableDatabase(t *testing.T) {
	pool, err := Open("postgres://ccl:ccl@localhost:1/ccl?sslmode=disable")
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer pool.Close()

	if err := Migrate(pool); err == nil {
		t.Error("Migrate() = nil, want connection error")
	}
}

// Integração: roda quando TEST_DATABASE_URL aponta para um Postgres real
// (service container no CI, container efêmero no dev local).
func TestMigrateCreatesBlogSchema(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL não definido; pulando teste de integração")
	}

	pool, err := Open(url)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer pool.Close()

	if err := Migrate(pool); err != nil {
		t.Fatalf("Migrate() error = %v", err)
	}
	// Idempotente: rodar de novo (ex.: segunda réplica no deploy) não pode falhar.
	if err := Migrate(pool); err != nil {
		t.Fatalf("second Migrate() error = %v", err)
	}

	var count int
	err = pool.QueryRow(
		`SELECT count(*) FROM information_schema.tables WHERE table_name = 'posts'`,
	).Scan(&count)
	if err != nil {
		t.Fatalf("query error = %v", err)
	}
	if count != 1 {
		t.Fatalf("posts table count = %d, want 1", count)
	}

	if _, err := pool.Exec(`DELETE FROM posts`); err != nil {
		t.Fatalf("cleanup error = %v", err)
	}
	_, err = pool.Exec(
		`INSERT INTO posts (slug, title, content) VALUES ('hello-world', 'Hello World', 'primeiro post')`,
	)
	if err != nil {
		t.Fatalf("insert error = %v", err)
	}

	// slug é único — segunda inserção com o mesmo slug deve falhar
	_, err = pool.Exec(
		`INSERT INTO posts (slug, title, content) VALUES ('hello-world', 'Outro título', 'outro corpo')`,
	)
	if err == nil {
		t.Error("insert with duplicate slug = nil error, want unique violation")
	}
}
