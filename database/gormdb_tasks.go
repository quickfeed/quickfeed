package database

import (
	"errors"
	"fmt"
	"sort"

	pb "github.com/autograde/quickfeed/ag"
	"gorm.io/gorm"
)

// TODO(Meling): Methods such as GetTasks and CreateTasks are not necessary, except for in tests. They therefore need to be a part of the interface, even though they are not actually used.
// Is there a better way of handeling this?

// GetTasks gets tasks based on query
func (db *GormDB) GetTasks(query *pb.Task) ([]*pb.Task, error) {
	var tasks []*pb.Task
	err := db.conn.Find(&tasks, query).Error
	if err != nil {
		return nil, err
	}
	if len(tasks) == 0 {
		return tasks, gorm.ErrRecordNotFound
	}
	return tasks, err
}

// CreateTasks creates a batch of tasks
func (db *GormDB) CreateTasks(tasks []*pb.Task) error {
	if len(tasks) == 0 {
		return nil
	}
	return db.conn.Create(tasks).Error
}

// getIssues gets issues based on query
func (db *GormDB) getIssues(query *pb.Issue) ([]*pb.Issue, error) {
	var issues []*pb.Issue
	err := db.conn.Find(&issues, query).Error
	if err != nil {
		return nil, err
	}
	return issues, err
}

// CreateIssues creates a batch of issues
func (db *GormDB) CreateIssues(issues []*pb.Issue) error {
	if len(issues) == 0 {
		return nil
	}
	return db.conn.Create(issues).Error
}

// SynchronizeAssignmentTasks synchronizes all tasks of each assignment in a given course. Returns created, updated and deleted tasks
func (db *GormDB) SynchronizeAssignmentTasks(course *pb.Course, taskMap map[uint32]map[string]*pb.Task) (createdTasks, updatedTasks []*pb.Task, err error) {
	createdTasks = []*pb.Task{}
	updatedTasks = []*pb.Task{}
	assignments, err := db.GetAssignmentsByCourse(course.GetID(), false)
	if err != nil {
		return nil, nil, err
	}

	err = db.conn.Transaction(func(tx *gorm.DB) error {
		for _, assignment := range assignments {
			existingTasks, err := db.GetTasks(&pb.Task{AssignmentID: assignment.GetID()})
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				// will rollback transaction
				return fmt.Errorf("failed to get tasks for assignment %d: %w", assignment.GetID(), err)
			}

			for _, existingTask := range existingTasks {
				task, ok := taskMap[assignment.Order][existingTask.Name]
				if !ok {
					// Find issues associated with the existing task and delete them
					var issues []*pb.Issue
					if err = tx.Delete(issues, &pb.Issue{TaskID: existingTask.ID}).Error; err != nil {
						return err // will rollback transaction
					}
					// Existing task in database not among the supplied tasks to synchronize.
					err = tx.Delete(existingTask).Error
					if err != nil {
						return err // will rollback transaction
					}
					existingTask.MarkDeleted()
					updatedTasks = append(updatedTasks, existingTask)
					continue
				}
				if existingTask.HasChanged(task) {
					// Task has been changed and must be updated.
					existingTask.Title = task.Title
					existingTask.Body = task.Body
					updatedTasks = append(updatedTasks, existingTask)
					err = tx.Model(&pb.Task{}).
						Where(&pb.Task{ID: existingTask.GetID()}).
						Updates(existingTask).Error
					if err != nil {
						return err // will rollback transaction
					}
				}
				delete(taskMap[assignment.Order], existingTask.Name)
			}

			// Find new tasks to be created for the current assignment
			for _, task := range taskMap[assignment.Order] {
				task.AssignmentID = assignment.ID
				createdTasks = append(createdTasks, task)
			}
		}

		// Tasks to be created must be sorted since map iteration order is non-deterministic
		sort.Slice(createdTasks, func(i, j int) bool {
			return createdTasks[i].Name < createdTasks[j].Name
		})

		// Create tasks that are not in the database
		for _, task := range createdTasks {
			if err = tx.Create(task).Error; err != nil {
				return err // will rollback transaction
			}
		}
		return nil
	})

	return createdTasks, updatedTasks, err
}
