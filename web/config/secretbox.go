package config

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"os"

	"golang.org/x/crypto/nacl/secretbox"
	"golang.org/x/term"
)

// ReadKey reads incrypted master key from a file and decrypts it
// with a passphrase. If fromEnv is true, reads passphrase from environment,
// otherwise asks for user input.
func (c *Config) ReadKey(fromEnv bool) error {
	var pass string
	if fromEnv {
		pass = os.Getenv("KEYPASS")
	}
	var inputBytes []byte
	if pass == "" {
		fmt.Println("Key: ")
		input, err := term.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			return err
		}
		inputBytes = input
	} else {
		inputBytes = []byte(pass)
	}

	passphrase, err := base64.RawStdEncoding.DecodeString(string(inputBytes))
	if err != nil {
		return err
	}
	if len(passphrase) != 32 {
		return fmt.Errorf("wrong key length, expected 32, got %d", len(passphrase))
	}
	path := os.Getenv(KeyEnv)
	if path == "" {
		return fmt.Errorf("missing file path")
	}
	keyfile, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	passkey := new([32]byte)
	nonce := new([24]byte)
	copy(passkey[:], passphrase[:32])
	copy(nonce[:], keyfile[:24])

	key, ok := secretbox.Open(nil, keyfile[24:], nonce, passkey)
	if !ok {
		return fmt.Errorf("error decrypting the key")
	}
	key32 := new([32]byte)
	copy(key32[:], key[:32])
	c.Secrets.key = key32
	return nil
}

func (c *Config) Cipher(token string) (string, error) {
	nonce := new([24]byte)
	_, err := io.ReadFull(rand.Reader, nonce[:])
	if err != nil {
		return "", fmt.Errorf("failed to geerate nonce: %w", err)
	}
	out := make([]byte, 24)
	copy(out, nonce[:])
	out = secretbox.Seal(out, []byte(token), nonce, c.Secrets.key)
	return base64.RawStdEncoding.EncodeToString(out), nil
}

func (c *Config) Decipher(tokenString string) (string, error) {
	tokenBytes, err := base64.RawStdEncoding.DecodeString(tokenString)
	if err != nil {
		return "", fmt.Errorf("failed to decode token string: %w", err)
	}
	nonce := new([24]byte)
	copy(nonce[:], tokenBytes[:24])
	token, ok := secretbox.Open(nil, tokenBytes[24:], nonce, c.Secrets.key)
	if !ok {
		return "", fmt.Errorf("failed to decrypt token")
	}
	return string(token), nil
}

// WithEncryption returns true if encryption key has been set.
func (c *Config) WithEncryption() bool {
	return c.Secrets.key != nil
}
