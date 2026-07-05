// Package geoip resolve IP -> país via base MaxMind GeoLite2. FromEnv espelha
// mail.FromEnv do backend-api: retorna nil quando não configurado, deixando o
// serviço rodar sem GeoIP (o país fica vazio).
package geoip

import (
	"log"
	"net"
	"os"

	"github.com/oschwald/geoip2-golang"
)

// Resolver resolve o país de um IP. Um *Resolver nil é um resolver
// desabilitado válido (todos os métodos tratam o receiver nil).
type Resolver struct {
	reader *geoip2.Reader
}

// FromEnv abre a base apontada por GEOIP_DB_PATH. Retorna nil (GeoIP
// desabilitado) quando a variável está vazia ou o arquivo não abre.
func FromEnv() *Resolver {
	path := os.Getenv("GEOIP_DB_PATH")
	if path == "" {
		return nil
	}
	reader, err := geoip2.Open(path)
	if err != nil {
		log.Printf("geoip: não foi possível abrir %q: %v (país desabilitado)", path, err)
		return nil
	}
	return &Resolver{reader: reader}
}

// Country resolve o código ISO do país (ex.: "BR"). Retorna "" para resolver
// nil, IP inválido ou país desconhecido.
func (r *Resolver) Country(ip string) string {
	if r == nil || r.reader == nil {
		return ""
	}
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return ""
	}
	record, err := r.reader.Country(parsed)
	if err != nil {
		return ""
	}
	return record.Country.IsoCode
}

// Close libera a base. Seguro em resolver nil.
func (r *Resolver) Close() error {
	if r == nil || r.reader == nil {
		return nil
	}
	return r.reader.Close()
}
