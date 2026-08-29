package database

import (
	"sort"

	"github.com/quickfeed/quickfeed/qf"
	"gorm.io/gorm"
)

// GetTasks gets tasks based on query
func (db *GormDB) GetTasks(query *qf.Task) ([]*qf.Task, error) {
	var tasks []*qf.Task
	err := db.conn.Find(&tasks, query).Error
	if err != nil {
		return nil, err
	}
	if len(tasks) == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return tasks, err
}

// CreateIssues creates a batch of issues
func (db *GormDB) CreateIssues(issues []*qf.Issue) error {
	if len(issues) == 0 {
		return nil
	}
	return db.conn.Create(issues).Error
}

// SynchronizeAssignmentTasks synchronizes all tasks of each assignment in a given course. Returns created, updated and deleted tasks
func (db *GormDB) SynchronizeAssignmentTasks(course *qf.Course, taskMap map[uint32]map[string]*qf.Task) (createdTasks, updatedTasks []*qf.Task, err error) {
	assignments, err := db.GetAssignmentsByCourse(course.GetID())
	if err != nil {
		return nil, nil, err
	}

	err = db.conn.Transaction(func(tx *gorm.DB) error {
		for _, assignment := range assignments {
			var existingTasks []*qf.Task
			if err := tx.Find(&existingTasks, &qf.Task{AssignmentID: assignment.GetID()}).Error; err != nil {
				return err // will rollback transaction
			}
			for _, existingTask := range existingTasks {
				task, ok := taskMap[assignment.GetOrder()][existingTask.GetName()]
				if !ok {
					// Find issues associated with the existing task and delete them
					var issues []*qf.Issue
					if err = tx.Delete(issues, &qf.Issue{TaskID: existingTask.GetID()}).Error; err != nil {
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
					existingTask.Title = task.GetTitle()
					existingTask.Body = task.GetBody()
					updatedTasks = append(updatedTasks, existingTask)
					err = tx.Model(&qf.Task{}).Select("*").
						Where(&qf.Task{ID: existingTask.GetID()}).
						Updates(existingTask).Error
					if err != nil {
						return err // will rollback transaction
					}
				}
				delete(taskMap[assignment.GetOrder()], existingTask.GetName())
			}

			// Find new tasks to be created for the current assignment
			for _, task := range taskMap[assignment.GetOrder()] {
				task.AssignmentID = assignment.GetID()
				createdTasks = append(createdTasks, task)
			}
		}

		// Tasks to be created must be sorted since map iteration order is non-deterministic
		sort.Slice(createdTasks, func(i, j int) bool {
			return createdTasks[i].GetID() < createdTasks[j].GetID()
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
