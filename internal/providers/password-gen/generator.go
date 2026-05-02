package password_gen

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

type Generator struct {
	size int
}

func New(size int) *Generator {
	if size <= 0 {
		size = 12
	}
	return &Generator{size: size}
}

func (g *Generator) Generate() (string, error) {
	max := big.NewInt(int64(len(alphabet)))
	out := make([]byte, g.size)
	for i := 0; i < g.size; i++ {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", fmt.Errorf("password generator: %w", err)
		}
		out[i] = alphabet[n.Int64()]
	}
	return string(out), nil
}
