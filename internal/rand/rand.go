package rand

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
)

func String() string {
	randomness := make([]byte, 10)
	if _, err := rand.Read(randomness); err != nil {
		panic("couldn't generate randomness")
	}
	return fmt.Sprintf("%x", sha256.Sum256(randomness))
}
