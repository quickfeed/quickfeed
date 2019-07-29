package auth_test

import (
	"os"
	"reflect"
	"testing"

	"github.com/autograde/aguis/web/auth"
	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/github"
)

const (
	key         = "KEY"
	secret      = "SECRET"
	callbackURL = "http://localhost/callback"
)

func TestEnableProvider(t *testing.T) {
	const (
		name      = "test"
		keyEnv    = "TEST_KEY"
		secretEnv = "TEST_SECRET"
	)

	oldProviderSet := goth.GetProviders()
	goth.ClearProviders()
	defer func() {
		goth.ClearProviders()
		for _, provider := range oldProviderSet {
			goth.UseProviders(provider)
		}
		if err := os.Unsetenv(keyEnv); err != nil {
			t.Fatal(err)
		}
		if err := os.Unsetenv(secretEnv); err != nil {
			t.Fatal(err)
		}
	}()

	if err := os.Setenv(keyEnv, key); err != nil {
		t.Fatal(err)
	}
	if err := os.Setenv(secretEnv, secret); err != nil {
		t.Fatal(err)
	}

	auth.EnableProvider(&auth.Provider{
		Name:        name,
		KeyEnv:      keyEnv,
		SecretEnv:   secretEnv,
		CallbackURL: callbackURL,
	}, func(key, secret, callback string, scopes ...string) goth.Provider {
		return github.New(key, secret, callback, scopes...)
	})

	if len(goth.GetProviders()) != 2 {
		t.Fatal("expected 2 providers to be enabled")
	}

	checkProvider(t, name)
	checkProvider(t, name+auth.TeacherSuffix)
}

func checkProvider(t *testing.T, name string) {
	provider, err := goth.GetProvider(name)
	if err != nil {
		t.Fatal(err)
	}

	githubProvider, ok := provider.(*github.Provider)
	if !ok {
		var want *github.Provider
		t.Fatalf("have provider type %v want %v", reflect.TypeOf(provider), reflect.TypeOf(want))
	}

	have := &github.Provider{
		ClientKey:   githubProvider.ClientKey,
		Secret:      githubProvider.Secret,
		CallbackURL: githubProvider.CallbackURL,
	}

	want := &github.Provider{
		ClientKey:   key,
		Secret:      secret,
		CallbackURL: callbackURL,
	}

	if !reflect.DeepEqual(have, want) {
		t.Errorf("have course %+v want %+v", have, want)
	}
}
