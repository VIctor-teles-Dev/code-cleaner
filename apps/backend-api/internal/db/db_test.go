package db

import "testing"

func TestOpenRejectsInvalidURL(t *testing.T) {
	if _, err := Open("://not-a-url"); err == nil {
		t.Error("Open() = nil error, want parse error")
	}
}

func TestOpenAcceptsValidURL(t *testing.T) {
	// Open does not connect; the connection is validated lazily (readyz/queries).
	pool, err := Open("postgres://wbc:wbc@localhost:5432/wbc?sslmode=disable")
	if err != nil {
		t.Fatalf("Open() error = %v, want nil", err)
	}
	defer pool.Close()

	if pool == nil {
		t.Error("Open() = nil pool, want non-nil")
	}
}
