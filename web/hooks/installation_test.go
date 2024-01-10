package hooks

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/quickfeed/quickfeed/ci"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/scm"
	"github.com/quickfeed/quickfeed/web/stream"
)

func TestReceiveInstallationEvent(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	_, manager := scm.MockSCMManager(t)

	wh := NewGitHubWebHook(qtest.Logger(t), db, manager, &ci.Local{}, "", stream.NewStreamServices())
	admin := qtest.CreateFakeCustomUser(t, db, &qf.User{
		Name:        "admin",
		Login:       "quickfeed",
		ScmRemoteID: 1,
	})

	router := http.NewServeMux()
	router.HandleFunc("/hook/", wh.Handle())
	server := httptest.NewServer(router)
	defer server.Close()

	body := []byte(`
		{
		"action": "created",
		"installation": {
			"id": 45223417,
			"account": {
			"login": "qf102-2022",
			"id": 1,
			"type": "Organization"
			}
		},
		"sender": {
			"login": "quickfeed",
			"id": 1
		}
	}`)

	req, err := http.NewRequest("POST", server.URL+"/hook/", bytes.NewReader(body))
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
	if resp.StatusCode != 200 {
		t.Errorf("got status %d, want 200", resp.StatusCode)
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
