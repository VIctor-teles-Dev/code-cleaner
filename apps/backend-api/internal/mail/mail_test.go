package mail

import (
	"strings"
	"testing"
)

func TestFromEnvDisabledWithoutConfig(t *testing.T) {
	t.Setenv("SMTP_HOST", "")
	t.Setenv("CONTACT_TO", "")

	if m := FromEnv(); m != nil {
		t.Errorf("FromEnv() = %+v, want nil when SMTP is not configured", m)
	}
}

func TestFromEnvDefaults(t *testing.T) {
	t.Setenv("SMTP_HOST", "smtp.example.com")
	t.Setenv("SMTP_PORT", "")
	t.Setenv("SMTP_USERNAME", "bot@example.com")
	t.Setenv("SMTP_PASSWORD", "secret")
	t.Setenv("CONTACT_TO", "victor@example.com")
	t.Setenv("CONTACT_FROM", "")

	m := FromEnv()
	if m == nil {
		t.Fatal("FromEnv() = nil, want a configured mailer")
	}
	if m.Port != "587" {
		t.Errorf("Port = %q, want default %q", m.Port, "587")
	}
	if m.From != "bot@example.com" {
		t.Errorf("From = %q, want fallback to SMTP_USERNAME", m.From)
	}
}

func TestBuildContactMessage(t *testing.T) {
	msg := string(BuildContactMessage(
		"bot@example.com", "victor@example.com",
		"Ana", "ana@example.com", "Quero um site!"))

	for _, want := range []string{
		"From: bot@example.com\r\n",
		"To: victor@example.com\r\n",
		"Reply-To: ana@example.com\r\n",
		"Subject: [code-cleaner] Novo contato de Ana\r\n",
		"Quero um site!",
	} {
		if !strings.Contains(msg, want) {
			t.Errorf("message missing %q\nmessage:\n%s", want, msg)
		}
	}
}

func TestBuildContactMessageBlocksHeaderInjection(t *testing.T) {
	msg := string(BuildContactMessage(
		"bot@example.com", "victor@example.com",
		"Ana\r\nBcc: spam@evil.com", "ana@example.com", "oi"))

	if strings.Contains(msg, "\r\nBcc: spam@evil.com") {
		t.Error("injected header line survived sanitization")
	}
	if !strings.Contains(msg, "Subject: [code-cleaner] Novo contato de Ana  Bcc: spam@evil.com\r\n") {
		t.Error("newlines in the name should be flattened into the subject")
	}
}
