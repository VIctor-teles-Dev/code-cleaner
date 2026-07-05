package shortener

import (
	"strings"
	"testing"
)

func TestValidateURL(t *testing.T) {
	cases := []struct {
		in    string
		valid bool
	}{
		{"https://go.dev", true},
		{"http://example.com/path?q=1", true},
		{"  https://trim.me  ", true},
		{"", false},
		{"ftp://example.com", false},
		{"javascript:alert(1)", false},
		{"https://", false},
		{strings.Repeat("a", MaxURLLen+1), false},
	}
	for _, tc := range cases {
		got := ValidateURL(tc.in) == ""
		if got != tc.valid {
			t.Errorf("ValidateURL(%q) válido = %v, want %v (msg=%q)",
				tc.in, got, tc.valid, ValidateURL(tc.in))
		}
	}
}

func TestValidateAlias(t *testing.T) {
	cases := []struct {
		in    string
		valid bool
	}{
		{"", true},
		{"minha-marca", true},
		{"Abc_123", true},
		{"com espaço", false},
		{"acento-ç", false},
		{"api", false},     // reservado
		{"HEALTHZ", false}, // reservado, case-insensitive
		{strings.Repeat("a", AliasMaxLen+1), false},
	}
	for _, tc := range cases {
		got := ValidateAlias(tc.in) == ""
		if got != tc.valid {
			t.Errorf("ValidateAlias(%q) válido = %v, want %v (msg=%q)",
				tc.in, got, tc.valid, ValidateAlias(tc.in))
		}
	}
}
