package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type fakeStore struct {
	err   error
	saved []string
}

func (f *fakeStore) SaveContactMessage(_ context.Context, name, email, message string) error {
	f.saved = append(f.saved, name+"|"+email+"|"+message)
	return f.err
}

type fakeMailer struct {
	err  error
	sent int
}

func (f *fakeMailer) SendContactNotification(_, _, _ string) error {
	f.sent++
	return f.err
}

func postContact(handler http.HandlerFunc, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/contact", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler(rec, req)
	return rec
}

const validBody = `{"name":"Ana","email":"ana@example.com","message":"Olá, Victor!"}`

func TestContactStoresMessageAndNotifies(t *testing.T) {
	store := &fakeStore{}
	mailer := &fakeMailer{}

	rec := postContact(Contact(store, mailer), validBody)

	if rec.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusCreated)
	}
	if len(store.saved) != 1 || store.saved[0] != "Ana|ana@example.com|Olá, Victor!" {
		t.Errorf("saved = %v, want the submitted message", store.saved)
	}
	if mailer.sent != 1 {
		t.Errorf("mailer.sent = %d, want 1", mailer.sent)
	}
	if body := rec.Body.String(); !strings.Contains(body, `"status":"received"`) {
		t.Errorf("body = %q, want it to contain %q", body, `"status":"received"`)
	}
}

func TestContactSucceedsWhenMailerFails(t *testing.T) {
	store := &fakeStore{}
	mailer := &fakeMailer{err: errors.New("smtp down")}

	rec := postContact(Contact(store, mailer), validBody)

	if rec.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d (message was stored)", rec.Code, http.StatusCreated)
	}
}

func TestContactWithoutMailerOnlyStores(t *testing.T) {
	store := &fakeStore{}

	rec := postContact(Contact(store, nil), validBody)

	if rec.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusCreated)
	}
	if len(store.saved) != 1 {
		t.Errorf("saved = %v, want 1 message", store.saved)
	}
}

func TestContactValidation(t *testing.T) {
	cases := []struct {
		label string
		body  string
	}{
		{"missing name", `{"email":"a@b.com","message":"oi"}`},
		{"missing email", `{"name":"Ana","message":"oi"}`},
		{"invalid email", `{"name":"Ana","email":"not-an-email","message":"oi"}`},
		{"missing message", `{"name":"Ana","email":"a@b.com"}`},
		{"blank message", `{"name":"Ana","email":"a@b.com","message":"   "}`},
		{"malformed json", `{`},
	}

	for _, tc := range cases {
		t.Run(tc.label, func(t *testing.T) {
			store := &fakeStore{}
			rec := postContact(Contact(store, nil), tc.body)

			if rec.Code != http.StatusBadRequest {
				t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
			}
			if len(store.saved) != 0 {
				t.Errorf("saved = %v, want nothing stored", store.saved)
			}
		})
	}
}

func TestContactFailsWhenStoreFails(t *testing.T) {
	store := &fakeStore{err: errors.New("db down")}

	rec := postContact(Contact(store, nil), validBody)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}

func TestContactWithoutStoreIsUnavailable(t *testing.T) {
	rec := postContact(Contact(nil, nil), validBody)

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusServiceUnavailable)
	}
}
