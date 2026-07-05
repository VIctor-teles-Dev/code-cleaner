package shortener

import (
	"crypto/rand"
	"math/big"
)

const (
	// alphabet base62: 62^7 ≈ 3,5 trilhões de slugs possíveis.
	alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	slugLen  = 7
)

// GenerateSlug gera um slug base62 aleatório de 7 caracteres com crypto/rand.
// Slugs imprevisíveis importam porque a analytics é exposta por slug — com
// math/rand eles seriam enumeráveis.
func GenerateSlug() (string, error) {
	max := big.NewInt(int64(len(alphabet)))
	b := make([]byte, slugLen)
	for i := range b {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		b[i] = alphabet[n.Int64()]
	}
	return string(b), nil
}
