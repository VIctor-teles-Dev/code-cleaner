// Package shortener define o domínio do encurtador: os tipos Link e ClickEvent,
// os erros de negócio, os contratos de persistência (Store, AnalyticsStore e
// ClickStore) e as validações puras compartilhadas entre handler e db.
package shortener

import (
	"context"
	"errors"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// Link é um link encurtado. ExpiresAt nulo significa que nunca expira.
type Link struct {
	Slug        string
	OriginalURL string
	CreatedAt   time.Time
	ExpiresAt   *time.Time
}

// ClickEvent é o evento bruto capturado no redirect. Country, Browser e Device
// são preenchidos pelo worker (assíncrono), fora do caminho do redirect.
type ClickEvent struct {
	Slug      string
	ClickedAt time.Time
	IP        string
	Country   string
	UserAgent string
	Browser   string
	Device    string
	Referrer  string
}

// Resolution é o que o cache guarda: destino + expiração + marcador de
// existência (para negative caching de 404).
type Resolution struct {
	OriginalURL string
	ExpiresAt   *time.Time
	Found       bool
}

// Expired informa se a resolução já passou da validade no instante t. A decisão
// 302/410 é sempre recalculada contra o relógio, nunca cacheada.
func (r Resolution) Expired(t time.Time) bool {
	return r.ExpiresAt != nil && !t.Before(*r.ExpiresAt)
}

// LabelCount é um par rótulo→contagem usado nos "top N" da analytics.
type LabelCount struct {
	Label string `json:"label"`
	Count int64  `json:"count"`
}

// DayCount é um ponto da série temporal diária.
type DayCount struct {
	Day   time.Time `json:"day"`
	Count int64     `json:"count"`
}

// Analytics agrega as métricas de clique de um slug.
type Analytics struct {
	Slug         string       `json:"slug"`
	TotalClicks  int64        `json:"total_clicks"`
	TimeSeries   []DayCount   `json:"time_series"`
	TopCountries []LabelCount `json:"top_countries"`
	TopReferrers []LabelCount `json:"top_referrers"`
	Browsers     []LabelCount `json:"browsers"`
	Devices      []LabelCount `json:"devices"`
}

var (
	// ErrNotFound: o slug não existe.
	ErrNotFound = errors.New("link not found")
	// ErrDuplicateSlug: o slug já está em uso (violação de unicidade).
	ErrDuplicateSlug = errors.New("slug already exists")
)

// Store persiste e resolve os links (escrita e caminho de resolução).
type Store interface {
	// Create insere o link. Retorna ErrDuplicateSlug em violação de unicidade
	// (Postgres 23505) — usado tanto para alias custom (409) quanto para o
	// retry do slug base62 gerado.
	Create(ctx context.Context, link Link) error
	// Resolve retorna o link pelo slug; ErrNotFound quando não existe. NÃO
	// filtra expiração no SQL — quem decide 410 vs 404 é o handler.
	Resolve(ctx context.Context, slug string) (Link, error)
}

// AnalyticsStore serve as leituras agregadas do endpoint de analytics.
type AnalyticsStore interface {
	Stats(ctx context.Context, slug string) (Analytics, error)
}

// ClickStore grava os eventos de clique em lote (usado pelo worker).
type ClickStore interface {
	InsertClicks(ctx context.Context, events []ClickEvent) error
}

const (
	// MaxURLLen limita o tamanho da URL de destino.
	MaxURLLen = 2048
	// AliasMaxLen limita o tamanho de um alias custom.
	AliasMaxLen = 40
)

var aliasPattern = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

// reservedAliases não podem virar slug custom: colidiriam com as rotas do
// serviço (a rota literal vence o wildcard, mas evita confusão e protege
// rotas futuras).
var reservedAliases = map[string]bool{
	"healthz":     true,
	"readyz":      true,
	"api":         true,
	"favicon.ico": true,
	"robots.txt":  true,
}

// ValidateURL valida a URL de destino. Retorna "" quando válida ou uma
// mensagem de erro (pt-BR) pronta para responder ao cliente.
func ValidateURL(raw string) string {
	raw = strings.TrimSpace(raw)
	switch {
	case raw == "":
		return "informe a URL de destino"
	case len(raw) > MaxURLLen:
		return "URL muito longa"
	}
	u, err := url.Parse(raw)
	if err != nil {
		return "URL inválida"
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "a URL deve começar com http:// ou https://"
	}
	if u.Host == "" {
		return "URL inválida"
	}
	return ""
}

// ValidateAlias valida um alias custom opcional. Alias vazio é válido (o slug
// será gerado). Retorna "" quando válido ou uma mensagem de erro (pt-BR).
func ValidateAlias(alias string) string {
	if alias == "" {
		return ""
	}
	switch {
	case len(alias) > AliasMaxLen:
		return "alias muito longo (máximo 40 caracteres)"
	case !aliasPattern.MatchString(alias):
		return "alias inválido (use letras, números, hífen e underscore)"
	case reservedAliases[strings.ToLower(alias)]:
		return "alias reservado"
	}
	return ""
}
