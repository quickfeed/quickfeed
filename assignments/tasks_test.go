package assignments

import (
	"context"
	pb "github.com/autograde/quickfeed/ag"
	"github.com/autograde/quickfeed/scm"
	"go.uber.org/zap"
	"testing"
)

// Running this test case will create given task on all the repositories in side the organization
func TestCreateTasks(t *testing.T) {
	qfTestOrg := scm.GetTestOrganization(t)
	accessToken := scm.GetAccessToken(t)

	s, err := scm.NewSCMClient(zap.NewNop().Sugar(), "github", accessToken)
	if err != nil {
		t.Fatal(err)
	}

	course := &pb.Course{
		Name:             "QuickFeed-Dev",
		OrganizationPath: qfTestOrg,
	}
	tasks := []*pb.Task{
		{
			Title: "Task 11",
			Body:  "Body of task 1",
		},
		{
			Title: "Task 22",
			Body:  "Body of task 2",
		},
	}
	assignment := []*pb.Assignment{
		{
			Tasks: tasks,
		},
	}

	err = SyncTasks(context.Background(), zap.NewNop().Sugar(), s, course, assignment)
	if err != nil {
		t.Fatal(err)
	}

}
