package database_test

import (
	"testing"

	pb "github.com/autograde/quickfeed/ag"
	"github.com/autograde/quickfeed/database"
	"github.com/autograde/quickfeed/internal/qtest"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"
	"gorm.io/gorm"
)

// Helper function
func getTasksFromAssignments(assignments []*pb.Assignment) map[uint32]map[string]*pb.Task {
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

// createCourseWithAssignments creates a course with two assignments.
func createCourseWithAssignments(t *testing.T, db database.Database) *pb.Course {
	t.Helper()
	admin := qtest.CreateFakeUser(t, db, uint64(1))
	qtest.CreateCourse(t, db, admin, &pb.Course{})

	course, err := db.GetCourse(1, false)
	if err != nil {
		t.Fatal(err)
	}
	assignments := []*pb.Assignment{
		{CourseID: course.GetID(), Name: "Lab1", Order: 1},
		{CourseID: course.GetID(), Name: "Lab2", Order: 2},
	}

	for _, assignment := range assignments {
		if err := db.CreateAssignment(assignment); err != nil {
			t.Error(err)
		}
	}
	return course
}

// initialAssignments simulates getting assignments parsed from tests repository.
func initialAssignments() ([]*pb.Assignment, []*pb.Task) {
	foundTasks := []*pb.Task{
		{AssignmentOrder: 1, Title: "Lab1, task1", Body: "Description of task1 in lab1", Name: "Lab1/1"},
		{AssignmentOrder: 1, Title: "Lab1, task2", Body: "Description of task2 in lab1", Name: "Lab1/2"},
		{AssignmentOrder: 2, Title: "Lab2, task1", Body: "Description of task1 in lab2", Name: "Lab2/1"},
		{AssignmentOrder: 2, Title: "Lab2, task2", Body: "Description of task2 in lab2", Name: "Lab2/2"},
	}

	foundAssignments := []*pb.Assignment{
		{Name: "Lab1", Order: 1, Tasks: foundTasks[:2]},
		{Name: "Lab2", Order: 2, Tasks: foundTasks[2:]},
	}
	return foundAssignments, foundTasks
}

func TestGormDBNonExistingTasksForAssignment(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()
	course := createCourseWithAssignments(t, db)

	assignments, err := db.GetAssignmentsByCourse(course.GetID(), false)
	if err != nil || len(assignments) == 0 {
		t.Fatal(err)
	}

	wantError := gorm.ErrRecordNotFound
	if _, gotError := db.GetTasks(&pb.Task{AssignmentID: assignments[0].GetID()}); gotError != wantError {
		t.Errorf("got error '%v' wanted '%v'", gotError, wantError)
	}
}

// TestSynchronizeTasks tests whether tasks are correctly updated in the database
func TestGormDBSynchronizeAssignmentTasks(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()
	course := createCourseWithAssignments(t, db)

	foundAssignments1, foundTasks1 := initialAssignments()

	// Should create a new database-record for each task in foundTasks
	if _, _, err := db.SynchronizeAssignmentTasks(course, getTasksFromAssignments(foundAssignments1)); err != nil {
		t.Fatal(err)
	}

	wantTasks1 := foundTasks1
	gotTasks1, err := db.GetTasks(&pb.Task{})
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(wantTasks1, gotTasks1, protocmp.Transform()); diff != "" {
		t.Errorf("Synchronization mismatch (-wantTasks1, +gotTasks1):\n%s", diff)
	}

	// -------------------------------------------------------------------------- //
	// Testing adding one new task and updating another
	// -------------------------------------------------------------------------- //
	foundAssignments2, foundTasks2 := initialAssignments()

	newTask := &pb.Task{
		AssignmentOrder: 2,
		Title:           "Lab2, task3",
		Body:            "Description of task3 in lab2",
		Name:            "Lab2/3",
	}
	foundTasks2 = append(foundTasks2, newTask)
	foundAssignments2[1].Tasks = append(foundAssignments2[1].Tasks, newTask)
	foundAssignments2[0].Tasks[0].Body = "New body for lab1 task1"

	if _, _, err = db.SynchronizeAssignmentTasks(course, getTasksFromAssignments(foundAssignments2)); err != nil {
		t.Fatal(err)
	}

	wantTasks2 := foundTasks1
	wantTasks2 = append(wantTasks2, foundTasks2[len(foundTasks2)-1])
	wantTasks2[0].Body = "New body for lab1 task1"

	gotTasks2, err := db.GetTasks(&pb.Task{})
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(wantTasks2, gotTasks2, protocmp.Transform()); diff != "" {
		t.Errorf("Synchronization mismatch (-wantTasks2, +gotTasks2):\n%s", diff)
	}

	// -------------------------------------------------------------------------- //
	// Test adding a new task to database not represented by tasks supplied to
	// SynchronizeAssignmentTasks, then finding the same tasks as in previous test.
	// -------------------------------------------------------------------------- //
	err = db.CreateTasks([]*pb.Task{
		{
			AssignmentID:    1,
			AssignmentOrder: 1,
			Title:           "Title title",
			Body:            "This task should not exists in db",
			Name:            "Fake name",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err = db.SynchronizeAssignmentTasks(course, getTasksFromAssignments(foundAssignments2)); err != nil {
		t.Fatal(err)
	}

	wantTasks3 := wantTasks2
	gotTasks3, err := db.GetTasks(&pb.Task{})
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(wantTasks3, gotTasks3, protocmp.Transform()); diff != "" {
		t.Errorf("Synchronization mismatch (-wantTasks3, +gotTasks3):\n%s", diff)
	}
}

// TestSynchronizeAssignmentTasksReturn tests if SynchronizeAssignmentTasks returns correct values
func TestGormDBReturnSynchronizeAssignmentTasks(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()
	course := createCourseWithAssignments(t, db)

	foundAssignments1, foundTasks1 := initialAssignments()

	// Creating four new tasks
	gotCreatedTasks, gotUpdatedTasks, err := db.SynchronizeAssignmentTasks(course, getTasksFromAssignments(foundAssignments1))
	if err != nil {
		t.Fatal(err)
	}

	wantUpdatedTasks := []*pb.Task{}
	wantCreatedTasks := foundTasks1

	if diff := cmp.Diff(wantCreatedTasks, gotCreatedTasks, protocmp.Transform()); diff != "" {
		t.Errorf("SynchronizeAssignmentTasks return mismatch (-wantCreatedTasks, +gotCreatedTasks):\n%s", diff)
	}
	if diff := cmp.Diff(wantUpdatedTasks, gotUpdatedTasks, protocmp.Transform()); diff != "" {
		t.Errorf("SynchronizeAssignmentTasks return mismatch (-wantUpdatedTasks, +gotUpdatedTasks):\n%s", diff)
	}

	// -------------------------------------------------------------------------- //
	// Creating three new tasks, updating two existing and deleting two existing
	// -------------------------------------------------------------------------- //
	foundAssignments2, foundTasks2 := initialAssignments()

	newTasks := []*pb.Task{
		{
			AssignmentOrder: 1,
			Title:           "New Task 1",
			Body:            "Description of New Task 1",
			Name:            "Lab1/3",
		},
		{
			AssignmentOrder: 1,
			Title:           "New Task 2",
			Body:            "Description of New Task 2",
			Name:            "Lab1/4",
		},
		{
			AssignmentOrder: 2,
			Title:           "New Task 1 in another assignment",
			Body:            "Description of New Task 1 in another assignment",
			Name:            "Lab2/3",
		},
	}

	foundTasks2 = append(foundTasks2, newTasks...)
	for _, assignment := range foundAssignments2 {
		tasks := []*pb.Task{}
		for _, task := range foundTasks2 {
			if assignment.Order == task.AssignmentOrder {
				tasks = append(tasks, task)
			}
		}
		assignment.Tasks = tasks
	}

	foundAssignments2[0].Tasks = append(foundAssignments2[0].Tasks[:1], foundAssignments2[0].Tasks[2:]...)
	foundAssignments2[1].Tasks = append(foundAssignments2[1].Tasks[:1], foundAssignments2[1].Tasks[2:]...)

	foundAssignments2[0].Tasks[0].Title = "New title for task 1 assignment 1"
	foundAssignments2[1].Tasks[0].Title = "New title for task 1 assignment 2"

	gotCreatedTasks, gotUpdatedTasks, err = db.SynchronizeAssignmentTasks(course, getTasksFromAssignments(foundAssignments2))
	if err != nil {
		t.Fatal(err)
	}
	wantCreatedTasks = newTasks

	foundAssignments1[0].Tasks[0].Title = "New title for task 1 assignment 1"
	foundAssignments1[0].Tasks[1].MarkDeleted()
	foundAssignments1[1].Tasks[0].Title = "New title for task 1 assignment 2"
	foundAssignments1[1].Tasks[1].MarkDeleted()
	wantUpdatedTasks = append(wantUpdatedTasks,
		foundAssignments1[0].Tasks[0], // updated task 1 of assignment 1
		foundAssignments1[0].Tasks[1], // deleted task 2 of assignment 1
		foundAssignments1[1].Tasks[0], // updated task 1 of assignment 2
		foundAssignments1[1].Tasks[1], // deleted task 2 of assignment 2
	)

	if diff := cmp.Diff(wantCreatedTasks, gotCreatedTasks, protocmp.Transform()); diff != "" {
		t.Errorf("SynchronizeAssignmentTasks return mismatch (-wantCreatedTasks, +gotCreatedTasks):\n%s", diff)
	}
	if diff := cmp.Diff(wantUpdatedTasks, gotUpdatedTasks, protocmp.Transform()); diff != "" {
		t.Errorf("SynchronizeAssignmentTasks return mismatch (-wantUpdatedTasks, +gotUpdatedTasks):\n%s", diff)
	}
}
