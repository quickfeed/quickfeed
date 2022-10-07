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

	"github.com/google/go-github/v45/github"
	"github.com/quickfeed/quickfeed/internal/env"
	"github.com/quickfeed/quickfeed/web"
	"github.com/quickfeed/quickfeed/web/auth"
)

const (
	appID        = "QUICKFEED_APP_ID"
	appKey       = "QUICKFEED_APP_KEY"
	clientID     = "QUICKFEED_CLIENT_ID"
	clientSecret = "QUICKFEED_CLIENT_SECRET"
)

type Manifest struct {
	domain  string
	handler http.Handler
	done    chan error
}

func New(domain string) *Manifest {
	m := &Manifest{
		domain: domain,
		done:   make(chan error),
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
				m.done <- fmt.Errorf("could not start web server for GitHub App creation flow: %v", err)
				return
			}
			// server was closed prematurely, e.g., ctrl-C
			m.done <- fmt.Errorf("server was closed prematurely")
		}
	}()
	log.Println("Important: The GitHub user that installs the QuickFeed App will become the server's admin user.")
	log.Printf("Go to https://%s/manifest to install the QuickFeed GitHub App.\n", env.Domain())
	if err := <-m.done; err != nil {
		return err
	}
	if err := server.Shutdown(context.Background()); err != nil {
		return err
	}
	// Refresh environment variables
	return env.Load("")
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
		config, resp, err := github.NewClient(nil).Apps.CompleteAppManifest(ctx, code)
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
		if err := os.WriteFile(appKeyFile, []byte(*config.PEM), 0o600); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Error: %s", err)
			retErr = err
			return
		}

		// Save the application configuration to the .env file
		envToUpdate := map[string]string{
			appID:        strconv.FormatInt(*config.ID, 10),
			appKey:       appKeyFile,
			clientID:     *config.ClientID,
			clientSecret: *config.ClientSecret,
		}
		if err := env.Save(".env", envToUpdate); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Error: %s", err)
			retErr = err
			return
		}

		// Print success message, and redirect to main page
		if err := success(w, config); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Error: %s", err)
			retErr = err
		}
	}
}

func (m *Manifest) createApp() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if err := form(w, m.domain); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Error: %s", err)
			// only signal done on error
			m.done <- err
		}
	}
}

func success(w http.ResponseWriter, config *github.AppConfig) error {
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
      <h2>{{.Name}} GitHub App created</h2>
	  <h3>Running webpack in the background</h3>
	  <h3>Please wait for <b>Done webpack</b> in server logs before logging in...</h3>
    </div>
  </div>
</body>
</html>

<script>
	setTimeout(function() {
		window.location.href = "/";
	}, 10000);
</script>
`

	log.Printf("Successfully created the %s GitHub App.", *config.Name)

	data := struct {
		Name string
	}{
		Name: *config.Name,
	}
	t := template.Must(template.New("success").Parse(tpl))
	if err := t.Execute(w, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}
	if err := env.Save("public/.env", map[string]string{
		"QUICKFEED_APP_URL": *config.HTMLURL,
	}); err != nil {
		return err
	}
	log.Printf("App URL saved in public/.env: %s", *config.HTMLURL)
	go runWebpack()
	return nil
}

func runWebpack() {
	log.Println("Running webpack...")
	c := exec.Command("webpack")
	c.Dir = "public"
	if err := c.Run(); err != nil {
		log.Print(c.Output())
		log.Print(err)
		log.Print("Failed to run webpack; trying npm ci")
		if ok := runNpmCi(); !ok {
			return
		}
	}
	log.Print("Done webpack")
}

func runNpmCi() bool {
	log.Println("Running npm ci...")
	c := exec.Command("npm", "ci")
	c.Dir = "public"
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
			"hook_attributes": {
				"active": false,
				"url": "",
			},
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
				"organization_hooks": "write",
			},
		})
		document.getElementById('create').submit()
	</script>
	`

	data := struct {
		URL         string
		Name        string
		CallbackURL string
	}{
		URL:         auth.GetBaseURL(domain),
		Name:        env.AppName(),
		CallbackURL: auth.GetCallbackURL(domain),
	}
	t := template.Must(template.New("form").Parse(tpl))
	if err := t.Execute(w, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}
	return nil
}
