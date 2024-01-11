package hooks

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"text/template"

	"github.com/quickfeed/quickfeed/ci"
	"github.com/quickfeed/quickfeed/database"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/scm"
	"github.com/quickfeed/quickfeed/web/stream"
)

func TestReceiveInstallationEvent(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	wh, server := setupWebhook(t, db)
	defer server.Close()
	admin := qtest.CreateFakeCustomUser(t, db, &qf.User{
		Name:        "admin",
		Login:       "quickfeed",
		ScmRemoteID: 1,
	})

	response := sendEvent(t, event{
		OrganizationLogin: "qf102-2022",
		OrganizationScmID: 1,
		UserLogin:         "quickfeed",
		UserScmID:         1,
	}, server)

	if response.StatusCode != 200 {
		t.Errorf("got status %d, want 200", response.StatusCode)
	}

	course, err := wh.db.GetCourseByOrganizationID(1)
	if err != nil {
		t.Fatal(err)
	}

	if course.Name != "qf102-2022" {
		t.Errorf("got course name %s, want qf102-2022", course.Name)
	}

	if course.CourseCreatorID != admin.ID {
		t.Errorf("got course creator id %d, want 1", course.CourseCreatorID)
	}
}

func TestNonExistingUser(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	wh, server := setupWebhook(t, db)
	defer server.Close()

	// Send an event with a valid organization, but an invalid user.
	_ = sendEvent(t, event{
		OrganizationLogin: "qf102-2022",
		OrganizationScmID: 1,
		UserLogin:         "quickfeed",
		UserScmID:         1000,
	}, server)

	_, err := wh.db.GetCourseByOrganizationID(1)
	if err == nil {
		t.Fatal(err)
	}
}

func TestNonExistingOrganization(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	admin := qtest.CreateFakeCustomUser(t, db, &qf.User{
		Name:        "admin",
		Login:       "quickfeed",
		ScmRemoteID: 1,
	})

	wh, server := setupWebhook(t, db)
	defer server.Close()

	// Send an event with a valid user, but an invalid organization.
	invalidOrgID := 1000
	_ = sendEvent(t, event{
		OrganizationLogin: "qf102-2022",
		OrganizationScmID: invalidOrgID,
		UserLogin:         "quickfeed",
		UserScmID:         int(admin.ScmRemoteID),
	}, server)

	_, err := wh.db.GetCourseByOrganizationID(uint64(invalidOrgID))
	if err != nil {
		t.Fatal(err)
	}
}

func TestInvalidAction(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	admin := qtest.CreateFakeCustomUser(t, db, &qf.User{
		Name:        "admin",
		Login:       "quickfeed",
		ScmRemoteID: 1,
	})

	wh, server := setupWebhook(t, db)
	defer server.Close()

	invalidOrgID := 1000
	_ = sendEvent(t, event{
		Action:            "deleted",
		OrganizationLogin: "qf102-2022",
		OrganizationScmID: invalidOrgID,
		UserLogin:         "quickfeed",
		UserScmID:         int(admin.ScmRemoteID),
	}, server)

	_, err := wh.db.GetCourseByOrganizationID(uint64(invalidOrgID))
	if err == nil {
		t.Fatal(err)
	}
}

func setupWebhook(t *testing.T, db database.Database) (*GitHubWebHook, *httptest.Server) {
	_, manager := scm.MockSCMManager(t)

	wh := NewGitHubWebHook(qtest.Logger(t), db, manager, &ci.Local{}, "", stream.NewStreamServices())

	router := http.NewServeMux()
	router.HandleFunc("/hook/", wh.Handle())
	server := httptest.NewServer(router)

	return wh, server
}

type event struct {
	Action            string
	OrganizationLogin string
	OrganizationScmID int
	UserLogin         string
	UserScmID         int
}

const eventTemplate = `
		{
		"action": "{{ .Action }}",
		"installation": {
			"id": 45223417,
			"account": {
			"login": "{{ .OrganizationLogin }}",
			"id": {{ .OrganizationScmID  }},
			"type": "Organization"
			}
		},
		"sender": {
			"login": "{{ .UserLogin }}",
			"id": {{ .UserScmID }}
		}
	}`

func sendEvent(t *testing.T, event event, server *httptest.Server) *http.Response {
	t.Helper()
	body, err := template.New("event").Parse(eventTemplate)
	if err != nil {
		t.Fatal(err)
	}

	if event.Action == "" {
		event.Action = "created"
	}

	var buf bytes.Buffer
	if err = body.Execute(&buf, event); err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("POST", server.URL+"/hook/", &buf)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-GitHub-Event", "installation")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	return resp
}
