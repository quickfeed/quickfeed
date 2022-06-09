package config_test

import (
	"os"
	"testing"

	"github.com/autograde/quickfeed/web/config"
)

func TestSecretbox(t *testing.T) {
	config := config.NewConfig("localhost", "public", "8080")
	// Needs a KEYFILE, KEYPASS environmental variables
	path := os.Getenv("KEYFILE")
	pass := os.Getenv("KEYPASS")
	if path == "" || pass == "" {
		t.Skip()
	}
	if err := config.ReadKey(true); err != nil {
		t.Fatal(err)
	}
	token := "TestAccessTokenWithSomeTextToMatchThelen"
	encrypted, err := config.Cipher(token)
	if err != nil {
		t.Fatal(err)
	}
	decrypted, err := config.Decipher(encrypted)
	if err != nil {
		t.Fatal(err)
	}
	if decrypted != token {
		t.Errorf("output mismatch: expected (%s), got (%s)", token, decrypted)
	}
}
