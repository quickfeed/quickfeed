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
	"github.com/quickfeed/quickfeed/web/auth"
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

	// Send another event with another organization.
	response = sendEvent(t, event{
		OrganizationLogin: "qf103-2022",
		OrganizationScmID: 2,
		UserLogin:         "quickfeed",
		UserScmID:         1,
	}, server)

	if response.StatusCode != 200 {
		t.Errorf("got status %d, want 200", response.StatusCode)
	}

	// Second course should be created successfully.
	course, err = wh.db.GetCourseByOrganizationID(2)
	if err != nil {
		t.Fatal(err)
	}

	if course.Name != "qf103-2022" {
		t.Errorf("got course name %s, want qf102-2022", course.Name)
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

	if _, err := wh.db.GetCourseByOrganizationID(uint64(invalidOrgID)); err == nil {
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

	// Send an event with an invalid (not being listened for) action.
	// This should not create a course, rather it should simply return.
	_ = sendEvent(t, event{
		Action:            "deleted",
		OrganizationLogin: "qf102-2022",
		OrganizationScmID: 1,
		UserLogin:         "quickfeed",
		UserScmID:         int(admin.ScmRemoteID),
	}, server)

	if _, err := wh.db.GetCourseByOrganizationID(1); err == nil {
		t.Fatal(err)
	}
}

func TestCheckUserClaims(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	wh, server := setupWebhook(t, db)
	defer server.Close()
	admin := qtest.CreateFakeCustomUser(t, db, &qf.User{
		Name:        "admin",
		Login:       "quickfeed",
		ScmRemoteID: 1,
	})

	token, err := wh.tm.NewAuthCookie(admin.ID)
	if err != nil {
		t.Fatal(err)
	}
	claims, err := wh.tm.GetClaims(token.String())
	if err != nil {
		t.Fatal(err)
	}

	if len(claims.Courses) != 0 {
		t.Errorf("got %d courses, want 0", len(claims.Courses))
	}

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

	// Check successful course creation queued the users claims to be updated.
	token, err = wh.tm.UpdateCookie(claims)
	if err != nil {
		t.Fatal(err)
	}

	claims, err = wh.tm.GetClaims(token.String())
	if err != nil {
		t.Fatal(err)
	}
	if len(claims.Courses) != 1 {
		t.Errorf("got %d courses, want 0", len(claims.Courses))
	}

	if claims.Courses[course.ID] != qf.Enrollment_TEACHER {
		t.Errorf("got %d status, want %d", claims.Courses[course.ID], qf.Enrollment_TEACHER)
	}
}

func setupWebhook(t *testing.T, db database.Database) (*GitHubWebHook, *httptest.Server) {
	_, manager := scm.MockSCMManager(t)
	tm, err := auth.NewTokenManager(db)
	if err != nil {
		t.Fatal(err)
	}
	wh := NewGitHubWebHook(qtest.Logger(t), db, manager, &ci.Local{}, "", stream.NewStreamServices(), tm)

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
