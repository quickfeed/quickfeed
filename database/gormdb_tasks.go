package database

import (
	"errors"
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
		return tasks, gorm.ErrRecordNotFound
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
	createdTasks = []*qf.Task{}
	updatedTasks = []*qf.Task{}
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
					existingTask.Title = task.Title
					existingTask.Body = task.Body
					updatedTasks = append(updatedTasks, existingTask)
					err = tx.Model(&qf.Task{}).Select("*").
						Where(&qf.Task{ID: existingTask.GetID()}).
						Updates(existingTask).Error
					if err != nil {
						return err // will rollback transaction
					}
				}
				delete(taskMap[assignment.Order], existingTask.Name)
			}

			// Find new tasks to be created for the current assignment
			for _, task := range taskMap[assignment.GetOrder()] {
				task.AssignmentID = assignment.GetID()
				createdTasks = append(createdTasks, task)
			}
		}

		// Tasks to be created must be sorted since map iteration order is non-deterministic
		sort.Slice(createdTasks, func(i, j int) bool {
			return createdTasks[i].ID < createdTasks[j].GetID()
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

// CreatePullRequest creates a pull request.
// It is initially in the "draft" stage, signaling that it is not yet ready for review
func (db *GormDB) CreatePullRequest(pullRequest *qf.PullRequest) error {
	if !pullRequest.Valid() {
		return errors.New("pull request is not valid for creation")
	}
	pullRequest.SetDraft()
	return db.conn.Create(pullRequest).Error
}

// GetPullRequest returns the pull request matching the given query
func (db *GormDB) GetPullRequest(query *qf.PullRequest) (*qf.PullRequest, error) {
	var pullRequest qf.PullRequest
	if err := db.conn.Where(query).Last(&pullRequest).Error; err != nil {
		return nil, err
	}
	return &pullRequest, nil
}

// HandleMergingPR handles merging a pull request
// If a pull request has not been approved, it should not have been merged.
// We therefore do not delete the associated issue.
// To resume a working state, students are expected to reopen
// the issue that was closed from this merging, and create a new PR for it.
func (db *GormDB) HandleMergingPR(pullRequest *qf.PullRequest) error {
	if !pullRequest.IsApproved() {
		return db.conn.Delete(pullRequest).Error
	}
	var associatedIssue *qf.Issue
	if err := db.conn.First(associatedIssue, &qf.Issue{ID: pullRequest.GetIssueID()}).Error; err != nil {
		return err
	}
	_ = db.conn.Delete(associatedIssue).Error
	return db.conn.Delete(pullRequest).Error
}

// DeletePullRequest updates the pull request matching the given query
func (db *GormDB) UpdatePullRequest(pullRequest *qf.PullRequest) error {
	return db.conn.Model(&qf.PullRequest{}).Select("*").
		Where(&qf.PullRequest{ID: pullRequest.GetID()}).
		Updates(pullRequest).Error
}
