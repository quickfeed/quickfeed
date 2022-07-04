package scm_test

import (
	"context"
	"testing"

	pb "github.com/quickfeed/quickfeed/ag"
	"github.com/quickfeed/quickfeed/kit/sh"
	"github.com/quickfeed/quickfeed/log"
	"github.com/quickfeed/quickfeed/scm"
)

func TestClone(t *testing.T) {
	qfTestOrg := scm.GetTestOrganization(t)
	accessToken := scm.GetAccessToken(t)

	s, err := scm.NewSCMClient(log.Zap(true).Sugar(), "github", accessToken)
	if err != nil {
		t.Fatal(err)
	}
	userName, err := s.GetUserName(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	dstDir := t.TempDir()
	assignmentDir, err := s.Clone(ctx, &scm.CloneOptions{
		Organization: qfTestOrg,
		Repository:   pb.StudentRepoName(userName),
		DestDir:      dstDir,
	})
	if err != nil {
		t.Fatal(err)
	}
	testsDir, err := s.Clone(ctx, &scm.CloneOptions{
		Organization: qfTestOrg,
		Repository:   pb.TestsRepo,
		DestDir:      dstDir,
	})
	if err != nil {
		t.Fatal(err)
	}
	o, err := sh.OutputA("ls", "-laR", dstDir)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("destDir: ", dstDir)
	t.Log("cloneDir:", assignmentDir)
	t.Log("testsDir:", testsDir)
	t.Log(o)
}
