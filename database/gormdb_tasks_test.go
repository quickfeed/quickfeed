package database_test

import (
	"errors"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf/types"
	"google.golang.org/protobuf/testing/protocmp"
	"gorm.io/gorm"
)

func TestGormDBNonExistingTasksForAssignment(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()
	admin := qtest.CreateFakeUser(t, db, uint64(1))
	course := &types.Course{}
	qtest.CreateCourse(t, db, admin, course)

	assignments := []*types.Assignment{
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
	if _, gotError := db.GetTasks(&types.Task{AssignmentID: assignments[0].GetID()}); gotError != wantError {
		t.Errorf("got error '%v' wanted '%v'", gotError, wantError)
	}
}

// TestGormDBSynchronizeAssignmentTasks tests whether SynchronizeAssignmentTasks
// correctly synchronizes tasks in the database, and whether it returns the correct created and updated tasks.
// It loops through possible assignment sequences.
func TestGormDBSynchronizeAssignmentTasks(t *testing.T) {
	tests := map[string]struct {
		foundAssignmentSequence [][]*types.Assignment
	}{
		"Create update delete": {
			foundAssignmentSequence: [][]*types.Assignment{
				{
					{Name: "Lab1", Order: 1, Tasks: []*types.Task{
						{AssignmentOrder: 1, Title: "x", Body: "x", Name: "1"},
						{AssignmentOrder: 1, Title: "x", Body: "x", Name: "2"},
					}},
					{Name: "Lab2", Order: 2, Tasks: []*types.Task{
						{AssignmentOrder: 2, Title: "x", Body: "x", Name: "1"},
						{AssignmentOrder: 2, Title: "x", Body: "x", Name: "2"},
					}},
				},
				{
					{Name: "Lab1", Order: 1, Tasks: []*types.Task{
						{AssignmentOrder: 1, Title: "x", Body: "x", Name: "1"},
						{AssignmentOrder: 1, Title: "x", Body: "y", Name: "2"},
						{AssignmentOrder: 1, Title: "x", Body: "x", Name: "3"},
					}},
					{Name: "Lab2", Order: 2, Tasks: []*types.Task{
						{AssignmentOrder: 2, Title: "y", Body: "x", Name: "1"},
					}},
				},
			},
		},
		"No initial tasks": {
			foundAssignmentSequence: [][]*types.Assignment{
				{
					{Name: "Lab1", Order: 1, Tasks: []*types.Task{}},
					{Name: "Lab2", Order: 2, Tasks: []*types.Task{}},
				},
				{
					{Name: "Lab1", Order: 1, Tasks: []*types.Task{}},
					{Name: "Lab2", Order: 2, Tasks: []*types.Task{}},
				},
				{
					{Name: "Lab1", Order: 1, Tasks: []*types.Task{
						{AssignmentOrder: 1, Title: "x", Body: "x", Name: "1"},
					}},
					{Name: "Lab2", Order: 2, Tasks: []*types.Task{
						{AssignmentOrder: 2, Title: "x", Body: "x", Name: "1"},
					}},
				},
			},
		},
		"Delete and recreate": {
			foundAssignmentSequence: [][]*types.Assignment{
				{
					{Name: "Lab1", Order: 1, Tasks: []*types.Task{
						{AssignmentOrder: 1, Title: "x", Body: "x", Name: "1"},
						{AssignmentOrder: 1, Title: "x", Body: "x", Name: "2"},
					}},
					{Name: "Lab2", Order: 2, Tasks: []*types.Task{
						{AssignmentOrder: 2, Title: "x", Body: "x", Name: "1"},
						{AssignmentOrder: 2, Title: "x", Body: "x", Name: "2"},
					}},
				},
				{
					{Name: "Lab1", Order: 1, Tasks: []*types.Task{}},
					{Name: "Lab2", Order: 2, Tasks: []*types.Task{}},
				},
				{
					{Name: "Lab1", Order: 1, Tasks: []*types.Task{}},
					{Name: "Lab2", Order: 2, Tasks: []*types.Task{}},
				},
				{
					{Name: "Lab1", Order: 1, Tasks: []*types.Task{
						{AssignmentOrder: 1, Title: "x", Body: "x", Name: "1"},
						{AssignmentOrder: 1, Title: "x", Body: "x", Name: "2"},
					}},
					{Name: "Lab2", Order: 2, Tasks: []*types.Task{
						{AssignmentOrder: 2, Title: "x", Body: "x", Name: "1"},
						{AssignmentOrder: 2, Title: "x", Body: "x", Name: "2"},
					}},
					{Name: "Lab3", Order: 3, Tasks: []*types.Task{
						{AssignmentOrder: 3, Title: "x", Body: "x", Name: "1"},
						{AssignmentOrder: 3, Title: "x", Body: "x", Name: "2"},
					}},
				},
			},
		},
		"Mirrored tasks": {
			foundAssignmentSequence: [][]*types.Assignment{
				{
					{Name: "Lab1", Order: 1, Tasks: []*types.Task{
						{AssignmentOrder: 1, Title: "x", Body: "x", Name: "hello_world"},
					}},
					{Name: "Lab2", Order: 2, Tasks: []*types.Task{
						{AssignmentOrder: 2, Title: "x", Body: "x", Name: "hello_world"},
					}},
				},
				{
					{Name: "Lab1", Order: 1, Tasks: []*types.Task{
						{AssignmentOrder: 1, Title: "y", Body: "y", Name: "hello_world"},
					}},
					{Name: "Lab2", Order: 2, Tasks: []*types.Task{
						{AssignmentOrder: 2, Title: "x", Body: "x", Name: "hello_world"},
					}},
					{Name: "Lab3", Order: 3, Tasks: []*types.Task{
						{AssignmentOrder: 3, Title: "x", Body: "x", Name: "hello_world"},
					}},
				},
				{
					{Name: "Lab1", Order: 1, Tasks: []*types.Task{
						{AssignmentOrder: 1, Title: "y", Body: "y", Name: "hello_world"},
					}},
					{Name: "Lab2", Order: 2, Tasks: []*types.Task{
						{AssignmentOrder: 2, Title: "y", Body: "y", Name: "hello_world"},
					}},
					{Name: "Lab3", Order: 3, Tasks: []*types.Task{
						{AssignmentOrder: 3, Title: "y", Body: "y", Name: "not_hello_world"},
					}},
				},
			},
		},
	}

	sortTasksByName := func(tasks []*types.Task) {
		sort.Slice(tasks, func(i, j int) bool {
			return tasks[i].ID < tasks[j].ID
		})
	}
	getTasksFromAssignments := func(assignments []*types.Assignment) map[uint32]map[string]*types.Task {
		taskMap := make(map[uint32]map[string]*types.Task)
		for _, assignment := range assignments {
			temp := make(map[string]*types.Task)
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
			course := &types.Course{}
			qtest.CreateCourse(t, db, admin, course)

			previousTasks := make(map[uint32]map[string]*types.Task)

			for _, foundAssignments := range tt.foundAssignmentSequence {
				wantTasks := []*types.Task{}
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
				gotTasks, err := db.GetTasks(&types.Task{})
				if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
					t.Fatal(err)
				}

				wantCreatedTasks := []*types.Task{}
				wantUpdatedTasks := []*types.Task{}
				for _, wantTask := range wantTasks {
					taskMap, ok := previousTasks[wantTask.GetAssignmentOrder()]
					if !ok {
						previousTasks[wantTask.GetAssignmentOrder()] = make(map[string]*types.Task)
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
