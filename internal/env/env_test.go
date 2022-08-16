package env_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
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
		"SOME_PATH":           "/quickfeed/root",
		"QUICKFEED_TEST_ENV":  "test",
		"QUICKFEED_TEST_ENV2": "test2",
		"QUICKFEED_TEST_ENV3": "test3",
		"QUICKFEED_TEST_ENV4": "test4 xyz",
		"QUICKFEED_TEST_ENV5": "test5 = zyx",
		"SOME_CERT_FILE":      "/quickfeed/root/cert/fullchain.pem",
		"SOME_KEY_FILE":       filepath.Join(os.Getenv("QUICKFEED"), "cert/fullchain.pem"),
		"WITHOUT_QUOTES":      filepath.Join(os.Getenv("QUICKFEED"), "cert/fullchain.pem"),
	}

	input := `QUICKFEED_TEST_ENV=test
QUICKFEED_TEST_ENV2= test2

QUICKFEED_TEST_ENV3=test3
# Comment
QUICKFEED_TEST_ENV4=test4 xyz
## Another comment
QUICKFEED_TEST_ENV5=test5 = zyx
# Variable to be expanded into other vars
SOME_PATH=/quickfeed/root
# Cert file and key file expanded
SOME_CERT_FILE=$SOME_PATH/cert/fullchain.pem
SOME_KEY_FILE=$QUICKFEED/cert/fullchain.pem
WITHOUT_QUOTES="$QUICKFEED/cert/fullchain.pem"
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

func TestWhitelist(t *testing.T) {
	test := []struct {
		domains string
		want    []string
		err     bool
	}{
		{"localhost", nil, true},
		{"localhost,example.com", nil, true},
		{"123.12.1.1", nil, true},
		{"example.com, www.example.com, localhost", nil, true},
		{"example.com, www.example.com,127.0.0.1:8080", nil, true},
		{"a.com, b.com, c.com", []string{"a.com", "b.com", "c.com"}, false},
		{"a.com,b.com,c.com", []string{"a.com", "b.com", "c.com"}, false},
		{"example.com, www.example.com", []string{"example.com", "www.example.com"}, false},
		//{"example.com, www.example.com,", []string{"example.com", "www.example.com"}, false},
	}

	for _, tc := range test {
		t.Setenv("QUICKFEED_WHITELIST", tc.domains)
		got, err := env.Whitelist()
		if err != nil && !tc.err {
			t.Errorf("Whitelist() = %v", err)
		}
		if diff := cmp.Diff(tc.want, got); diff != "" {
			t.Errorf("Whitelist() mismatch (-want +got):\n%s", diff)
		}
	}
}
