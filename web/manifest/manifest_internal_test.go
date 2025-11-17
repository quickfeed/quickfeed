package manifest

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-github/v62/github"
	"github.com/quickfeed/quickfeed/internal/env"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/internal/ui"
	"github.com/quickfeed/quickfeed/scm"
)

func TestForm(t *testing.T) {
	tests := []struct {
		name       string
		domain     string
		status     int
		hasWebhook bool
	}{
		{
			name:       "no_webhook",
			domain:     "localhost",
			status:     http.StatusOK,
			hasWebhook: false,
		},
		{
			name:       "webhook",
			domain:     "example.com",
			status:     http.StatusOK,
			hasWebhook: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			t.Setenv("DOMAIN", tt.domain)
			if err := form(rr); err != nil {
				t.Fatalf("form() failed with error: %v", err)
			}
			if status := rr.Code; status != tt.status {
				t.Fatalf("form() returned wrong status code: got %v want %v", status, tt.status)
			}
			body := rr.Body.String()
			if tt.hasWebhook {
				if !strings.Contains(body, "default_events") {
					t.Errorf("form() returned body without default_events")
				}
				if !strings.Contains(body, "hook_attributes") {
					t.Errorf("form() returned body without hook_attributes")
				}
				if !strings.Contains(body, `"active": true`) {
					t.Errorf("form() returned body without active webhook")
				}
			} else {
				if strings.Contains(body, "default_events") {
					t.Errorf("form() returned body with default_events")
				}
				if strings.Contains(body, "hook_attributes") {
					t.Errorf("form() returned body with hook_attributes")
				}
				if strings.Contains(body, `"active": true`) {
					t.Errorf("form() returned body with active webhook")
				}
			}
			if t.Failed() {
				t.Log(body)
			}
		})
	}
}

func TestConversion(t *testing.T) {
	pemPath := env.Root("testdata", "private-key.pem")
	t.Setenv("QUICKFEED_APP_KEY", pemPath)
	tmpEnvFile := qtest.CreateTempFile(t, env.Root())
	t.Cleanup(func() {
		// Remove also the backup file if it exists
		bakFileName := tmpEnvFile + ".bak"
		if _, err := os.Stat(bakFileName); err == nil {
			if err := os.Remove(bakFileName); err != nil {
				t.Error(err)
			}
		}
	})

	tests := []struct {
		name string
		// code is used to simulate the received code from the GitHub callback
		// and is used to fetch the corresponding app config from the mock SCM client.
		code string
		want map[string]string
		fail bool
	}{
		{
			name: "empty config",
			code: "1",
			want: map[string]string{
				appID:         "0",
				clientID:      "",
				clientSecret:  "",
				webhookSecret: "",
			},
		},
		{
			name: "full config",
			code: "2",
			want: map[string]string{
				appID:         "1",
				clientID:      "client",
				clientSecret:  "secret",
				webhookSecret: "webhook-secret",
			},
		},
		{
			name: "full config",
			code: "3",
			want: map[string]string{
				appID:         "123",
				clientID:      "some-id",
				clientSecret:  "some-other-secret",
				webhookSecret: "some-webhook-secret",
			},
		},
		{
			name: "invalid code",
			code: "",
			want: map[string]string{},
			fail: true,
		},
		{
			name: "status not created",
			code: "4000",
			want: map[string]string{},
			fail: true,
		},
	}

	config := map[string]github.AppConfig{
		"1": {},
		"2": {
			Name:          qtest.Ptr("test"),
			ID:            qtest.Ptr(int64(1)),
			ClientID:      qtest.Ptr("client"),
			ClientSecret:  qtest.Ptr("secret"),
			HTMLURL:       qtest.Ptr("https://example.com"),
			PEM:           qtest.Ptr("secret"),
			WebhookSecret: qtest.Ptr("webhook-secret"),
		},
		"3": {
			Name:          qtest.Ptr("test"),
			ID:            qtest.Ptr(int64(123)),
			ClientID:      qtest.Ptr("some-id"),
			ClientSecret:  qtest.Ptr("some-other-secret"),
			HTMLURL:       qtest.Ptr("https://another-example.com"),
			PEM:           qtest.Ptr("super-secret"),
			WebhookSecret: qtest.Ptr("some-webhook-secret"),
		},
		// TODO: Test with webhook config (manifest with non-private address)
	}

	scmClient := scm.NewMockedGithubSCMClient(qtest.Logger(t), scm.WithMockAppConfig(config))
	t.Setenv("DOMAIN", "localhost")
	manifest := Manifest{
		client:  scmClient.Client(),
		envFile: filepath.Base(tmpEnvFile),
		done:    make(chan error, 1),
		build:   func() error { return nil }, // Avoid building UI when testing
	}

	mux := http.NewServeMux()
	mux.Handle("/manifest/callback", manifest.conversion())
	server := httptest.NewServer(mux)
	defer server.Close()

	for _, tt := range tests {
		// Send a POST request to our conversion handler
		// This will simulate the callback from GitHub
		// with the code from the test case.
		url := fmt.Sprintf("%s/manifest/callback?code=%s", server.URL, tt.code)
		_, err := server.Client().Post(url, "application/json", nil)
		if err != nil {
			t.Fatalf("failed to send request: %v", err)
		}

		// Wait for the conversion flow to finish
		err = <-manifest.done
		if err != nil && !tt.fail {
			t.Errorf("unexpected error in done channel: %v", err)
		}
		if err == nil && tt.fail {
			t.Error("expected error in done channel")
		}

		// In some cases we expect the conversion flow to fail,
		// such as when the code is invalid or the status is not "created",
		// so we skip the environment variable checks
		if tt.fail {
			continue
		}

		for k := range tt.want {
			// Unset all relevant environment variables
			// to prevent interference between tests
			os.Unsetenv(k)
		}

		// Load the environment variables from the updated .env file
		// after the conversion flow.
		// This is done by the StartAppCreationFlow function in production,
		// but for testing purposes we need to do it manually.
		if err := env.Load(tmpEnvFile); err != nil {
			t.Fatalf("failed to load .env file: %v", err)
		}
		for k, v := range tt.want {
			// We expect the environment variables to be correctly set
			// after the conversion flow
			if got := os.Getenv(k); got != v {
				t.Errorf("os.Getenv(%q) = %q, wanted %q", k, got, v)
			}
		}

		pem, err := os.ReadFile(pemPath)
		if err != nil {
			t.Fatalf("failed to read pem file: %v", err)
		}

		cfg, ok := config[tt.code]
		if !ok {
			t.Fatalf("config for code %q not found", tt.code)
		}
		if diff := cmp.Diff(cfg.GetPEM(), string(pem)); diff != "" {
			t.Errorf("pem file content mismatch (-want +got):\n%s", diff)
		}
	}
}

func TestBuildUI(t *testing.T) {
	if os.Getenv("CI") == "true" {
		t.Skipf("Skipping %s when running on GitHub", t.Name())
	}
	m := Manifest{
		build: func() error { return ui.Build(t.TempDir(), true) },
	}
	if err := m.buildUI(); err != nil {
		t.Errorf("buildUI() failed with error: %v", err)
	}
}
