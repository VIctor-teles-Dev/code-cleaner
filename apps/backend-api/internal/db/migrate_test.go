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

	// Schema multilíngue: posts + tabelas de tradução existem.
	for _, table := range []string{"posts", "post_translations", "tags", "tag_translations"} {
		var count int
		if err := pool.QueryRow(
			`SELECT count(*) FROM information_schema.tables WHERE table_name = $1`, table,
		).Scan(&count); err != nil {
			t.Fatalf("query %s error = %v", table, err)
		}
		if count != 1 {
			t.Fatalf("%s table count = %d, want 1", table, count)
		}
	}

	if _, err := pool.Exec(`TRUNCATE posts CASCADE`); err != nil {
		t.Fatalf("cleanup error = %v", err)
	}
	var postID int64
	if err := pool.QueryRow(
		`INSERT INTO posts (slug) VALUES ('hello-world') RETURNING id`,
	).Scan(&postID); err != nil {
		t.Fatalf("insert post error = %v", err)
	}
	if _, err := pool.Exec(
		`INSERT INTO post_translations (post_id, locale, title, content)
		 VALUES ($1, 'pt-BR', 'Hello World', 'primeiro post')`, postID,
	); err != nil {
		t.Fatalf("insert translation error = %v", err)
	}

	// slug é único — segunda inserção com o mesmo slug deve falhar
	if _, err := pool.Exec(`INSERT INTO posts (slug) VALUES ('hello-world')`); err == nil {
		t.Error("insert with duplicate slug = nil error, want unique violation")
	}
}
