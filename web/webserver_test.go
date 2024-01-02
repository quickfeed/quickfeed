package web_test

import (
	"bytes"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf/qfconnect"
	"github.com/quickfeed/quickfeed/scm"
	"github.com/quickfeed/quickfeed/web"
	"github.com/quickfeed/quickfeed/web/auth"
)

func TestRegisterRouter(t *testing.T) {
	logger := qtest.Logger(t).Desugar()
	db, close := qtest.TestDB(t)
	defer close()

	_, mgr := scm.MockSCMManager(t)
	qf := web.NewQuickFeedService(logger, db, mgr, web.BaseHookOptions{}, nil)

	authConfig := auth.NewGitHubConfig("", &scm.Config{})
	mux := qf.RegisterRouter(&auth.TokenManager{}, authConfig, "../public")

	server := httptest.NewTLSServer(mux)
	defer server.Close()

	resp, err := server.Client().Get(server.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("'/': expected code 200, got %d", resp.StatusCode)
	}

	body := bytes.NewReader([]byte("{}"))
	partialUrl := server.URL + "/" + qfconnect.QuickFeedServiceName + "/"
	qfType := reflect.TypeOf(qfconnect.UnimplementedQuickFeedServiceHandler{})
	for i := 0; i < qfType.NumMethod(); i++ {
		method := qfType.Method(i)
		resp, err = server.Client().Post(partialUrl+method.Name, "application/json", body)
		if err != nil {
			t.Fatal(err)
		}
		if !(resp.StatusCode == 401 || resp.StatusCode == 400 || resp.StatusCode == 415) {
			t.Errorf("'%s': expected code 401, 400 or 415, got %d\n", method.Name, resp.StatusCode)
		}
	}

	resp, err = server.Client().Post(partialUrl+"NonExistingMethod", "application/json", body)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != 404 {
		t.Errorf("expected 404, got %d\n", resp.StatusCode)
	}
}
