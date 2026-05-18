package database

import (
	"context"
	"database/sql"
	"errors"
	"sort"

	"github.com/quickfeed/quickfeed/qf"
	"github.com/uptrace/bun"
)

// GetTasks returns tasks based on the given query.
func (db *BunDB) GetTasks(query *qf.Task) ([]*qf.Task, error) {
	ctx := context.Background()
	var tasks []*qf.Task
	q := db.conn.NewSelect().Model(&tasks)
	if query.GetID() > 0 {
		q = q.Where("id = ?", query.GetID())
	}
	if query.GetAssignmentID() > 0 {
		q = q.Where("assignment_id = ?", query.GetAssignmentID())
	}
	if query.GetName() != "" {
		q = q.Where("name = ?", query.GetName())
	}
	if err := q.Scan(ctx); err != nil {
		return nil, err
	}
	if len(tasks) == 0 {
		return nil, toDBError(sql.ErrNoRows)
	}
	return tasks, nil
}

// CreateIssues creates a batch of issues.
func (db *BunDB) CreateIssues(issues []*qf.Issue) error {
	if len(issues) == 0 {
		return nil
	}
	ctx := context.Background()
	_, err := db.conn.NewInsert().Model(&issues).Exec(ctx)
	return err
}

// SynchronizeAssignmentTasks synchronizes all tasks of each assignment in a given course.
// Returns created and updated tasks.
func (db *BunDB) SynchronizeAssignmentTasks(course *qf.Course, taskMap map[uint32]map[string]*qf.Task) (createdTasks, updatedTasks []*qf.Task, err error) {
	assignments, err := db.GetAssignmentsByCourse(course.GetID())
	if err != nil {
		return nil, nil, err
	}

	ctx := context.Background()
	err = db.conn.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		for _, assignment := range assignments {
			var existingTasks []*qf.Task
			if err := tx.NewSelect().Model(&existingTasks).
				Where("assignment_id = ?", assignment.GetID()).
				Scan(ctx); err != nil {
				return err
			}
			for _, existingTask := range existingTasks {
				task, ok := taskMap[assignment.GetOrder()][existingTask.GetName()]
				if !ok {
					if _, err := tx.NewDelete().Model((*qf.Issue)(nil)).
						Where("task_id = ?", existingTask.GetID()).Exec(ctx); err != nil {
						return err
					}
					if _, err := tx.NewDelete().Model(existingTask).WherePK().Exec(ctx); err != nil {
						return err
					}
					existingTask.MarkDeleted()
					updatedTasks = append(updatedTasks, existingTask)
					continue
				}
				if existingTask.HasChanged(task) {
					existingTask.Title = task.GetTitle()
					existingTask.Body = task.GetBody()
					if _, err := tx.NewUpdate().Model(existingTask).WherePK().Exec(ctx); err != nil {
						return err
					}
					updatedTasks = append(updatedTasks, existingTask)
				}
				delete(taskMap[assignment.GetOrder()], existingTask.GetName())
			}

			for _, task := range taskMap[assignment.GetOrder()] {
				task.AssignmentID = assignment.GetID()
				createdTasks = append(createdTasks, task)
			}
		}

		sort.Slice(createdTasks, func(i, j int) bool {
			return createdTasks[i].GetID() < createdTasks[j].GetID()
		})

		for _, task := range createdTasks {
			if _, err := tx.NewInsert().Model(task).Exec(ctx); err != nil {
				return err
			}
		}
		return nil
	})

	return createdTasks, updatedTasks, err
}

// CreatePullRequest creates a pull request.
// It is initially in the "draft" stage, signaling that it is not yet ready for review.
func (db *BunDB) CreatePullRequest(pullRequest *qf.PullRequest) error {
	if !pullRequest.Valid() {
		return errors.New("pull request is not valid for creation")
	}
	pullRequest.SetDraft()
	ctx := context.Background()
	_, err := db.conn.NewInsert().Model(pullRequest).Exec(ctx)
	return err
}

// GetPullRequest returns the pull request matching the given query.
func (db *BunDB) GetPullRequest(query *qf.PullRequest) (*qf.PullRequest, error) {
	ctx := context.Background()
	var pullRequest qf.PullRequest
	q := db.conn.NewSelect().Model(&pullRequest)
	if query.GetID() > 0 {
		q = q.Where("id = ?", query.GetID())
	}
	if query.GetScmRepositoryID() > 0 {
		q = q.Where("scm_repository_id = ?", query.GetScmRepositoryID())
	}
	if query.GetUserID() > 0 {
		q = q.Where("user_id = ?", query.GetUserID())
	}
	if query.GetIssueID() > 0 {
		q = q.Where("issue_id = ?", query.GetIssueID())
	}
	if err := q.OrderExpr("id DESC").Limit(1).Scan(ctx); err != nil {
		return nil, err
	}
	return &pullRequest, nil
}

// HandleMergingPR handles a merged pull request.
// If not approved, deletes only the pull request.
// If approved, deletes the associated issue and then the pull request.
func (db *BunDB) HandleMergingPR(pullRequest *qf.PullRequest) error {
	ctx := context.Background()
	if !pullRequest.IsApproved() {
		_, err := db.conn.NewDelete().Model(pullRequest).WherePK().Exec(ctx)
		return err
	}
	var issue qf.Issue
	if err := db.conn.NewSelect().Model(&issue).
		Where("id = ?", pullRequest.GetIssueID()).Scan(ctx); err != nil {
		return err
	}
	if _, err := db.conn.NewDelete().Model(&issue).WherePK().Exec(ctx); err != nil {
		return err
	}
	_, err := db.conn.NewDelete().Model(pullRequest).WherePK().Exec(ctx)
	return err
}

// UpdatePullRequest updates the pull request matching the given query.
func (db *BunDB) UpdatePullRequest(pullRequest *qf.PullRequest) error {
	ctx := context.Background()
	_, err := db.conn.NewUpdate().Model(pullRequest).WherePK().Exec(ctx)
	return err
}
