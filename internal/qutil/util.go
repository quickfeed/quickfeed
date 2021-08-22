package qutil

import (
	"crypto/rand"
	"crypto/sha1"
	"fmt"
)

func RandomString() string {
	randomness := make([]byte, 10)
	if _, err := rand.Read(randomness); err != nil {
		panic("couldn't generate randomness")
	}
	return fmt.Sprintf("%x", sha1.Sum(randomness))
}
