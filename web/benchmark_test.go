package web_test

import (
	"errors"
	"testing"

	"connectrpc.com/connect"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/scm"
	"github.com/quickfeed/quickfeed/web"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestCreateBenchmark(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	t.Cleanup(cleanup)
	client := web.NewMockClient(t, db, scm.WithMockOrgs(), web.WithInterceptors())
	teacher, course, assignment, _ := qtest.SetupCourseAssignmentTeacherStudent(t, db)
	cookie := client.Cookie(t, teacher)

	tests := []struct {
		name      string
		benchmark *qf.GradingBenchmark
		wantErr   error
	}{
		{
			name: "Valid benchmark",
			benchmark: &qf.GradingBenchmark{
				CourseID:     course.GetID(),
				AssignmentID: assignment.GetID(),
				Heading:      "Benchmark 1",
				Comment:      "comment 1",
			},
		},
		{
			name: "Non-existing assignment",
			benchmark: &qf.GradingBenchmark{
				CourseID:     course.GetID(),
				AssignmentID: 111,
				Heading:      "This is a test benchmark",
				Comment:      "",
			},
			wantErr: connect.NewError(connect.CodeInvalidArgument, errors.New("failed to create benchmark")),
		},
	}
	for i, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gotBenchmark, err := client.CreateBenchmark(t.Context(), qtest.RequestWithCookie(test.benchmark, cookie))
			qtest.CheckError(t, err, test.wantErr)

			if test.wantErr == nil {
				test.benchmark.ID = uint64(i + 1)
				qtest.Diff(t, "CreateBenchmark mismatch", gotBenchmark.Msg, test.benchmark, protocmp.Transform())
			}
		})
	}
}

func TestUpdateBenchmark(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	t.Cleanup(cleanup)
	client := web.NewMockClient(t, db, scm.WithMockOrgs(), web.WithInterceptors())
	teacher, course, assignment, _ := qtest.SetupCourseAssignmentTeacherStudent(t, db)
	qtest.CreateBenchmark(t, db, &qf.GradingBenchmark{AssignmentID: assignment.GetID(), CourseID: course.GetID()})
	cookie := client.Cookie(t, teacher)

	wantBenchmark := &qf.GradingBenchmark{
		ID:           1,
		CourseID:     course.GetID(),
		AssignmentID: assignment.GetID(),
		Heading:      "Updated Benchmark",
		Comment:      "Updated comment",
	}

	if _, err := client.UpdateBenchmark(t.Context(), qtest.RequestWithCookie(wantBenchmark, cookie)); err != nil {
		t.Fatal(err)
	}
	benchmarks := qtest.GetBenchmarks(t, db, assignment.GetID())
	if len(benchmarks) != 1 {
		t.Fatalf("expected 1 benchmark, got %d", len(benchmarks))
	}
	gotBenchmark := benchmarks[0]
	qtest.Diff(t, "UpdateBenchmark mismatch", gotBenchmark, wantBenchmark, protocmp.Transform())
}

func TestDeleteBenchmark(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	t.Cleanup(cleanup)
	client := web.NewMockClient(t, db, scm.WithMockOrgs(), web.WithInterceptors())
	teacher, course, assignment, _ := qtest.SetupCourseAssignmentTeacherStudent(t, db)
	benchmark := &qf.GradingBenchmark{
		CourseID:     course.GetID(),
		AssignmentID: assignment.GetID(),
		Heading:      "Benchmark 1",
		Comment:      "comment 1",
	}
	qtest.CreateBenchmark(t, db, benchmark)
	cookie := client.Cookie(t, teacher)

	if _, err := client.DeleteBenchmark(t.Context(), qtest.RequestWithCookie(benchmark, cookie)); err != nil {
		t.Fatalf("DeleteBenchmark failed: %v", err)
	}

	benchmarks := qtest.GetBenchmarks(t, db, assignment.GetID())
	if len(benchmarks) != 0 {
		t.Fatalf("expected 0 benchmarks, got %d", len(benchmarks))
	}
}
