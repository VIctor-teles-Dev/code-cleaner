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

	tagsBySlug, err := s.postTags(ctx, true)
	if err != nil {
		return nil, err
	}
	for i := range posts {
		posts[i].Tags = tagsBySlug[posts[i].Slug]
	}
	return posts, nil
}

// ListAll retorna todos os posts (inclusive rascunhos), do mais recente ao
// mais antigo. Ordena por created_at porque rascunhos não têm published_at.
func (s PostStore) ListAll(ctx context.Context) ([]blog.Post, error) {
	rows, err := s.DB.QueryContext(ctx,
		`SELECT p.slug, p.title, p.content, p.published_at
		   FROM posts p
		  ORDER BY p.created_at DESC`)
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

	tagsBySlug, err := s.postTags(ctx, false)
	if err != nil {
		return nil, err
	}
	for i := range posts {
		posts[i].Tags = tagsBySlug[posts[i].Slug]
	}
	return posts, nil
}

// postTags carrega as tags dos posts de uma vez, evitando uma query por post
// na listagem. publishedOnly restringe aos posts publicados (listagem pública).
func (s PostStore) postTags(ctx context.Context, publishedOnly bool) (map[string][]blog.Tag, error) {
	query := `SELECT p.slug, t.slug, t.name
		   FROM posts p
		   JOIN post_tags pt ON pt.post_id = p.id
		   JOIN tags t ON t.id = pt.tag_id`
	if publishedOnly {
		query += ` WHERE ` + publishedFilter
	}
	query += ` ORDER BY t.name`

	rows, err := s.DB.QueryContext(ctx, query)
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

// GetBySlug retorna o post por slug independente de estar publicado (edição
// no admin). Reaproveita GetPublishedBySlug quando o post está publicado e,
// caso contrário, busca sem o filtro de publicação.
func (s PostStore) GetBySlug(ctx context.Context, slug string) (blog.Post, error) {
	var p blog.Post
	err := s.DB.QueryRowContext(ctx,
		`SELECT p.slug, p.title, p.content, p.published_at
		   FROM posts p
		  WHERE p.slug = $1`,
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

	if err := upsertPostTags(ctx, tx, postID, post.Tags); err != nil {
		return err
	}
	return tx.Commit()
}

// Update altera o post pelo slug (imutável) e substitui suas tags. A data de
// publicação é preservada quando o post continua publicado: passar
// PublishedAt não-nil publica mantendo a data original (COALESCE); nil
// despublica.
func (s PostStore) Update(ctx context.Context, post blog.Post) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	var postID int64
	err = tx.QueryRowContext(ctx,
		`UPDATE posts
		    SET title = $2,
		        content = $3,
		        published_at = CASE
		            WHEN $4::timestamptz IS NULL THEN NULL
		            ELSE COALESCE(published_at, $4)
		        END,
		        updated_at = now()
		  WHERE slug = $1
		  RETURNING id`,
		post.Slug, post.Title, post.Content, post.PublishedAt).Scan(&postID)
	if errors.Is(err, sql.ErrNoRows) {
		return blog.ErrNotFound
	}
	if err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx,
		`DELETE FROM post_tags WHERE post_id = $1`, postID); err != nil {
		return err
	}
	if err := upsertPostTags(ctx, tx, postID, post.Tags); err != nil {
		return err
	}
	return tx.Commit()
}

// Delete remove o post pelo slug. Os vínculos em post_tags somem via
// ON DELETE CASCADE. As tags em si ficam (podem ser usadas por outros posts).
func (s PostStore) Delete(ctx context.Context, slug string) error {
	res, err := s.DB.ExecContext(ctx, `DELETE FROM posts WHERE slug = $1`, slug)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return blog.ErrNotFound
	}
	return nil
}

// upsertPostTags faz upsert das tags pelo slug e vincula ao post, ignorando
// vínculos já existentes. Compartilhado por Create e Update.
func upsertPostTags(ctx context.Context, tx *sql.Tx, postID int64, tags []blog.Tag) error {
	for _, tag := range tags {
		var tagID int64
		if err := tx.QueryRowContext(ctx,
			`INSERT INTO tags (slug, name) VALUES ($1, $2)
			 ON CONFLICT (slug) DO UPDATE SET name = EXCLUDED.name
			 RETURNING id`,
			tag.Slug, tag.Name).Scan(&tagID); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO post_tags (post_id, tag_id) VALUES ($1, $2)
			 ON CONFLICT DO NOTHING`, postID, tagID); err != nil {
			return err
		}
	}
	return nil
}
