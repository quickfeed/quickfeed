package database

import (
	pb "github.com/autograde/quickfeed/ag"
)

// CreateTasks creates slice of tasks
func (db *GormDB) CreateTasks(tasks []*pb.Task) (err error) {
	if len(tasks) == 0 {
		return nil
	}
	return db.conn.Create(tasks).Error
}

// UpdateTasks updates slice of tasks
func (db *GormDB) UpdateTasks(tasks []*pb.Task) (err error) {
	for _, task := range tasks {
		err = db.conn.Model(&pb.Task{}).
			Where(&pb.Task{ID: task.GetID()}).
			Updates(task).Error
		if err != nil {
			return err
		}
	}
	return nil
}

// GetTasks gets tasks based on query
func (db *GormDB) GetTasks(query *pb.Task) ([]*pb.Task, error) {
	var tasks []*pb.Task
	err := db.conn.Find(&tasks, query).Error
	if err != nil {
		return nil, err
	}
	return tasks, err
}

// CreateIssues creates a batch of issues
func (db *GormDB) CreateIssues(issues []*pb.Issue) error {
	if len(issues) == 0 {
		return nil
	}
	return db.conn.Create(issues).Error
}

// UpdateIssues updates a batch of issues
func (db *GormDB) UpdateIssues(issues []*pb.Issue) (err error) {
	for _, issue := range issues {
		err = db.conn.Model(&pb.Issue{}).
			Where(&pb.Issue{ID: issue.GetID()}).
			Updates(issue).Error
		if err != nil {
			return err
		}
	}
	return nil
}
