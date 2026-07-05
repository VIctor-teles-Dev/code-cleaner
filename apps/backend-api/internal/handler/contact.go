package handler

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	netmail "net/mail"
	"strings"
	"time"
	"unicode/utf8"
)

// ContactStore persists contact messages, satisfied by db.ContactStore.
type ContactStore interface {
	SaveContactMessage(ctx context.Context, name, email, message string) error
}

// Mailer sends the notification email for a new contact message.
// A nil mailer means notifications are disabled (message is only stored).
type Mailer interface {
	SendContactNotification(name, email, message string) error
}

type contactRequest struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Message string `json:"message"`
}

type contactResponse struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

const (
	maxNameLen    = 200
	maxMessageLen = 5000
	maxBodyBytes  = 16 << 10
)

func writeContactJSON(w http.ResponseWriter, status int, body contactResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func validateContact(req contactRequest) string {
	name := strings.TrimSpace(req.Name)
	message := strings.TrimSpace(req.Message)

	switch {
	case name == "":
		return "informe seu nome"
	case utf8.RuneCountInString(name) > maxNameLen:
		return "nome muito longo"
	case req.Email == "":
		return "informe seu email"
	case message == "":
		return "escreva uma mensagem"
	case utf8.RuneCountInString(message) > maxMessageLen:
		return "mensagem muito longa"
	}

	if _, err := netmail.ParseAddress(req.Email); err != nil {
		return "email inválido"
	}
	return ""
}

// Contact receives a contact-form message, stores it and, when a mailer is
// configured, sends a best-effort email notification. Storage is the source
// of truth: a mail failure does not fail the request.
func Contact(store ContactStore, mailer Mailer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if store == nil {
			writeContactJSON(w, http.StatusServiceUnavailable,
				contactResponse{Status: "unavailable"})
			return
		}

		var req contactRequest
		r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeContactJSON(w, http.StatusBadRequest,
				contactResponse{Status: "invalid", Error: "corpo da requisição inválido"})
			return
		}

		if msg := validateContact(req); msg != "" {
			writeContactJSON(w, http.StatusBadRequest,
				contactResponse{Status: "invalid", Error: msg})
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		name := strings.TrimSpace(req.Name)
		message := strings.TrimSpace(req.Message)
		if err := store.SaveContactMessage(ctx, name, req.Email, message); err != nil {
			log.Printf("contact: save failed: %v", err)
			writeContactJSON(w, http.StatusInternalServerError,
				contactResponse{Status: "error"})
			return
		}

		if mailer != nil {
			if err := mailer.SendContactNotification(name, req.Email, message); err != nil {
				log.Printf("contact: notification email failed (message stored): %v", err)
			}
		}

		writeContactJSON(w, http.StatusCreated, contactResponse{Status: "received"})
	}
}
