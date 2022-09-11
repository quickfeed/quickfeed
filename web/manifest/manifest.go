package manifest

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/google/go-github/v45/github"
	"github.com/quickfeed/quickfeed/internal/env"
	"github.com/quickfeed/quickfeed/web"
)

// TODO(meling) Should reuse those in env package.
const (
	defaultPath  = "internal/config/github/quickfeed.pem"
	appID        = "QUICKFEED_APP_ID"
	appKey       = "QUICKFEED_APP_KEY"
	clientID     = "QUICKFEED_CLIENT_ID"
	clientSecret = "QUICKFEED_CLIENT_SECRET"
)

type manifest struct {
	handler http.Handler
	done    chan bool
}

func New() *manifest {
	m := &manifest{
		done: make(chan bool),
	}
	router := http.NewServeMux()
	router.Handle("/manifest/callback", m.conversion())
	router.Handle("/manifest", createApp())
	m.handler = router
	return m
}

func (m *manifest) Handler() http.Handler {
	return m.handler
}

func (m *manifest) StartAppCreationFlow(server *web.Server) error {
	if err := check(); err != nil {
		return err
	}
	go func() {
		if err := server.Serve(); err != nil {
			fmt.Printf("Failed to start web server: %v\n", err)
			m.done <- true
		}
	}()
	defer server.Shutdown(context.Background())
	fmt.Printf("Go to https://%s/manifest to create an app.\n", env.Domain())
	<-m.done
	// Refresh environment variables
	return env.Load("")
}

func (m *manifest) conversion() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		defer func() {
			m.done <- true
		}()
		if code == "" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "No code provided")
			return
		}
		ctx := context.Background()
		config, resp, err := github.NewClient(nil).Apps.CompleteAppManifest(ctx, code)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "Error: %s", err)
			return
		}
		// GitHub returns 201 Created on success
		if resp.StatusCode != http.StatusCreated {
			w.WriteHeader(resp.StatusCode)
			fmt.Fprintf(w, "Error: %s", resp.Status)
			return
		}

		// Save PEM file to default location
		if err := os.MkdirAll(filepath.Dir(defaultPath), 0o700); err != nil {
			fmt.Println("Failed to create directory:", err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Error: %s", err)
			return
		}

		// Write PEM file
		if err := os.WriteFile(defaultPath, []byte(*config.PEM), 0o644); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Error: %s", err)
			return
		}

		// Save the application configuration to the .env file
		envToUpdate := map[string]string{
			appID:        strconv.FormatInt(*config.ID, 10),
			appKey:       env.AppKey(),
			clientID:     *config.ClientID,
			clientSecret: *config.ClientSecret,
		}
		if err := env.Save(".env", envToUpdate); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Error: %s", err)
			return
		}

		// Print success message, and redirect to main page
		success(w)
	}
}

func createApp() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		form(w)
	}
}

func success(w http.ResponseWriter) {
	const tpl = `
		<html>
			Successfully created app.
		</html>

		<script>
			setTimeout(function() {
				window.location.href = "/";
			}, 5000);
		</script>
	`
	fmt.Fprint(w, tpl)
}

func form(w http.ResponseWriter) {
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
				"pull_requests": "write"
			},
		})
		document.getElementById('create').submit()
	</script>
	`
	t := template.Must(template.New("form").Parse(tpl))

	data := struct {
		URL         string
		Name        string
		CallbackURL string
	}{
		URL:         "https://" + env.Domain(),
		Name:        env.AppName(),
		CallbackURL: "https://" + env.Domain() + "/auth/callback",
	}

	if err := t.Execute(w, data); err != nil {
		fmt.Printf("Failed to execute template: %v", err)
	}
}

func check() error {
	if env.HasAppEnvs() {
		fmt.Println("WARNING: Backup any existing app configuration. Continuing will delete all existing app configuration.")
		if !answer() {
			return fmt.Errorf("aborting GitHub app creation")
		}
	}
	if env.Domain() == "localhost" || env.Domain() == "127.0.0.1" {
		fmt.Printf("WARNING: You are creating an app on %s. Only for development purposes.\n", env.Domain())
		if !answer() {
			return fmt.Errorf("aborting GitHub app creation")
		}
	}
	return nil
}

func answer() bool {
	fmt.Println("Continue? (Y/n) ")
	var answer string
	fmt.Scanln(&answer)
	return answer == "Y" || answer == "y"
}
