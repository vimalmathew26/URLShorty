package id

import (
	"context"

	"urlshorty/internal/core"
)

// Generator implements core.CodeGenerator using base62 random strings.
type Generator struct {
	length int
}

// NewGenerator creates a code generator with a fixed length (default 7 if <=0).
func NewGenerator(length int) *Generator {
	if length <= 0 {
		length = 7
	}
	return &Generator{length: length}
}

// NewCode generates a new random code.
func (g *Generator) NewCode(_ context.Context) (string, error) {
	return RandomBase62(g.length)
}

// Ensure *Generator satisfies the interface at compile-time.
var _ core.CodeGenerator = (*Generator)(nil)
