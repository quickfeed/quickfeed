package hooks

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"text/template"
	"time"

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

	wantCourse1 := qtest.MockCourses[0]
	response := sendEvent(t, event{
		OrganizationLogin: wantCourse1.GetScmOrganizationName(),
		OrganizationScmID: int(wantCourse1.GetScmOrganizationID()),
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
	if course.GetScmOrganizationName() != wantCourse1.GetScmOrganizationName() {
		t.Errorf("got course %q, want %q", course.GetScmOrganizationName(), wantCourse1.GetScmOrganizationName())
	}
	if course.GetCourseCreatorID() != admin.GetID() {
		t.Errorf("got course creator id %d, want 1", course.GetCourseCreatorID())
	}

	// Send another event with another organization.
	wantCourse2 := qtest.MockCourses[1]
	response = sendEvent(t, event{
		OrganizationLogin: wantCourse2.GetScmOrganizationName(),
		OrganizationScmID: int(wantCourse2.GetScmOrganizationID()),
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
	if course.GetScmOrganizationName() != wantCourse2.GetScmOrganizationName() {
		t.Errorf("got course %s, want %s", course.GetScmOrganizationName(), wantCourse2.GetScmOrganizationName())
	}
}

// To verify that we are not creating a new course if the course repositories already exist.
// We cannot check this in this test directly, since we cannot pass the actual error through the webhook.
// Hence, this test should be run with LOG=1 to see the error message.
//
//	LOG=1 go test -v -run TestAlreadyExistingCourse
func TestAlreadyExistingCourse(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	wh, server := setupWebhook(t, db)
	defer server.Close()
	_ = qtest.CreateFakeCustomUser(t, db, &qf.User{Name: "admin", Login: "quickfeed", ScmRemoteID: 1})

	wantCourse := qtest.MockCourses[0]
	response := sendEvent(t, event{
		OrganizationLogin: wantCourse.GetScmOrganizationName(),
		OrganizationScmID: int(wantCourse.GetScmOrganizationID()),
		UserLogin:         "quickfeed",
		UserScmID:         1,
	}, server)
	if response.StatusCode != 200 {
		t.Errorf("got status %d, want 200", response.StatusCode)
	}

	course, err := wh.db.GetCourseByOrganizationID(wantCourse.GetScmOrganizationID())
	if err != nil {
		t.Fatal(err)
	}
	if course.GetScmOrganizationName() != wantCourse.GetScmOrganizationName() {
		t.Errorf("got course %s, want %s", course.GetScmOrganizationName(), wantCourse.GetScmOrganizationName())
	}

	// Send the same event again, this should not create a new course.
	// This should log an scm.ErrAlreadyExists error; check running with LOG=1.
	response = sendEvent(t, event{
		OrganizationLogin: wantCourse.GetScmOrganizationName(),
		OrganizationScmID: int(wantCourse.GetScmOrganizationID()),
		UserLogin:         "quickfeed",
		UserScmID:         1,
	}, server)
	if response.StatusCode != 200 {
		t.Errorf("got status %d, want 200", response.StatusCode)
	}

	courses, err := wh.db.GetCourses()
	if err != nil {
		t.Fatal(err)
	}
	if len(courses) != 1 {
		t.Errorf("got %d courses, want 1", len(courses))
	}

	course, err = wh.db.GetCourseByOrganizationID(wantCourse.GetScmOrganizationID())
	if err != nil {
		t.Fatal(err)
	}
	if course.GetScmOrganizationName() != wantCourse.GetScmOrganizationName() {
		t.Errorf("got course %s, want %s", course.GetScmOrganizationName(), wantCourse.GetScmOrganizationName())
	}
}

func TestNonAdminUser(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	// first created user is admin
	_ = qtest.CreateFakeUser(t, db)

	user := qtest.CreateFakeCustomUser(t, db, &qf.User{
		Name:        "user",
		Login:       "quickfeed",
		ScmRemoteID: 1000,
	})

	wh, server := setupWebhook(t, db)
	defer server.Close()

	// Send an event with a valid organization, but an invalid user.
	_ = sendEvent(t, event{
		OrganizationLogin: "qf102-2022",
		OrganizationScmID: 1,
		UserLogin:         "quickfeed",
		UserScmID:         int(user.GetScmRemoteID()),
	}, server)

	course, err := wh.db.GetCourseByOrganizationID(1)
	if err == nil {
		// expect error: record not found
		t.Errorf("got course %v, want error", course)
	}

	if course != nil {
		t.Errorf("got course %v, want nil", course)
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
		UserScmID:         int(admin.GetScmRemoteID()),
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
		UserScmID:         int(admin.GetScmRemoteID()),
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

	token, err := wh.tm.NewAuthCookie(admin.GetID())
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

	if course.GetName() != "qf102-2022" {
		t.Errorf("got course name %s, want qf102-2022", course.GetName())
	}

	if course.GetCourseCreatorID() != admin.GetID() {
		t.Errorf("got course creator id %d, want 1", course.GetCourseCreatorID())
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

	if claims.Courses[course.GetID()] != qf.Enrollment_TEACHER {
		t.Errorf("got %d status, want %d", claims.Courses[course.GetID()], qf.Enrollment_TEACHER)
	}
}

func setupWebhook(t *testing.T, db database.Database) (*GitHubWebHook, *httptest.Server) {
	mgr := scm.MockManager(t, scm.WithMockOrgs("quickfeed"))
	tm, err := auth.NewTokenManager(db)
	if err != nil {
		t.Fatal(err)
	}
	wh := NewGitHubWebHook(qtest.Logger(t), db, mgr, &ci.Local{}, "", stream.NewStreamServices(), tm)

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

func Test_defaultYearAndTag(t *testing.T) {
	tests := []struct {
		name     string
		now      time.Time
		wantYear uint32
		wantTag  string
	}{
		{
			name:     "january",
			now:      time.Date(2022, time.January, 1, 0, 0, 0, 0, time.UTC),
			wantYear: 2022,
			wantTag:  "Spring",
		},
		{
			name:     "december",
			now:      time.Date(2022, time.December, 31, 0, 0, 0, 0, time.UTC),
			wantYear: 2023,
			wantTag:  "Spring",
		},
		{
			name:     "november",
			now:      time.Date(2022, time.November, 30, 0, 0, 0, 0, time.UTC),
			wantYear: 2023,
			wantTag:  "Spring",
		},
		{
			name:     "october",
			now:      time.Date(2022, time.October, 31, 0, 0, 0, 0, time.UTC),
			wantYear: 2022,
			wantTag:  "Fall",
		},
		{
			name:     "april",
			now:      time.Date(2022, time.April, 30, 0, 0, 0, 0, time.UTC),
			wantYear: 2022,
			wantTag:  "Fall",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := defaultYear(tt.now); got != tt.wantYear {
				t.Errorf("defaultYear() = %v, want %v", got, tt.wantYear)
			}
			if got := defaultTag(tt.now); got != tt.wantTag {
				t.Errorf("defaultTag() = %v, want %v", got, tt.wantTag)
			}
		})
	}
}
