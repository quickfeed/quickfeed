package web_test

import (
	"fmt"
	"net/http"
	"os"
	"reflect"
	"slices"
	"testing"

	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf/qfconnect"
	"github.com/quickfeed/quickfeed/scm"
	"github.com/quickfeed/quickfeed/web"
	"github.com/quickfeed/quickfeed/web/auth"
	"github.com/steinfletcher/apitest"
)

func TestRegisterRouter(t *testing.T) {
	logger := qtest.Logger(t).Desugar()
	db, stop := qtest.TestDB(t)
	defer stop()

	mgr := scm.MockManager(t, scm.WithMockOrgs())
	qf := web.NewQuickFeedService(logger, db, mgr, nil, &auth.TokenManager{})

	public := createTempPublicDir(t)
	webHookSecret := ""
	mux := qf.RegisterRouter(webHookSecret, public)

	apitest.New("Index").
		Handler(mux).
		Get("/").
		Expect(t).
		Status(http.StatusOK).
		Body("hello, world!").
		End()

	apitest.New("Robots").
		Handler(mux).
		Get("/robots.txt").
		Expect(t).
		Status(http.StatusOK).
		Body("User-agent: *\nDisallow: /").
		End()

	partialUrl := "/" + qfconnect.QuickFeedServiceName + "/"
	qfType := reflect.TypeOf(qfconnect.UnimplementedQuickFeedServiceHandler{})
	for i := 0; i < qfType.NumMethod(); i++ {
		method := qfType.Method(i)
		apitest.New(method.Name).
			Handler(mux).
			Post(partialUrl+method.Name).
			Header("Content-Type", "application/json").
			Body("{}").
			Expect(t).Assert(func(resp *http.Response, req *http.Request) error {
			// 415 (Unsupported Media Type) is returned for requests with unsupported content type
			// 		- this applies to all streaming methods
			// 400 (Bad Request) is returned if the request is malformed, e.g. missing required fields
			// 		- for a majority of the unary methods, "{}" is considered a malformed request
			// 401 (Unauthorized) is returned if the user is not authenticated
			// 		- this applies to all methods where "{}" is a valid request, but the user is not authenticated
			if !wantStatus(resp, http.StatusUnauthorized, http.StatusBadRequest, http.StatusUnsupportedMediaType) {
				return fmt.Errorf("%s: expected status code 401, 400 or 415, got %d", method.Name, resp.StatusCode)
			}
			return nil
		}).End()
	}

	// Invalid (non-existing) RPC request should return 404 (Not Found)
	apitest.New("Invalid method").
		Handler(mux).
		Post(partialUrl+"NonExistingMethod").
		Header("Content-Type", "application/json").
		Body("{}").
		Expect(t).
		Status(http.StatusNotFound).
		End()
}

func wantStatus(resp *http.Response, wantStatus ...int) bool {
	return slices.Contains(wantStatus, resp.StatusCode)
}

func createTempPublicDir(t *testing.T) string {
	t.Helper()
	publicDir := t.TempDir() + "/public"
	if err := os.MkdirAll(publicDir+"/assets", 0o700); err != nil {
		t.Fatal(err)
	}
	if err := createFile(t, publicDir+"/assets/index.html", "hello, world!"); err != nil {
		t.Fatal(err)
	}
	if err := createFile(t, publicDir+"/assets/robots.txt", "User-agent: *\nDisallow: /"); err != nil {
		t.Fatal(err)
	}
	return publicDir
}

func createFile(t *testing.T, path, content string) error {
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := file.WriteString(content); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	return nil
}
