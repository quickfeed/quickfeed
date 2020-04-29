package ag_test

import (
	"fmt"
	"testing"
	"time"

	pb "github.com/autograde/aguis/ag"
)

const (
	layout = "2006-01-02T15:04:05"
	days   = time.Duration(24 * time.Hour)
)

var (
	testNow = time.Now()

	course = &pb.Course{
		ID:       1,
		SlipDays: 5,
		Name:     "opsys",
	}

	a = func(daysFromNow int32) *pb.Assignment {
		return &pb.Assignment{
			CourseID: course.ID,
			Deadline: testNow.Add(time.Duration(daysFromNow) * days).Format(layout),
		}
	}
)

var slipTests = []struct {
	name        string
	labs        []*pb.Assignment
	submissions [][]int32
	remaining   [][]int32
}{
	{"One assignment with deadline two days ago, two submissions same day",
		[]*pb.Assignment{a(-2)},
		[][]int32{{0, 0}},
		[][]int32{{3, 3}},
	},
	{"One assignment with deadline in two days, two submissions same day",
		[]*pb.Assignment{a(2)},
		[][]int32{{0, 0}},
		[][]int32{{5, 5}},
	},
	{"One assignment with deadline in two days, five submissions one day apart",
		[]*pb.Assignment{a(2)},
		[][]int32{{0, 1, 1, 1, 1}},
		[][]int32{{5, 5, 5, 4, 3}},
	},
	{"One assignment with deadline in two days, ten submissions one day apart",
		[]*pb.Assignment{a(2)},
		[][]int32{{0, 1, 1, 1, 1, 1, 1, 1, 1, 1}},
		[][]int32{{5, 5, 5, 4, 3, 2, 1, 0, -1, -2}},
	},
	{"Four assignments with different deadlines, five or more submissions for each assignment",
		[]*pb.Assignment{a(0), a(2), a(5), a(20)},
		[][]int32{ // each number is a submission; if >0 the number is #days since previous submission; they carry over from one line to the next.
			{0, 0, 0, 0, 0},        // 0 ==> five submissions on the same day
			{0, 1, 0, 1, 0},        // 0+2 ==> 1 submission on 0th day, two submissions on day 1, and two submissions on day 2.
			{0, 3, 1, 1, 1},        // 2+6 ==> 1 submission on day 2 (carried over from previous line), one submission on day 2+3, one submission on day 2+3+1, and so on.
			{0, 10, 1, 1, 1, 1, 1}, // 2+6+15 ==> 1 submission on day 2+6, one submission on day 2+6+10, and so on. (23 days total)
		},
		[][]int32{
			{5, 5, 5, 5, 5},        // no slip days used
			{5, 5, 5, 5, 5},        // no slip days used
			{5, 5, 4, 3, 2},        // used 3 slip days
			{2, 2, 2, 2, 1, 0, -1}, // used up all slip days and one more
		},
	},
}

func TestSlipDays(t *testing.T) {
	for _, sd := range slipTests {
		testNow = time.Now()
		enrol := &pb.Enrollment{
			Course:       course,
			CourseID:     course.ID,
			UsedSlipDays: make([]*pb.SlipDays, 0),
		}

		for i := range sd.labs {
			t.Run(fmt.Sprintf("%s#%d", sd.name, i), func(t *testing.T) {
				if len(sd.submissions) != len(sd.remaining) {
					t.Fatalf("faulty test case: len(sd.submissions)=%d != len(sd.remaining)=%d", len(sd.submissions), len(sd.remaining))
				}
				sd.labs[i].ID = uint64(i + 1)
				for j := range sd.submissions[i] {
					if len(sd.submissions[i]) != len(sd.remaining[i]) {
						t.Fatalf("faulty test case: len(sd.submissions[%d])=%d != len(sd.remaining[%d])=%d", i, len(sd.submissions[i]), i, len(sd.remaining[i]))
					}
					subm := &pb.Submission{
						AssignmentID: sd.labs[i].ID,
						Approved:     false,
					}
					// emulate advancing time for this submission
					testNow = testNow.Add(time.Duration(sd.submissions[i][j]) * days)

					// functions to test
					err := enrol.UpdateSlipDays(testNow, sd.labs[i], subm)
					if err != nil {
						t.Fatal(err)
					}
					remaining := enrol.RemainingSlipDays()
					if remaining != sd.remaining[i][j] {
						t.Errorf("UpdateSlipdays(%q, %q, %q, %q) == %d, want %d", testNow.Format(layout), sd.labs[i], subm, enrol, remaining, sd.remaining[i][j])
					}
				}
			})
		}
	}
}

func TestBadDeadlineFormat(t *testing.T) {
	enrol := &pb.Enrollment{
		Course:       course,
		CourseID:     course.ID,
		UsedSlipDays: make([]*pb.SlipDays, 0),
	}
	// lab1's deadline is incorrectly formatted
	lab1 := &pb.Assignment{
		CourseID: course.ID,
		Deadline: "14-Sep-2020",
	}
	lab1.ID = 1
	subm := &pb.Submission{Approved: false, AssignmentID: lab1.ID}
	err := enrol.UpdateSlipDays(testNow, lab1, subm)
	if err == nil {
		t.Errorf("expected parsing error due to incorrect deadline date format")
	}
}

func TestMismatchingAssignmentID(t *testing.T) {
	enrol := &pb.Enrollment{
		Course:       course,
		CourseID:     course.ID,
		UsedSlipDays: make([]*pb.SlipDays, 0),
	}
	// lab1's deadline is incorrectly formatted
	lab1 := &pb.Assignment{
		CourseID: course.ID,
		Deadline: testNow.Add(time.Duration(2) * days).Format(layout),
	}
	lab1.ID = 1
	subm := &pb.Submission{Approved: false, AssignmentID: lab1.ID + 1}
	err := enrol.UpdateSlipDays(testNow, lab1, subm)
	if err == nil {
		t.Errorf("expected invariant violation since (assignment.ID != submission.AssignmentID)")
	}
}

func TestMismatchingCourseID(t *testing.T) {
	enrol := &pb.Enrollment{
		Course:       course,
		CourseID:     course.ID,
		UsedSlipDays: make([]*pb.SlipDays, 0),
	}
	// lab1's deadline is incorrectly formatted
	lab1 := &pb.Assignment{
		CourseID: course.ID + 1,
		Deadline: testNow.Add(time.Duration(2) * days).Format(layout),
	}
	lab1.ID = 1
	subm := &pb.Submission{Approved: false, AssignmentID: lab1.ID}
	err := enrol.UpdateSlipDays(testNow, lab1, subm)
	if err == nil {
		t.Errorf("expected invariant violation since (enrollment.CourseID != assignment.CourseID)")
	}
}

func ExampleEnrollment_GetUsedSlipDays() {
	enrol := &pb.Enrollment{
		Course:       course,
		CourseID:     course.ID,
		UsedSlipDays: make([]*pb.SlipDays, 0),
	}
	// lab1's deadline passed two days ago
	lab1 := a(-2)
	lab1.ID = 1
	subm := &pb.Submission{Approved: false, AssignmentID: lab1.ID}
	fmt.Println(enrol.GetUsedSlipDays())
	err := enrol.UpdateSlipDays(testNow, lab1, subm)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(enrol.GetUsedSlipDays())
	// Output:
	// []
	// [assignmentID:1 usedSlipDays:2 ]
}
