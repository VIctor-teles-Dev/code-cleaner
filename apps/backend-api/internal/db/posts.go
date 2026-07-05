package db

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/VIctor-teles-Dev/code-cleaner/apps/backend-api/internal/blog"
)

// PostStore implementa blog.Store sobre Postgres.
type PostStore struct {
	DB *sql.DB
}

const publishedFilter = "p.published_at IS NOT NULL AND p.published_at <= now()"

// LATERAL que escolhe a tradução do locale pedido ($1) com fallback para o
// default ($2). Reusado nas queries de post e de tag.
const postTranslation = `JOIN LATERAL (
	SELECT title, content FROM post_translations
	 WHERE post_id = p.id AND locale IN ($1, $2)
	 ORDER BY (locale = $1) DESC LIMIT 1
) tr ON true`

const tagTranslation = `JOIN LATERAL (
	SELECT name FROM tag_translations
	 WHERE tag_id = t.id AND locale IN ($1, $2)
	 ORDER BY (locale = $1) DESC LIMIT 1
) tt ON true`

func (s PostStore) ListPublished(ctx context.Context, locale, tagSlug string) ([]blog.Post, error) {
	query := `SELECT p.slug, tr.title, tr.content, p.published_at
	            FROM posts p ` + postTranslation + `
	           WHERE ` + publishedFilter
	args := []any{locale, blog.DefaultLocale}
	if tagSlug != "" {
		query += ` AND EXISTS (
		    SELECT 1 FROM post_tags pt JOIN tags t ON t.id = pt.tag_id
		     WHERE pt.post_id = p.id AND t.slug = $3)`
		args = append(args, tagSlug)
	}
	query += " ORDER BY p.published_at DESC"

	rows, err := s.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []blog.Post
	for rows.Next() {
		var p blog.Post
		if err := rows.Scan(&p.Slug, &p.Title, &p.Content, &p.PublishedAt); err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	tagsBySlug, err := s.publishedPostTags(ctx, locale)
	if err != nil {
		return nil, err
	}
	for i := range posts {
		posts[i].Tags = tagsBySlug[posts[i].Slug]
	}
	return posts, nil
}

// publishedPostTags carrega as tags (no locale) de todos os posts publicados
// de uma vez, evitando uma query por post na listagem.
func (s PostStore) publishedPostTags(ctx context.Context, locale string) (map[string][]blog.Tag, error) {
	rows, err := s.DB.QueryContext(ctx,
		`SELECT p.slug, t.slug, tt.name
		   FROM posts p
		   JOIN post_tags pt ON pt.post_id = p.id
		   JOIN tags t ON t.id = pt.tag_id `+tagTranslation+`
		  WHERE `+publishedFilter+`
		  ORDER BY tt.name`, locale, blog.DefaultLocale)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tags := make(map[string][]blog.Tag)
	for rows.Next() {
		var postSlug string
		var tag blog.Tag
		if err := rows.Scan(&postSlug, &tag.Slug, &tag.Name); err != nil {
			return nil, err
		}
		tags[postSlug] = append(tags[postSlug], tag)
	}
	return tags, rows.Err()
}

func (s PostStore) GetPublishedBySlug(ctx context.Context, locale, slug string) (blog.Post, error) {
	var p blog.Post
	err := s.DB.QueryRowContext(ctx,
		`SELECT p.slug, tr.title, tr.content, p.published_at
		   FROM posts p `+postTranslation+`
		  WHERE p.slug = $3 AND `+publishedFilter,
		locale, blog.DefaultLocale, slug).
		Scan(&p.Slug, &p.Title, &p.Content, &p.PublishedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return blog.Post{}, blog.ErrNotFound
	}
	if err != nil {
		return blog.Post{}, err
	}

	rows, err := s.DB.QueryContext(ctx,
		`SELECT t.slug, tt.name
		   FROM post_tags pt
		   JOIN tags t ON t.id = pt.tag_id
		   JOIN posts p ON p.id = pt.post_id `+tagTranslation+`
		  WHERE p.slug = $3
		  ORDER BY tt.name`, locale, blog.DefaultLocale, slug)
	if err != nil {
		return blog.Post{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var tag blog.Tag
		if err := rows.Scan(&tag.Slug, &tag.Name); err != nil {
			return blog.Post{}, err
		}
		p.Tags = append(p.Tags, tag)
	}
	return p, rows.Err()
}

func (s PostStore) Create(ctx context.Context, post blog.PostInput) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	var postID int64
	err = tx.QueryRowContext(ctx,
		`INSERT INTO posts (slug, published_at) VALUES ($1, $2) RETURNING id`,
		post.Slug, post.PublishedAt).Scan(&postID)

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return blog.ErrDuplicateSlug
	}
	if err != nil {
		return err
	}

	for locale, tr := range post.Translations {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO post_translations (post_id, locale, title, content)
			 VALUES ($1, $2, $3, $4)`,
			postID, locale, tr.Title, tr.Content); err != nil {
			return err
		}
	}

	for _, tag := range post.Tags {
		var tagID int64
		// upsert por slug; DO UPDATE no-op só para o RETURNING funcionar no conflito.
		err = tx.QueryRowContext(ctx,
			`INSERT INTO tags (slug) VALUES ($1)
			 ON CONFLICT (slug) DO UPDATE SET slug = EXCLUDED.slug
			 RETURNING id`, tag.Slug).Scan(&tagID)
		if err != nil {
			return err
		}
		for locale, name := range tag.Names {
			if _, err := tx.ExecContext(ctx,
				`INSERT INTO tag_translations (tag_id, locale, name)
				 VALUES ($1, $2, $3)
				 ON CONFLICT (tag_id, locale) DO UPDATE SET name = EXCLUDED.name`,
				tagID, locale, name); err != nil {
				return err
			}
		}
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO post_tags (post_id, tag_id) VALUES ($1, $2)
			 ON CONFLICT DO NOTHING`, postID, tagID); err != nil {
			return err
		}
	}

	return tx.Commit()
}
