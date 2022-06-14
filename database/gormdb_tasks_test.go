package database_test

import (
	"errors"
	"sort"
	"testing"

	pb "github.com/autograde/quickfeed/ag"
	"github.com/autograde/quickfeed/internal/qtest"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"
	"gorm.io/gorm"
)

func TestGormDBNonExistingTasksForAssignment(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()
	admin := qtest.CreateFakeUser(t, db, uint64(1))
	course := &pb.Course{}
	qtest.CreateCourse(t, db, admin, course)

	assignments := []*pb.Assignment{
		{CourseID: course.GetID(), Name: "Lab1", Order: 1},
		{CourseID: course.GetID(), Name: "Lab2", Order: 2},
	}

	for _, assignment := range assignments {
		if err := db.CreateAssignment(assignment); err != nil {
			t.Error(err)
		}
	}

	assignments, err := db.GetAssignmentsByCourse(course.GetID(), false)
	if err != nil {
		t.Error(err)
	}
	if len(assignments) == 0 {
		t.Errorf("len(assignments) == %d, expected 2", len(assignments))
	}

	wantError := gorm.ErrRecordNotFound
	if _, gotError := db.GetTasks(&pb.Task{AssignmentID: assignments[0].GetID()}); gotError != wantError {
		t.Errorf("got error '%v' wanted '%v'", gotError, wantError)
	}
}

// TestGormDBSynchronizeAssignmentTasks tests whether SynchronizeAssignmentTasks
// correctly synchronizes tasks in the database, and whether it returns the correct created and updated tasks.
// It loops through possible assignment sequences.
func TestGormDBSynchronizeAssignmentTasks(t *testing.T) {
	tests := map[string]struct {
		foundAssignmentSequence [][]*pb.Assignment
	}{
		"Create update delete": {
			foundAssignmentSequence: [][]*pb.Assignment{
				{
					{Name: "Lab1", Order: 1, Tasks: []*pb.Task{
						{AssignmentOrder: 1, Title: "x", Body: "x", Name: "Lab1/1"},
						{AssignmentOrder: 1, Title: "x", Body: "x", Name: "Lab1/2"},
					}},
					{Name: "Lab2", Order: 2, Tasks: []*pb.Task{
						{AssignmentOrder: 2, Title: "x", Body: "x", Name: "Lab2/1"},
						{AssignmentOrder: 2, Title: "x", Body: "x", Name: "Lab2/2"},
					}},
				},
				{
					{Name: "Lab1", Order: 1, Tasks: []*pb.Task{
						{AssignmentOrder: 1, Title: "x", Body: "x", Name: "Lab1/1"},
						{AssignmentOrder: 1, Title: "x", Body: "y", Name: "Lab1/2"},
						{AssignmentOrder: 1, Title: "x", Body: "x", Name: "Lab1/3"},
					}},
					{Name: "Lab2", Order: 2, Tasks: []*pb.Task{
						{AssignmentOrder: 2, Title: "y", Body: "x", Name: "Lab2/1"},
					}},
				},
			},
		},
		"No initial tasks": {
			foundAssignmentSequence: [][]*pb.Assignment{
				{
					{Name: "Lab1", Order: 1, Tasks: []*pb.Task{}},
					{Name: "Lab2", Order: 2, Tasks: []*pb.Task{}},
				},
				{
					{Name: "Lab1", Order: 1, Tasks: []*pb.Task{}},
					{Name: "Lab2", Order: 2, Tasks: []*pb.Task{}},
				},
				{
					{Name: "Lab1", Order: 1, Tasks: []*pb.Task{
						{AssignmentOrder: 1, Title: "x", Body: "x", Name: "Lab1/1"},
					}},
					{Name: "Lab2", Order: 2, Tasks: []*pb.Task{
						{AssignmentOrder: 2, Title: "x", Body: "x", Name: "Lab2/1"},
					}},
				},
			},
		},
		"Delete and recreate": {
			foundAssignmentSequence: [][]*pb.Assignment{
				{
					{Name: "Lab1", Order: 1, Tasks: []*pb.Task{
						{AssignmentOrder: 1, Title: "x", Body: "x", Name: "Lab1/1"},
						{AssignmentOrder: 1, Title: "x", Body: "x", Name: "Lab1/2"},
					}},
					{Name: "Lab2", Order: 2, Tasks: []*pb.Task{
						{AssignmentOrder: 2, Title: "x", Body: "x", Name: "Lab2/1"},
						{AssignmentOrder: 2, Title: "x", Body: "x", Name: "Lab2/2"},
					}},
				},
				{
					{Name: "Lab1", Order: 1, Tasks: []*pb.Task{}},
					{Name: "Lab2", Order: 2, Tasks: []*pb.Task{}},
				},
				{
					{Name: "Lab1", Order: 1, Tasks: []*pb.Task{}},
					{Name: "Lab2", Order: 2, Tasks: []*pb.Task{}},
				},
				{
					{Name: "Lab1", Order: 1, Tasks: []*pb.Task{
						{AssignmentOrder: 1, Title: "x", Body: "x", Name: "Lab1/1"},
						{AssignmentOrder: 1, Title: "x", Body: "x", Name: "Lab1/2"},
					}},
					{Name: "Lab2", Order: 2, Tasks: []*pb.Task{
						{AssignmentOrder: 2, Title: "x", Body: "x", Name: "Lab2/1"},
						{AssignmentOrder: 2, Title: "x", Body: "x", Name: "Lab2/2"},
					}},
					{Name: "Lab3", Order: 3, Tasks: []*pb.Task{
						{AssignmentOrder: 3, Title: "x", Body: "x", Name: "Lab3/1"},
						{AssignmentOrder: 3, Title: "x", Body: "x", Name: "Lab3/2"},
					}},
				},
			},
		},
	}

	sortTasksByName := func(tasks []*pb.Task) {
		sort.Slice(tasks, func(i, j int) bool {
			return tasks[i].Name < tasks[j].Name
		})
	}
	getTasksFromAssignments := func(assignments []*pb.Assignment) map[uint32]map[string]*pb.Task {
		taskMap := make(map[uint32]map[string]*pb.Task)
		for _, assignment := range assignments {
			temp := make(map[string]*pb.Task)
			for _, task := range assignment.Tasks {
				temp[task.Name] = task
			}
			taskMap[assignment.Order] = temp
		}
		return taskMap
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			db, cleanup := qtest.TestDB(t)
			defer cleanup()
			admin := qtest.CreateFakeUser(t, db, 1)
			course := &pb.Course{}
			qtest.CreateCourse(t, db, admin, course)

			previousTasks := make(map[string]*pb.Task)

			for _, foundAssignments := range tt.foundAssignmentSequence {
				wantTasks := []*pb.Task{}
				for _, assignment := range foundAssignments {
					assignment.CourseID = course.GetID()
					if err := db.CreateAssignment(assignment); err != nil {
						t.Error(err)
					}
					wantTasks = append(wantTasks, assignment.Tasks...)
				}
				gotCreatedTasks, gotUpdatedTasks, err := db.SynchronizeAssignmentTasks(course, getTasksFromAssignments(foundAssignments))
				if err != nil {
					t.Error(err)
				}
				gotTasks, err := db.GetTasks(&pb.Task{})
				if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
					t.Fatal(err)
				}

				wantCreatedTasks := []*pb.Task{}
				wantUpdatedTasks := []*pb.Task{}
				for _, wantTask := range wantTasks {
					task, ok := previousTasks[wantTask.GetName()]
					if ok {
						wantTask.ID = task.GetID()
						wantTask.AssignmentID = task.GetAssignmentID()
						if task.HasChanged(wantTask) {
							wantUpdatedTasks = append(wantUpdatedTasks, wantTask)
						}
					} else {
						wantCreatedTasks = append(wantCreatedTasks, wantTask)
					}
					delete(previousTasks, wantTask.GetName())
				}
				for name, deletedTask := range previousTasks {
					deletedTask.MarkDeleted()
					wantUpdatedTasks = append(wantUpdatedTasks, deletedTask)
					delete(previousTasks, name)
				}

				sortTasksByName(wantTasks)
				sortTasksByName(gotTasks)
				if diff := cmp.Diff(wantTasks, gotTasks, protocmp.Transform()); diff != "" {
					t.Errorf("Synchronization mismatch (-wantTasks, +gotTasks):\n%s", diff)
				}

				sortTasksByName(wantCreatedTasks)
				// gotCreatedTasks is already sorted by SynchronizeAssignmentTasks
				if diff := cmp.Diff(wantCreatedTasks, gotCreatedTasks, protocmp.Transform()); diff != "" {
					t.Errorf("Synchronization return mismatch (-wantCreatedTasks, +gotCreatedTasks):\n%s", diff)
				}

				sortTasksByName(wantUpdatedTasks)
				sortTasksByName(gotUpdatedTasks)
				if diff := cmp.Diff(wantUpdatedTasks, gotUpdatedTasks, protocmp.Transform()); diff != "" {
					t.Errorf("Synchronization return mismatch (-wantUpdatedTasks, +gotUpdatedTasks):\n%s", diff)
				}

				for _, task := range wantTasks {
					previousTasks[task.GetName()] = task
				}
			}
		})
	}
}

// TODO(Espeland): These tests don't really do anything atm
func TestCreatePullRequest(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	wantPullRequest := &pb.PullRequest{
		ScmRepositoryID: 1,
		TaskID:          1,
		IssueID:         1,
		UserID:          1,
		SourceBranch:    "A",
		Number:          1,
	}
	if err := db.CreatePullRequest(wantPullRequest); err != nil {
		t.Fatal(err)
	}
	gotPullRequest, err := db.GetPullRequest(&pb.PullRequest{})
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(wantPullRequest, gotPullRequest, protocmp.Transform()); diff != "" {
		t.Errorf("CreatePullRequest mismatch (-wantPullRequest, +gotPullRequest):\n%s", diff)
	}
}

func TestHandleMergingPR(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	if err := db.CreateIssues([]*pb.Issue{{IssueNumber: 10}}); err != nil {
		t.Fatal(err)
	}

	pullRequest := &pb.PullRequest{
		ScmRepositoryID: 1,
		TaskID:          1,
		IssueID:         1,
		UserID:          1,
		SourceBranch:    "A",
		Number:          1,
	}
	if err := db.CreatePullRequest(pullRequest); err != nil {
		t.Fatal(err)
	}

	if err := db.HandleMergingPR(pullRequest); err != nil {
		t.Fatal(err)
	}
}
