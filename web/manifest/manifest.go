package manifest

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/google/go-github/v45/github"
	"github.com/quickfeed/quickfeed/internal/env"
	"github.com/quickfeed/quickfeed/web"
)

const (
	appID        = "QUICKFEED_APP_ID"
	appKey       = "QUICKFEED_APP_KEY"
	clientID     = "QUICKFEED_CLIENT_ID"
	clientSecret = "QUICKFEED_CLIENT_SECRET"
)

type Manifest struct {
	handler http.Handler
	done    chan error
}

func New() *Manifest {
	m := &Manifest{
		done: make(chan error),
	}
	router := http.NewServeMux()
	router.Handle("/manifest/callback", m.conversion())
	router.Handle("/manifest", createApp())
	m.handler = router
	return m
}

func (m *Manifest) Handler() http.Handler {
	return m.handler
}

func (m *Manifest) StartAppCreationFlow(server *web.Server) error {
	if err := check(); err != nil {
		return err
	}
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
      <h2>QuickFeed GitHub App installed</h2>
      <h3>Redirecting...</h3>
    </div>
  </div>
</body>
</html>

<script>
	setTimeout(function() {
		window.location.href = "/";
	}, 5000);
</script>
	`
	fmt.Fprint(w, tpl)
	log.Println("Successfully installed the QuickFeed GitHub App.")
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
