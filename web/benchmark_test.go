package web_test

import (
	"context"
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
	defer cleanup()
	client := web.MockClient(t, db, scm.WithMockOrgs(), nil)
	_, _, assignment := qtest.SetupCourseAssignment(t, db)

	tests := []struct {
		name      string
		benchmark *qf.GradingBenchmark
		wantErr   error
	}{
		{
			name: "Valid benchmark",
			benchmark: &qf.GradingBenchmark{
				AssignmentID: assignment.GetID(),
				Heading:      "Benchmark 1",
				Comment:      "comment 1",
			},
		},
		{
			name: "Non-existing assignment",
			benchmark: &qf.GradingBenchmark{
				AssignmentID: 111,
				Heading:      "This is a test benchmark",
				Comment:      "",
			},
			wantErr: connect.NewError(connect.CodeInvalidArgument, errors.New("failed to create benchmark")),
		},
	}
	for i, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gotBenchmark, err := client.CreateBenchmark(context.Background(), &connect.Request[qf.GradingBenchmark]{Msg: test.benchmark})
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
	defer cleanup()
	client := web.MockClient(t, db, scm.WithMockOrgs(), nil)
	_, _, assignment := qtest.SetupCourseAssignment(t, db)
	qtest.CreateBenchmark(t, db, &qf.GradingBenchmark{AssignmentID: assignment.GetID()})

	wantBenchmark := &qf.GradingBenchmark{
		ID:           1,
		AssignmentID: assignment.GetID(),
		Heading:      "Updated Benchmark",
		Comment:      "Updated comment",
	}

	if _, err := client.UpdateBenchmark(context.Background(), &connect.Request[qf.GradingBenchmark]{Msg: wantBenchmark}); err != nil {
		t.Fatal(err)
	}
	benchmarks := qtest.GetBenchmarks(t, db, 1)
	if len(benchmarks) != 1 {
		t.Fatalf("expected 0 benchmarks, got %d", len(benchmarks))
	}
	gotBenchmark := benchmarks[0]
	qtest.Diff(t, "CreateBenchmark mismatch", gotBenchmark, wantBenchmark, protocmp.Transform())
}

func TestDeleteBenchmark(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()
	client := web.MockClient(t, db, scm.WithMockOrgs(), nil)
	_, _, assignment := qtest.SetupCourseAssignment(t, db)
	benchmark := &qf.GradingBenchmark{
		AssignmentID: assignment.GetID(),
		Heading:      "Benchmark 1",
		Comment:      "comment 1",
	}
	qtest.CreateBenchmark(t, db, benchmark)

	if _, err := client.DeleteBenchmark(context.Background(), &connect.Request[qf.GradingBenchmark]{Msg: &qf.GradingBenchmark{
		ID: benchmark.GetID(),
	}}); err != nil {
		t.Fatalf("DeleteBenchmark failed: %v", err)
	}

	benchmarks := qtest.GetBenchmarks(t, db, benchmark.GetID())
	if len(benchmarks) != 0 {
		t.Fatalf("expected 0 benchmarks, got %d", len(benchmarks))
	}
}
