package web_test

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/scm"
	"github.com/quickfeed/quickfeed/web"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestCreateCriterion(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()
	client := web.MockClient(t, db, scm.WithMockOrgs(), nil)
	_, _, assignment := qtest.SetupCourseAssignment(t, db)
	qtest.CreateBenchmark(t, db, &qf.GradingBenchmark{AssignmentID: assignment.GetID()})

	wantCriterion := &qf.GradingCriterion{
		ID:          1,
		Description: "A great criterion",
		Points:      10,
		Comment:     "comment 1",
	}

	gotCriterion, err := client.CreateCriterion(context.Background(), &connect.Request[qf.GradingCriterion]{Msg: &qf.GradingCriterion{
		Description: "A great criterion",
		Points:      10,
		Comment:     "comment 1",
	}})
	if err != nil {
		t.Fatalf("CreateCriterion failed: %v", err)
	}

	qtest.Diff(t, "CreateCriterion() mismatch", gotCriterion.Msg, wantCriterion, protocmp.Transform())
}

func TestUpdateCriterion(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()
	client := web.MockClient(t, db, scm.WithMockOrgs(), nil)
	_, _, assignment := qtest.SetupCourseAssignment(t, db)
	qtest.CreateBenchmark(t, db, &qf.GradingBenchmark{AssignmentID: assignment.GetID()})
	qtest.CreateCriterion(t, db, &qf.GradingCriterion{BenchmarkID: 1})

	wantCriterion := &qf.GradingCriterion{
		ID:          1,
		BenchmarkID: 1,
		Points:      10,
		Description: "A great criterion",
		Comment:     "Stupid comment",
	}

	if _, err := client.UpdateCriterion(context.Background(), &connect.Request[qf.GradingCriterion]{Msg: wantCriterion}); err != nil {
		t.Fatal(err)
	}
	benchmarks := qtest.GetBenchmarks(t, db, 1)
	if len(benchmarks) != 1 {
		t.Fatalf("expected 0 benchmarks, got %d", len(benchmarks))
	}
	criteria := benchmarks[0].GetCriteria()
	if len(criteria) != 1 {
		t.Fatalf("expected 1 criteria, got %d", len(criteria))
	}
	gotCriterion := criteria[0]
	qtest.Diff(t, "CreateBenchmark mismatch", gotCriterion, wantCriterion, protocmp.Transform())
}

func TestDeleteCriterion(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()
	client := web.MockClient(t, db, scm.WithMockOrgs(), nil)
	_, _, assignment := qtest.SetupCourseAssignment(t, db)
	qtest.CreateBenchmark(t, db, &qf.GradingBenchmark{AssignmentID: assignment.GetID()})
	qtest.CreateCriterion(t, db, &qf.GradingCriterion{BenchmarkID: 1})

	if _, err := client.DeleteCriterion(context.Background(), &connect.Request[qf.GradingCriterion]{Msg: &qf.GradingCriterion{
		ID: 1,
	}}); err != nil {
		t.Fatalf("DeleteBenchmark failed: %v", err)
	}
	benchmarks := qtest.GetBenchmarks(t, db, 1)
	if len(benchmarks) != 1 {
		t.Fatalf("expected 0 benchmarks, got %d", len(benchmarks))
	}
	if len(benchmarks[0].GetCriteria()) != 0 {
		t.Fatalf("expected 0 criteria, got %d", len(benchmarks[0].GetCriteria()))
	}
}
