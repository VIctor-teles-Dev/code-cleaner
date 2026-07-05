package db

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/VIctor-teles-Dev/code-cleaner/apps/url-shortener/internal/shortener"
)

// LinkStore implementa shortener.Store sobre Postgres.
type LinkStore struct {
	DB *sql.DB
}

func (s LinkStore) Create(ctx context.Context, link shortener.Link) error {
	_, err := s.DB.ExecContext(ctx,
		`INSERT INTO links (slug, original_url, expires_at) VALUES ($1, $2, $3)`,
		link.Slug, link.OriginalURL, link.ExpiresAt)

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return shortener.ErrDuplicateSlug
	}
	return err
}

func (s LinkStore) Resolve(ctx context.Context, slug string) (shortener.Link, error) {
	link := shortener.Link{Slug: slug}
	err := s.DB.QueryRowContext(ctx,
		`SELECT original_url, created_at, expires_at FROM links WHERE slug = $1`,
		slug).Scan(&link.OriginalURL, &link.CreatedAt, &link.ExpiresAt)
	if errors.Is(err, sql.ErrNoRows) {
		return shortener.Link{}, shortener.ErrNotFound
	}
	if err != nil {
		return shortener.Link{}, err
	}
	return link, nil
}
