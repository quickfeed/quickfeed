package scm_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/scm"
)

func TestMockClone(t *testing.T) {
	s := scm.NewMockSCMClient()
	ctx := context.Background()
	cloneTests := []struct {
		name     string
		opt      *scm.CloneOptions
		wantPath string
		wantErr  bool
	}{
		{
			name: "student repository",
			opt: &scm.CloneOptions{
				Organization: qtest.MockOrg,
				Repository:   qf.StudentRepoName("user"),
			},
			wantPath: filepath.Join("testdata", "assignments"),
			wantErr:  false,
		},
		{
			name: "tests repository",
			opt: &scm.CloneOptions{
				Organization: qtest.MockOrg,
				Repository:   qf.TestsRepo,
			},
			wantPath: filepath.Join("testdata", "tests"),
			wantErr:  false,
		},
		{
			name: "missing organization info",
			opt: &scm.CloneOptions{
				Repository: qf.StudentRepoName("user"),
			},
			wantPath: "",
			wantErr:  true,
		},
	}
	for _, tt := range cloneTests {
		path, err := s.Clone(ctx, tt.opt)
		if (err != nil) != tt.wantErr {
			t.Errorf("%s: expected error %v, got = %v, ", tt.name, tt.wantErr, err)
		}
		if path != tt.wantPath {
			t.Errorf("%s: expected path '%s', got '%s'", tt.name, tt.wantPath, path)
		}
	}
}
