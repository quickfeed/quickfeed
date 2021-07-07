package ag_test

import (
	"testing"

	pb "github.com/autograde/quickfeed/ag"
)

func TestIsManual(t *testing.T) {
	assignmentWithoutBenchmarks := &pb.Assignment{
		ID:       1,
		CourseID: 24,
	}
	got := assignmentWithoutBenchmarks.IsManual()
	expected := false
	if got != expected {
		t.Errorf("IsManual()=%t, expected %t", got, expected)
	}

	got = assignmentWithoutBenchmarks.HasTests()
	expected = true
	if got != expected {
		t.Errorf("HasTests()=%t, expected %t", got, expected)
	}

	assignmentWithBenchmarks := &pb.Assignment{
		ID:                1,
		CourseID:          24,
		GradingBenchmarks: []*pb.GradingBenchmark{{}},
	}
	got = assignmentWithBenchmarks.IsManual()
	expected = true
	if got != expected {
		t.Errorf("IsManual()=%t, expected %t", got, expected)
	}

	got = assignmentWithBenchmarks.HasTests()
	expected = false
	if got != expected {
		t.Errorf("HasTests()=%t, expected %t", got, expected)
	}
}
