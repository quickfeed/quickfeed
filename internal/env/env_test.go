package env_test

import (
	"os"
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
	fi, err := os.Create(".env")
	if err != nil {
		t.Errorf("os.Create() = %v", err)
	}
	defer fi.Close()

	want := map[string]string{
		"QUICKFEED_TEST_ENV":  "test",
		"QUICKFEED_TEST_ENV2": "test2",
		"QUICKFEED_TEST_ENV3": "test3",
		"QUICKFEED_TEST_ENV4": "test4 xyz",
		"QUICKFEED_TEST_ENV5": "test5 = zyx",
	}

	input := `QUICKFEED_TEST_ENV=test
QUICKFEED_TEST_ENV2= test2

QUICKFEED_TEST_ENV3=test3
# Comment
QUICKFEED_TEST_ENV4=test4 xyz
## Another comment
QUICKFEED_TEST_ENV5=test5 = zyx
`
	if _, err = fi.WriteString(input); err != nil {
		t.Fatal(err)
	}

	if err = env.Load(""); err != nil {
		t.Fatal(err)
	}

	for k, v := range want {
		if got := os.Getenv(k); got != v {
			t.Errorf("os.Getenv(%q) = %q, wanted %q", k, got, v)
		}
	}

	if err = os.Remove(".env"); err != nil {
		t.Fatal(err)
	}
}
