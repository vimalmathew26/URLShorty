package id

import (
	"crypto/rand"
	"math/big"
	"strings"
)

const alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz" // 62 chars
var base = big.NewInt(int64(len(alphabet)))

// RandomBase62 returns a cryptographically-strong random base62 string of length n.
func RandomBase62(n int) (string, error) {
	if n <= 0 {
		n = 7
	}
	var b strings.Builder
	b.Grow(n)
	for i := 0; i < n; i++ {
		idx, err := rand.Int(rand.Reader, base) // uniform in [0,62)
		if err != nil {
			return "", err
		}
		b.WriteByte(alphabet[idx.Int64()])
	}
	return b.String(), nil
}

// Alphabet exposes the base62 alphabet (useful for tests, if any).
func Alphabet() string { return alphabet }
