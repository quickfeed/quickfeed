package env_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/quickfeed/quickfeed/internal/env"
)

func TestScmProviderEnv(t *testing.T) {
	want := "github"
	got := env.ScmProvider()
	if got != want {
		t.Errorf("ScmProvider() = %s, wanted %s", got, want)
	}

	env.SetFakeProvider(t)
	want = "fake"
	got = env.ScmProvider()
	if got != want {
		t.Errorf("ScmProvider() = %s, wanted %s", got, want)
	}
}

func TestLoad(t *testing.T) {
	fi, err := os.CreateTemp("", ".env")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		fi.Close()
		if err = os.Remove(fi.Name()); err != nil {
			t.Fatal(err)
		}
	}()

	want := map[string]string{
		"QUICKFEED":           os.Getenv("QUICKFEED"),
		"QUICKFEED_PATH":      "/quickfeed/root",
		"QUICKFEED_TEST_ENV":  "test",
		"QUICKFEED_TEST_ENV2": "test2",
		"QUICKFEED_TEST_ENV3": "test3",
		"QUICKFEED_TEST_ENV4": "test4 xyz",
		"QUICKFEED_TEST_ENV5": "test5 = zyx",
		"QUICKFEED_CERT_FILE": "/quickfeed/root/cert/fullchain.pem",
		"QUICKFEED_KEY_FILE":  filepath.Join(os.Getenv("QUICKFEED"), "cert/fullchain.pem"),
	}

	input := `QUICKFEED_TEST_ENV=test
QUICKFEED_TEST_ENV2= test2

QUICKFEED_TEST_ENV3=test3
# Comment
QUICKFEED_TEST_ENV4=test4 xyz
## Another comment
QUICKFEED_TEST_ENV5=test5 = zyx
# Variable to be expanded into other vars
QUICKFEED_PATH=/quickfeed/root
# Cert file and key file expanded
QUICKFEED_CERT_FILE=$QUICKFEED_PATH/cert/fullchain.pem
QUICKFEED_KEY_FILE=$QUICKFEED/cert/fullchain.pem
`
	if _, err = fi.WriteString(input); err != nil {
		t.Fatal(err)
	}

	if err = env.Load(fi.Name()); err != nil {
		t.Fatal(err)
	}

	for k, v := range want {
		if got := os.Getenv(k); got != v {
			t.Errorf("os.Getenv(%q) = %q, wanted %q", k, got, v)
		}
	}
}
