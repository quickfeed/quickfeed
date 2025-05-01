package manifest

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/google/go-github/v62/github"
	"github.com/quickfeed/quickfeed/internal/env"
	"github.com/quickfeed/quickfeed/internal/ui"
	"github.com/quickfeed/quickfeed/web"
	"github.com/quickfeed/quickfeed/web/auth"
)

const (
	appID         = "QUICKFEED_APP_ID"
	appKey        = "QUICKFEED_APP_KEY"
	appUrl        = "QUICKFEED_APP_URL"
	clientID      = "QUICKFEED_CLIENT_ID"
	clientSecret  = "QUICKFEED_CLIENT_SECRET"  // skipcq: SCT-A000
	webhookSecret = "QUICKFEED_WEBHOOK_SECRET" // skipcq: SCT-A000
)

// ReadyForAppCreation returns nil if the environment configuration (envFile)
// is ready for creating a new GitHub App. Otherwise, it returns an error,
// e.g., if the envFile already contains App information or if the .env is
// missing and there is a corresponding .env.bak file. The optional chkFn
// functions are called to perform additional checks.
func ReadyForAppCreation(envFile string, chkFns ...func() error) error {
	if env.HasAppID() {
		return fmt.Errorf("%s already contains App information", envFile)
	}
	// Check for missing .env file and if .env.bak already exists
	for _, envFile := range []string{env.RootEnv(envFile), env.PublicEnv(envFile)} {
		if err := env.Prepared(envFile); err != nil {
			return err
		}
	}
	for _, checker := range chkFns {
		if err := checker(); err != nil {
			return err
		}
	}
	return nil
}

func CreateNewQuickFeedApp(srvFn web.ServerType, envFile string, dev bool) error {
	m := New(env.DomainWithPort(), envFile, dev)
	server, err := srvFn(m.Handler())
	if err != nil {
		return err
	}
	return m.StartAppCreationFlow(server)
}

type Manifest struct {
	domain  string
	envFile string
	handler http.Handler
	done    chan error
	client  *github.Client // optional, for testing
	build   func() error
}

func New(domain, envFile string, dev bool) *Manifest {
	m := &Manifest{
		domain:  domain,
		envFile: envFile,
		client:  github.NewClient(nil),
		done:    make(chan error),
		build:   func() error { return ui.Build("", dev) },
	}
	router := http.NewServeMux()
	router.Handle("/manifest/callback", m.conversion())
	router.Handle("/manifest", m.createApp())
	m.handler = router
	return m
}

func (m *Manifest) Handler() http.Handler {
	return m.handler
}

func (m *Manifest) StartAppCreationFlow(server *web.Server) error {
	go func() {
		if err := server.Serve(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				m.done <- fmt.Errorf("could not start web server for GitHub App creation flow: %w", err)
				return
			}
			// server was closed prematurely, e.g., ctrl-C
			m.done <- fmt.Errorf("server was closed prematurely")
		}
	}()
	log.Println("Important: The GitHub user that installs the QuickFeed App will become the server's admin user.")
	log.Printf("Go to https://%s/manifest to install the QuickFeed GitHub App.\n", env.DomainWithPort())
	if err := <-m.done; err != nil {
		return err
	}
	return server.Shutdown(context.Background())
}

func (m *Manifest) conversion() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var retErr error
		code := r.URL.Query().Get("code")
		defer func() {
			m.done <- retErr
		}()
		if code == "" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "No code provided")
			retErr = errors.New("no code provided")
			return
		}
		ctx := context.Background()
		config, resp, err := m.client.Apps.CompleteAppManifest(ctx, code)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "Error: %s", err)
			retErr = err
			return
		}
		// GitHub returns 201 Created on success
		if resp.StatusCode != http.StatusCreated {
			w.WriteHeader(resp.StatusCode)
			fmt.Fprintf(w, "Error: %s", resp.Status)
			retErr = fmt.Errorf("unexpected response code: %s", resp.Status)
			return
		}

		// Create directories on path to PEM file, if not exists
		appKeyFile := env.AppKey()
		if err := os.MkdirAll(filepath.Dir(appKeyFile), 0o700); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Error: %s", err)
			retErr = err
			return
		}

		// Write PEM file
		if err := os.WriteFile(appKeyFile, []byte(config.GetPEM()), 0o600); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Error: %s", err)
			retErr = err
			return
		}

		// Save the application configuration to the envFile
		envToUpdate := map[string]string{
			appID:         strconv.FormatInt(config.GetID(), 10),
			appKey:        appKeyFile,
			appUrl:        config.GetHTMLURL(),
			clientID:      config.GetClientID(),
			clientSecret:  config.GetClientSecret(),
			webhookSecret: config.GetWebhookSecret(),
		}

		if err := env.Save(env.RootEnv(m.envFile), envToUpdate); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Error: %s", err)
			retErr = err
			return
		}

		// Print success message, and redirect to main page
		if err := m.success(w, config); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Error: %s", err)
			retErr = err
			return
		}

		// Build the UI if needed
		if err := m.buildUI(); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Error: %s", err)
			retErr = err
		} else {
			log.Printf("UI built successfully")
		}
	}
}

func (m *Manifest) createApp() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		if err := form(w, m.domain); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Error: %s", err)
			// only signal done on error
			m.done <- err
		}
	}
}

func (m *Manifest) success(w http.ResponseWriter, config *github.AppConfig) error {
	const tpl = `<!DOCTYPE html>
<html>
<head>
<style>
body {
  background: #aaa;
}

.container {
  height: 300px;
}

.center {
  position: absolute;
  font-family: verdana;
  color: #40a;
  top: 50%;
  left: 50%;
  -ms-transform: translate(-50%, -50%);
  transform: translate(-50%, -50%);
}
</style>
</head>
<body>
  <div class="container">
    <div class="center">
      <h2>{{.}} GitHub App created</h2>
	  <h3>Running esbuild in the background</h3>
	  <h3>Building the UI, please wait for "UI built successfully" in server logs before logging in<h3>
	  <h4>Reloading soon...</h4>
    </div>
  </div>
</body>
</html>

<script>
	setTimeout(function() {
		window.location.href = "/";
	}, 500);
</script>
`
	log.Printf("Successfully created the %s GitHub App.", config.GetName())
	t := template.Must(template.New("success").Parse(tpl))
	return t.Execute(w, config.GetName())
}

// buildUI builds the UI. If it fails, it runs npm ci and tries again.
// This is useful when the UI may not be built yet.
// The build function can be overridden for testing purposes.
func (m *Manifest) buildUI() error {
	if err := m.build(); err != nil {
		if ok := runNpmCi(); !ok {
			return fmt.Errorf("failed to run npm ci: %w", err)
		}
		// Attempt to build again
		if err := m.build(); err != nil {
			return fmt.Errorf("failed to rebuild the UI: %w", err)
		}
	}
	return nil
}

func runNpmCi() bool {
	log.Println("Running npm ci...")
	c := exec.Command("npm", "ci")
	c.Dir = env.PublicDir()
	if err := c.Run(); err != nil {
		log.Print(c.Output())
		log.Print(err)
		log.Print("Failed to run npm ci; giving up")
		return false
	}
	log.Print("Done npm ci")
	return true
}

func form(w http.ResponseWriter, domain string) error {
	const tpl = `
	<html>
		<form id="create" action="https://github.com/settings/apps/new" method="post">
			<input type="hidden" name="manifest" id="manifest"><br>
			<input type="hidden" value="Submit">
	   	</form>
	</html>

	<script>
		input = document.getElementById("manifest")
		input.value = JSON.stringify({
			"name": "{{.Name}}",
			"url": "{{.URL}}",
{{- if .WebhookActive}}
			"hook_attributes": {
				"active": true,
				"url": "{{.WebhookURL}}",
			},
{{- end}}
			"callback_urls": [
				"{{.CallbackURL}}"
			],
			"redirect_url": "{{.URL}}/manifest/callback",
			"public": true,
			"default_permissions": {
				"administration": "write",
				"contents": "write",
				"issues": "write",
				"members": "write",
				"organization_administration": "write",
				"pull_requests": "write",
			},
{{- if .WebhookActive}}
			"default_events": [
				"push",
				"pull_request",
				"pull_request_review"
			]
{{- end}}
		})
		document.getElementById('create').submit()
	</script>
	`

	data := struct {
		URL           string
		Name          string
		CallbackURL   string
		WebhookURL    string
		WebhookActive bool
	}{
		URL:           auth.GetBaseURL(domain),
		Name:          env.AppName(),
		CallbackURL:   auth.GetCallbackURL(domain),
		WebhookURL:    auth.GetEventsURL(domain),
		WebhookActive: true,
	}

	if env.IsLocal(domain) {
		// Disable webhook for localhost, or any other non-public domain
		data.WebhookActive = false
	}
	t := template.Must(template.New("form").Parse(tpl))
	if err := t.Execute(w, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}
	return nil
}
