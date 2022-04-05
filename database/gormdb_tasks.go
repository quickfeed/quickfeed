package database

import (
	"errors"
	"fmt"

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

// DeleteIssuesOfAssociatedTasks deletes a batch of issues
func (db *GormDB) DeleteIssuesOfAssociatedTasks(tasks []*pb.Task) error {
	err := db.conn.Transaction(func(tx *gorm.DB) error {
		for _, task := range tasks {
			issues, err := db.getIssues(&pb.Issue{TaskID: task.ID})
			if err != nil {
				return err
			}

			if err = tx.Delete(issues).Error; err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

// SynchronizeAssignmentTasks synchronizes all tasks of each assignment in a given course. Returns created, updated and deleted tasks
func (db *GormDB) SynchronizeAssignmentTasks(course *pb.Course, taskMap map[uint32]map[string]*pb.Task) (createdTasks, updatedTasks, deletedTasks []*pb.Task, err error) {
	createdTasks = []*pb.Task{}
	updatedTasks = []*pb.Task{}
	deletedTasks = []*pb.Task{}
	assignments, err := db.GetAssignmentsByCourse(course.GetID(), false)
	if err != nil {
		return createdTasks, updatedTasks, deletedTasks, err
	}

	err = db.conn.Transaction(func(tx *gorm.DB) error {
		for _, assignment := range assignments {
			existingTasks, err := db.GetTasks(&pb.Task{AssignmentID: assignment.GetID()})
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("failed to get tasks for assignment %d: %w", assignment.GetID(), err)
			}

			for _, existingTask := range existingTasks {
				task, ok := taskMap[assignment.Order][existingTask.Name]
				if !ok {
					// There exists a task in db, that is not represented by any supplied task.
					deletedTasks = append(deletedTasks, existingTask)
					_ = tx.Delete(existingTask)
					continue
				}
				if existingTask.HasChanged(task) {
					// Task has been changed, and is therefore updated.
					existingTask.Title = task.Title
					existingTask.Body = task.Body
					updatedTasks = append(updatedTasks, existingTask)
					err = tx.Model(&pb.Task{}).
						Where(&pb.Task{ID: existingTask.GetID()}).
						Updates(existingTask).Error
					if err != nil {
						return err
					}
				}
				delete(taskMap[assignment.Order], existingTask.Name)
			}

			// Only tasks that there is no existing record of remain. These are created.
			for _, task := range taskMap[assignment.Order] {
				task.AssignmentID = assignment.ID
				createdTasks = append(createdTasks, task)
				if err = tx.Create(task).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})

	return createdTasks, updatedTasks, deletedTasks, err
}

// CreatePullRequest creates a pull request
func (db *GormDB) CreatePullRequest(pullRequest *pb.PullRequest) error {
	return db.conn.Create(pullRequest).Error
}

// GetPullRequest returns the pull request matching the given query
func (db *GormDB) GetPullRequest(query *pb.PullRequest) (*pb.PullRequest, error) {
	var pullRequest *pb.PullRequest
	err := db.conn.Last(pullRequest, query).Error
	return pullRequest, err
}

// DeletePullRequest deletes the pull request matching the given query
func (db *GormDB) DeletePullRequest(pullRequest *pb.PullRequest) error {
	return db.conn.Delete(pullRequest).Error
}

// DeletePullRequest updates the pull request matching the given query
func (db *GormDB) UpdatePullRequest(pullRequest *pb.PullRequest) error {
	return db.conn.Model(&pb.PullRequest{}).
		Where(&pb.PullRequest{ID: pullRequest.GetID()}).
		Updates(pullRequest).Error
}
