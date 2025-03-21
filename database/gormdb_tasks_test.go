package database_test

import (
	"errors"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
	"google.golang.org/protobuf/testing/protocmp"
	"gorm.io/gorm"
)

func TestGormDBNonExistingTasksForAssignment(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()
	admin := qtest.CreateFakeUser(t, db)
	course := &qf.Course{}
	qtest.CreateCourse(t, db, admin, course)

	assignments := []*qf.Assignment{
		{CourseID: course.GetID(), Name: "Lab1", Order: 1},
		{CourseID: course.GetID(), Name: "Lab2", Order: 2},
	}

	for _, assignment := range assignments {
		if err := db.CreateAssignment(assignment); err != nil {
			t.Error(err)
		}
	}

	assignments, err := db.GetAssignmentsByCourse(course.GetID())
	if err != nil {
		t.Error(err)
	}
	if len(assignments) == 0 {
		t.Errorf("len(assignments) == %d, expected 2", len(assignments))
	}

	wantError := gorm.ErrRecordNotFound
	if _, gotError := db.GetTasks(&qf.Task{AssignmentID: assignments[0].GetID()}); gotError != wantError {
		t.Errorf("got error '%v' wanted '%v'", gotError, wantError)
	}
}

// TestGormDBSynchronizeAssignmentTasks tests whether SynchronizeAssignmentTasks
// correctly synchronizes tasks in the database, and whether it returns the correct created and updated tasks.
// It loops through possible assignment sequences.
func TestGormDBSynchronizeAssignmentTasks(t *testing.T) {
	tests := map[string]struct {
		foundAssignmentSequence [][]*qf.Assignment
	}{
		"Create update delete": {
			foundAssignmentSequence: [][]*qf.Assignment{
				{
					{Name: "Lab1", Order: 1, Tasks: []*qf.Task{
						{AssignmentOrder: 1, Title: "x", Body: "x", Name: "1"},
						{AssignmentOrder: 1, Title: "x", Body: "x", Name: "2"},
					}},
					{Name: "Lab2", Order: 2, Tasks: []*qf.Task{
						{AssignmentOrder: 2, Title: "x", Body: "x", Name: "1"},
						{AssignmentOrder: 2, Title: "x", Body: "x", Name: "2"},
					}},
				},
				{
					{Name: "Lab1", Order: 1, Tasks: []*qf.Task{
						{AssignmentOrder: 1, Title: "x", Body: "x", Name: "1"},
						{AssignmentOrder: 1, Title: "x", Body: "y", Name: "2"},
						{AssignmentOrder: 1, Title: "x", Body: "x", Name: "3"},
					}},
					{Name: "Lab2", Order: 2, Tasks: []*qf.Task{
						{AssignmentOrder: 2, Title: "y", Body: "x", Name: "1"},
					}},
				},
			},
		},
		"No initial tasks": {
			foundAssignmentSequence: [][]*qf.Assignment{
				{
					{Name: "Lab1", Order: 1, Tasks: []*qf.Task{}},
					{Name: "Lab2", Order: 2, Tasks: []*qf.Task{}},
				},
				{
					{Name: "Lab1", Order: 1, Tasks: []*qf.Task{}},
					{Name: "Lab2", Order: 2, Tasks: []*qf.Task{}},
				},
				{
					{Name: "Lab1", Order: 1, Tasks: []*qf.Task{
						{AssignmentOrder: 1, Title: "x", Body: "x", Name: "1"},
					}},
					{Name: "Lab2", Order: 2, Tasks: []*qf.Task{
						{AssignmentOrder: 2, Title: "x", Body: "x", Name: "1"},
					}},
				},
			},
		},
		"Delete and recreate": {
			foundAssignmentSequence: [][]*qf.Assignment{
				{
					{Name: "Lab1", Order: 1, Tasks: []*qf.Task{
						{AssignmentOrder: 1, Title: "x", Body: "x", Name: "1"},
						{AssignmentOrder: 1, Title: "x", Body: "x", Name: "2"},
					}},
					{Name: "Lab2", Order: 2, Tasks: []*qf.Task{
						{AssignmentOrder: 2, Title: "x", Body: "x", Name: "1"},
						{AssignmentOrder: 2, Title: "x", Body: "x", Name: "2"},
					}},
				},
				{
					{Name: "Lab1", Order: 1, Tasks: []*qf.Task{}},
					{Name: "Lab2", Order: 2, Tasks: []*qf.Task{}},
				},
				{
					{Name: "Lab1", Order: 1, Tasks: []*qf.Task{}},
					{Name: "Lab2", Order: 2, Tasks: []*qf.Task{}},
				},
				{
					{Name: "Lab1", Order: 1, Tasks: []*qf.Task{
						{AssignmentOrder: 1, Title: "x", Body: "x", Name: "1"},
						{AssignmentOrder: 1, Title: "x", Body: "x", Name: "2"},
					}},
					{Name: "Lab2", Order: 2, Tasks: []*qf.Task{
						{AssignmentOrder: 2, Title: "x", Body: "x", Name: "1"},
						{AssignmentOrder: 2, Title: "x", Body: "x", Name: "2"},
					}},
					{Name: "Lab3", Order: 3, Tasks: []*qf.Task{
						{AssignmentOrder: 3, Title: "x", Body: "x", Name: "1"},
						{AssignmentOrder: 3, Title: "x", Body: "x", Name: "2"},
					}},
				},
			},
		},
		"Mirrored tasks": {
			foundAssignmentSequence: [][]*qf.Assignment{
				{
					{Name: "Lab1", Order: 1, Tasks: []*qf.Task{
						{AssignmentOrder: 1, Title: "x", Body: "x", Name: "hello_world"},
					}},
					{Name: "Lab2", Order: 2, Tasks: []*qf.Task{
						{AssignmentOrder: 2, Title: "x", Body: "x", Name: "hello_world"},
					}},
				},
				{
					{Name: "Lab1", Order: 1, Tasks: []*qf.Task{
						{AssignmentOrder: 1, Title: "y", Body: "y", Name: "hello_world"},
					}},
					{Name: "Lab2", Order: 2, Tasks: []*qf.Task{
						{AssignmentOrder: 2, Title: "x", Body: "x", Name: "hello_world"},
					}},
					{Name: "Lab3", Order: 3, Tasks: []*qf.Task{
						{AssignmentOrder: 3, Title: "x", Body: "x", Name: "hello_world"},
					}},
				},
				{
					{Name: "Lab1", Order: 1, Tasks: []*qf.Task{
						{AssignmentOrder: 1, Title: "y", Body: "y", Name: "hello_world"},
					}},
					{Name: "Lab2", Order: 2, Tasks: []*qf.Task{
						{AssignmentOrder: 2, Title: "y", Body: "y", Name: "hello_world"},
					}},
					{Name: "Lab3", Order: 3, Tasks: []*qf.Task{
						{AssignmentOrder: 3, Title: "y", Body: "y", Name: "not_hello_world"},
					}},
				},
			},
		},
	}

	sortTasksByName := func(tasks []*qf.Task) {
		sort.Slice(tasks, func(i, j int) bool {
			return tasks[i].GetID() < tasks[j].GetID()
		})
	}
	getTasksFromAssignments := func(assignments []*qf.Assignment) map[uint32]map[string]*qf.Task {
		taskMap := make(map[uint32]map[string]*qf.Task)
		for _, assignment := range assignments {
			temp := make(map[string]*qf.Task)
			for _, task := range assignment.GetTasks() {
				temp[task.GetName()] = task
			}
			taskMap[assignment.GetOrder()] = temp
		}
		return taskMap
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			db, cleanup := qtest.TestDB(t)
			defer cleanup()
			admin := qtest.CreateFakeUser(t, db)
			course := &qf.Course{}
			qtest.CreateCourse(t, db, admin, course)

			previousTasks := make(map[uint32]map[string]*qf.Task)

			for _, foundAssignments := range tt.foundAssignmentSequence {
				var wantTasks []*qf.Task
				for _, assignment := range foundAssignments {
					assignment.CourseID = course.GetID()
					if err := db.CreateAssignment(assignment); err != nil {
						t.Error(err)
					}
					wantTasks = append(wantTasks, assignment.GetTasks()...)
				}
				gotCreatedTasks, gotUpdatedTasks, err := db.SynchronizeAssignmentTasks(course, getTasksFromAssignments(foundAssignments))
				if err != nil {
					t.Error(err)
				}
				gotTasks, err := db.GetTasks(&qf.Task{})
				if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
					t.Fatal(err)
				}

				var wantCreatedTasks []*qf.Task
				var wantUpdatedTasks []*qf.Task
				for _, wantTask := range wantTasks {
					taskMap, ok := previousTasks[wantTask.GetAssignmentOrder()]
					if !ok {
						previousTasks[wantTask.GetAssignmentOrder()] = make(map[string]*qf.Task)
					}
					task, ok := taskMap[wantTask.GetName()]
					if ok {
						// wantTask in previousTasks map; it must have been updated
						wantTask.ID = task.GetID()
						wantTask.AssignmentID = task.GetAssignmentID()
						if task.HasChanged(wantTask) {
							wantUpdatedTasks = append(wantUpdatedTasks, wantTask)
						}
					} else {
						// wantTask not in previousTasks map; it must have been created
						wantCreatedTasks = append(wantCreatedTasks, wantTask)
					}
					delete(taskMap, wantTask.GetName())
				}

				// All tasks remaining in previousTasks must have been deleted.
				for _, taskMap := range previousTasks {
					for name, deletedTask := range taskMap {
						deletedTask.MarkDeleted()
						wantUpdatedTasks = append(wantUpdatedTasks, deletedTask)
						delete(taskMap, name)
					}
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
					previousTasks[task.GetAssignmentOrder()][task.GetName()] = task
				}
			}
		})
	}
}
