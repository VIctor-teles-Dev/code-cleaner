// Package mail envia notificações por SMTP (relay externo via STARTTLS,
// porta 587). O envio parte da própria VPS; apenas a entrega final passa
// pelo relay, o que evita os bloqueios de porta 25 e as blocklists que
// afetam servidores de email próprios.
package mail

import (
	"fmt"
	"net/smtp"
	"os"
	"strings"
)

// SMTP is a Mailer backed by an external SMTP relay.
type SMTP struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
	To       string
}

// FromEnv builds an SMTP mailer from SMTP_* / CONTACT_* env vars.
// Returns nil (notifications disabled) unless SMTP_HOST and CONTACT_TO are set.
func FromEnv() *SMTP {
	host := os.Getenv("SMTP_HOST")
	to := os.Getenv("CONTACT_TO")
	if host == "" || to == "" {
		return nil
	}

	port := os.Getenv("SMTP_PORT")
	if port == "" {
		port = "587"
	}
	username := os.Getenv("SMTP_USERNAME")
	from := os.Getenv("CONTACT_FROM")
	if from == "" {
		from = username
	}

	return &SMTP{
		Host:     host,
		Port:     port,
		Username: username,
		Password: os.Getenv("SMTP_PASSWORD"),
		From:     from,
		To:       to,
	}
}

// sanitizeHeader impede header injection em campos vindos do formulário.
func sanitizeHeader(s string) string {
	return strings.NewReplacer("\r", " ", "\n", " ").Replace(s)
}

// BuildContactMessage monta a mensagem RFC 5322 da notificação. O Reply-To
// aponta para o visitante, então responder o email responde a pessoa.
func BuildContactMessage(from, to, name, email, message string) []byte {
	var b strings.Builder
	fmt.Fprintf(&b, "From: %s\r\n", sanitizeHeader(from))
	fmt.Fprintf(&b, "To: %s\r\n", sanitizeHeader(to))
	fmt.Fprintf(&b, "Reply-To: %s\r\n", sanitizeHeader(email))
	fmt.Fprintf(&b, "Subject: [code-cleaner] Novo contato de %s\r\n", sanitizeHeader(name))
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString("Content-Type: text/plain; charset=utf-8\r\n")
	b.WriteString("\r\n")
	fmt.Fprintf(&b, "Nome: %s\nEmail: %s\n\n%s\n", sanitizeHeader(name), sanitizeHeader(email), message)
	return []byte(b.String())
}

func (s *SMTP) SendContactNotification(name, email, message string) error {
	msg := BuildContactMessage(s.From, s.To, name, email, message)

	var auth smtp.Auth
	if s.Username != "" {
		auth = smtp.PlainAuth("", s.Username, s.Password, s.Host)
	}

	return smtp.SendMail(s.Host+":"+s.Port, auth, s.From, []string{s.To}, msg)
}
