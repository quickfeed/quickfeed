package scm

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/go-github/v62/github"
	"github.com/migueleliasweb/go-github-mock/src/mock"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/shurcooL/githubv4"
	"go.uber.org/zap"
)

const (
	organizationName = "foobar"
)

// Note: GetByID uses the undocumented GitHub API endpoint /organizations/:id.
var GetByID mock.EndpointPattern = mock.EndpointPattern{
	Pattern: "/organizations/{id}",
	Method:  "GET",
}

// NewGithubSCMClient returns a new Github client implementing the SCM interface.
func NewMockGithubSCMClient(logger *zap.SugaredLogger) *GithubSCM {
	mockedHTTPClient := mock.NewMockedHTTPClient(
		mock.WithRequestMatchHandler(
			GetByID,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Printf("GetByID: %v\n", r.PathValue("id"))
				fmt.Printf("GetByID: %v\n", r.URL)

				w.Write(mock.MustMarshal(github.Organization{
					ID:    github.Int64(456),
					Login: github.String("foobar_org_mocked_x"),
				}))
			}),
		),
		mock.WithRequestMatchHandler(
			mock.GetOrgsByOrg,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Printf("GetOrgsByOrg: %v\n", r.URL.Path)

				w.Write(mock.MustMarshal(github.Organization{
					ID:    github.Int64(123),
					Login: github.String("foobar_org_mocked"),
				}))
			}),
		),
		// mock.WithRequestMatch(
		// 	mock.GetOrgsByOrg,
		// 	github.Organization{
		// 		ID:    github.Int64(123),
		// 		Login: github.String(organizationName),
		// 	},
		// 	github.Organization{
		// 		ID:    github.Int64(456),
		// 		Login: github.String("bar_org"),
		// 	},
		// ),
		mock.WithRequestMatch(
			mock.GetOrgsReposByOrg,
			[]github.Repository{
				{Name: github.String("bar_org")},
			},
		),
	)
	httpClient := mockedHTTPClient

	return &GithubSCM{
		logger:      logger,
		client:      github.NewClient(httpClient),
		clientV4:    githubv4.NewClient(httpClient),
		providerURL: "github.com",
	}
}

func TestGetOrganization2(t *testing.T) {
	s := NewMockGithubSCMClient(qtest.Logger(t))
	org, err := s.GetOrganization(context.Background(), &OrganizationOptions{
		Name: "bar_org",
		// Username: "foobar",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("org: %v", org)

	org, err = s.GetOrganization(context.Background(), &OrganizationOptions{
		Name: organizationName,
		// Username: "foobar",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("org: %v", org)

	org, err = s.GetOrganization(context.Background(), &OrganizationOptions{
		ID: 123,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("org: %v", org)
}

func TestGetRepositories(t *testing.T) {
	s := NewMockGithubSCMClient(qtest.Logger(t))
	repo, err := s.GetRepositories(context.Background(), &qf.Organization{
		ScmOrganizationID:   123,
		ScmOrganizationName: "foobar_org2",
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range repo {
		t.Logf("repo: %v", r)
	}
}
