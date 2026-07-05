package shortener

import (
	"strings"
	"testing"
)

func TestGenerateSlug(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 1000; i++ {
		slug, err := GenerateSlug()
		if err != nil {
			t.Fatalf("GenerateSlug: %v", err)
		}
		if len(slug) != slugLen {
			t.Fatalf("len(%q) = %d, want %d", slug, len(slug), slugLen)
		}
		for _, c := range slug {
			if !strings.ContainsRune(alphabet, c) {
				t.Fatalf("slug %q contém caractere fora do alfabeto: %q", slug, c)
			}
		}
		seen[slug] = true
	}
	// Com 62^7 de espaço, 1000 slugs devem ser praticamente todos distintos.
	if len(seen) < 990 {
		t.Errorf("gerou muitos slugs repetidos: %d distintos de 1000", len(seen))
	}
}
