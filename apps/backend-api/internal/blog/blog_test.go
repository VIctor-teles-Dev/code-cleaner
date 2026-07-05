package blog

import "testing"

func TestSlugify(t *testing.T) {
	cases := []struct{ in, want string }{
		{"minhas aplicações", "minhas-aplicacoes"},
		{"Go & Kubernetes!", "go-kubernetes"},
		{"  Café com Código  ", "cafe-com-codigo"},
		{"---já-um-slug---", "ja-um-slug"},
	}
	for _, tc := range cases {
		if got := Slugify(tc.in); got != tc.want {
			t.Errorf("Slugify(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
