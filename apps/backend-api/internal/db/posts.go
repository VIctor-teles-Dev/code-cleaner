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

func (s PostStore) ListPublished(ctx context.Context, tagSlug string) ([]blog.Post, error) {
	query := `SELECT p.slug, p.title, p.content, p.published_at
	            FROM posts p
	           WHERE ` + publishedFilter
	var args []any
	if tagSlug != "" {
		query += ` AND EXISTS (
		    SELECT 1 FROM post_tags pt JOIN tags t ON t.id = pt.tag_id
		     WHERE pt.post_id = p.id AND t.slug = $1)`
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

	tagsBySlug, err := s.publishedPostTags(ctx)
	if err != nil {
		return nil, err
	}
	for i := range posts {
		posts[i].Tags = tagsBySlug[posts[i].Slug]
	}
	return posts, nil
}

// publishedPostTags carrega as tags de todos os posts publicados de uma vez,
// evitando uma query por post na listagem.
func (s PostStore) publishedPostTags(ctx context.Context) (map[string][]blog.Tag, error) {
	rows, err := s.DB.QueryContext(ctx,
		`SELECT p.slug, t.slug, t.name
		   FROM posts p
		   JOIN post_tags pt ON pt.post_id = p.id
		   JOIN tags t ON t.id = pt.tag_id
		  WHERE `+publishedFilter+`
		  ORDER BY t.name`)
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

func (s PostStore) GetPublishedBySlug(ctx context.Context, slug string) (blog.Post, error) {
	var p blog.Post
	err := s.DB.QueryRowContext(ctx,
		`SELECT p.slug, p.title, p.content, p.published_at
		   FROM posts p
		  WHERE p.slug = $1 AND `+publishedFilter,
		slug).Scan(&p.Slug, &p.Title, &p.Content, &p.PublishedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return blog.Post{}, blog.ErrNotFound
	}
	if err != nil {
		return blog.Post{}, err
	}

	rows, err := s.DB.QueryContext(ctx,
		`SELECT t.slug, t.name
		   FROM post_tags pt
		   JOIN tags t ON t.id = pt.tag_id
		   JOIN posts p ON p.id = pt.post_id
		  WHERE p.slug = $1
		  ORDER BY t.name`, slug)
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

func (s PostStore) Create(ctx context.Context, post blog.Post) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	var postID int64
	err = tx.QueryRowContext(ctx,
		`INSERT INTO posts (slug, title, content, published_at)
		 VALUES ($1, $2, $3, $4) RETURNING id`,
		post.Slug, post.Title, post.Content, post.PublishedAt).Scan(&postID)

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return blog.ErrDuplicateSlug
	}
	if err != nil {
		return err
	}

	for _, tag := range post.Tags {
		var tagID int64
		err = tx.QueryRowContext(ctx,
			`INSERT INTO tags (slug, name) VALUES ($1, $2)
			 ON CONFLICT (slug) DO UPDATE SET name = EXCLUDED.name
			 RETURNING id`,
			tag.Slug, tag.Name).Scan(&tagID)
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO post_tags (post_id, tag_id) VALUES ($1, $2)
			 ON CONFLICT DO NOTHING`, postID, tagID); err != nil {
			return err
		}
	}

	return tx.Commit()
}
