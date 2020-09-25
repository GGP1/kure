package passgen

import (
	"crypto/rand"
	"math/big"
)

// Generator is the interface that wraps secrets functions.
type Generator interface {
	Generate() (string, error)
	Entropy() float64
}

// randInt returns a cryptographically secure random integer in [0, max).
//
// Error handling is skipped as we are not using zero or negative numbers.
func randInt(max int) int {
	randN, _ := rand.Int(rand.Reader, big.NewInt(int64(max)))

	return int(randN.Int64())
}
