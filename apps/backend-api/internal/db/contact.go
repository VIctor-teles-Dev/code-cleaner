package db

import (
	"context"
	"database/sql"
)

// ContactStore persists contact-form messages in Postgres.
type ContactStore struct {
	DB *sql.DB
}

func (s ContactStore) SaveContactMessage(ctx context.Context, name, email, message string) error {
	_, err := s.DB.ExecContext(ctx,
		`INSERT INTO contact_messages (name, email, message) VALUES ($1, $2, $3)`,
		name, email, message)
	return err
}
