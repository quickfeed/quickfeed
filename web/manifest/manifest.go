package manifest

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/google/go-github/v45/github"
	"github.com/quickfeed/quickfeed/internal/env"
)

type Manifest struct {
	http.Server
	done chan bool
}

func NewManifest(addr string) *Manifest {
	router := http.NewServeMux()
	m := &Manifest{
		done: make(chan bool),
	}
	router.Handle("/manifest/callback", m.Conversion())
	router.Handle("/manifest", m.CreateApp())
	m.Server = http.Server{
		Addr:    addr,
		Handler: router,
	}
	return m
}

func StartFlow(addr string) error {
	m := NewManifest(addr)
	if err := m.check(); err != nil {
		return err
	}
	go func() {
		if err := m.ListenAndServe(); err != nil {
			fmt.Printf("Failed to start web server: %v\n", err)
			m.done <- true
		}
	}()
	defer m.Close()

	fmt.Printf("Go to https://%s/manifest to create an app.\n", env.Domain())

	<-m.done
	// Refresh enviroment variables
	return env.Load("")
}

func (m *Manifest) Conversion() http.HandlerFunc {
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
		if resp.StatusCode != http.StatusCreated {
			w.WriteHeader(resp.StatusCode)
			fmt.Fprintf(w, "Error: %s", resp.Status)
			return
		}
		env.Save("", *config.PEM, *config.ClientID)
	}
}

func (m *Manifest) CreateApp() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		m.form(w)
	}
}

func (m Manifest) form(w http.ResponseWriter) {
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
			"name": "QuickFeed",
			"url": "{{.URL}}",
			"hook_attributes": {
				"active": false,
				"url": "",
			},
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
		URL string
	}{
		URL: "https://" + env.Domain(),
	}

	if err := t.Execute(w, data); err != nil {
		log.Fatalf("Failed to execute template: %v", err)
	}
}

func (m Manifest) check() error {
	if env.HasAppEnvs() {
		fmt.Println("WARNING: Backup any existing app configuration. Continuing will delete all existing app configuration.")
		if !answer() {
			return fmt.Errorf("aborting GitHub app creation")
		}
	}
	if env.Domain() == "localhost" || env.Domain() == "127.0.0.1" {
		fmt.Printf("WARNING: You are creating an app on %s. Only do this for development purposes.\n", env.Domain())
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
