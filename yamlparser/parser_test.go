package yamlparser_test

import (
	"reflect"
	"testing"

	"github.com/autograde/aguis/yamlparser"
)

func TestParseWithInvalidDir(t *testing.T) {
	const dir = "invalid/dir"
	_, err := yamlparser.Parse(dir)
	if err == nil {
		t.Errorf("want no such file or directory error, got nil")
	}
}

func TestParse(t *testing.T) {
	const dir = "testrepos"
	var (
		wantAssignment1 = yamlparser.NewAssignmentRequest{
			AssignmentID: 2,
			Name:         "Lab1",
			Language:     "Java",
			CourseCode:   "DAT100",
			Deadline:     "27-08-2018 12:00",
			Autoapprove:  false,
		}
		wantAssignment2 = yamlparser.NewAssignmentRequest{
			AssignmentID: 1,
			Name:         "Lab1",
			Language:     "GO",
			CourseCode:   "DAT100",
			Deadline:     "27-08-2017 12:00",
			Autoapprove:  false,
		}
	)

	assgns, err := yamlparser.Parse(dir)
	if err != nil {
		t.Fatal(err)
	}

	if len(assgns) < 0 {
		t.Error("have 0 assignments, want 2")
	}

	if !reflect.DeepEqual(assgns[0], wantAssignment1) {
		t.Errorf("have assignment %+v want %+v", assgns[0], wantAssignment1)
	}

	if !reflect.DeepEqual(assgns[1], wantAssignment2) {
		t.Errorf("have assignment %+v want %+v", assgns[1], wantAssignment2)
	}

	// To save assignment to DB
	//for _, assign := range assgns {
	//	course, err := db.GetCourseByCode(assign.CourseCode)
	//	if err == nil {
	//		date, err := web.ParseDate("2-1-2006 15:04", assign.Deadline)
	//		if err == nil {
	//			assignment := &models.Assignment{
	//				Name:         assign.Name,
	//				Language:     assign.Language,
	//				CourseID:     course.ID,
	//				Deadline:     date,
	//				AutoApprove:  assign.Autoapprove,
	//				AssignmentID: assign.AssignmentID,
	//			}
	//			db.CreateAssignment(assignment)
	//		} else {
	//			t.Fatal(err)
	//		}
	//	} else {
	//		t.Fatal(err)
	//	}
	//}
}
