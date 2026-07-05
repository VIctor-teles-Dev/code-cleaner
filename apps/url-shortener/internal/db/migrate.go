package db

import (
	"database/sql"
	"embed"
	"errors"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// migrationsTable isola o histórico de migrations deste serviço. O backend-api
// compartilha o mesmo banco Postgres e usa a tabela padrão (schema_migrations);
// se usássemos a mesma tabela, um serviço veria a versão do outro e pularia as
// próprias migrations.
const migrationsTable = "shortener_schema_migrations"

// Migrate applies pending migrations embedded in the binary. Safe to run from
// every replica at startup: the postgres driver serializes via advisory lock,
// and an already-migrated database is a no-op.
func Migrate(pool *sql.DB) error {
	src, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return err
	}

	driver, err := postgres.WithInstance(pool, &postgres.Config{
		MigrationsTable: migrationsTable,
	})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithInstance("iofs", src, "postgres", driver)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}
