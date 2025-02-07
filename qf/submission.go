package qf

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

func (s *Submission) IsApproved(userID uint64) bool {
	for _, grade := range s.GetGrades() {
		if grade.GetUserID() == userID && grade.GetStatus() == Submission_APPROVED {
			return true
		}
	}
	return false
}

func (s *Submission) IsAllApproved() bool {
	for _, grade := range s.GetGrades() {
		if grade.GetStatus() != Submission_APPROVED {
			return false
		}
	}
	return true
}

func (s *Submission) GetStatuses() []Submission_Status {
	statuses := make([]Submission_Status, len(s.GetGrades()))
	for idx, grade := range s.GetGrades() {
		statuses[idx] = grade.GetStatus()
	}
	return statuses
}

func (s *Submission) GetStatusByUser(userID uint64) Submission_Status {
	for idx, grade := range s.GetGrades() {
		if grade.GetUserID() == userID {
			return s.Grades[idx].GetStatus()
		}
	}
	return Submission_NONE
}

func (s *Submission) SetGrade(userID uint64, status Submission_Status) {
	for idx, grade := range s.GetGrades() {
		if grade.GetUserID() == userID {
			s.Grades[idx].Status = status
			return
		}
	}
}

func (s *Submission) SetGradeAll(status Submission_Status) {
	for idx := range s.GetGrades() {
		s.Grades[idx].Status = status
	}
}

// NewestSubmissionDate returns the submission's submission date if newer than the provided date.
// Otherwise, the provided date is returned, i.e., if it is newer.
func (s *Submission) NewestSubmissionDate(submissionDate time.Time) time.Time {
	currentSubmissionDate := s.GetBuildInfo().GetSubmissionDate().AsTime()
	if currentSubmissionDate.After(submissionDate) {
		return currentSubmissionDate
	}
	return submissionDate
}

func (s *Submission) ByUser(userID uint64) bool {
	return s.GetGroupID() == 0 && s.GetUserID() > 0 && s.GetUserID() == userID
}

func (s *Submission) ByGroup(groupID uint64) bool {
	return s.GetUserID() == 0 && s.GetGroupID() > 0 && s.GetGroupID() == groupID
}

// Clean removes any score or reviews from the submission if it is not released.
// This is to prevent users from seeing the score or reviews of a submission that has not been released.
func (s *Submissions) Clean(userID uint64) {
	for _, submission := range s.Submissions {
		// Group submissions may have multiple grades, so we need to filter the grades by the user.
		submission.Grades = []*Grade{{
			UserID:       userID,
			SubmissionID: submission.GetID(),
			Status:       submission.GetStatusByUser(userID),
		}}
		// Released submissions, or submissions with no reviews need no cleaning.
		if submission.GetReleased() || len(submission.GetReviews()) == 0 {
			continue
		}
		// Remove any score, grades, or reviews if the submission is not released.
		submission.Score = 0
		submission.Grades = nil
		submission.Reviews = nil
	}
}

// BeforeCreate is called before a new submission is created.
// This method adds grades for any user or group related to the submission
// which are then saved to the database upon creation of the submission.
func (s *Submission) BeforeCreate(tx *gorm.DB) error {
	if s.GetUserID() == 0 && s.GetGroupID() == 0 {
		return errors.New("submission must have either user or group ID")
	}
	if s.GetUserID() > 0 {
		// Add a grade for the user if the submission is not a group submission.
		// Create a new grade for the user.
		s.Grades = []*Grade{{
			UserID:       s.GetUserID(),
			SubmissionID: s.GetID(),
			Status:       s.GetStatusByUser(s.GetUserID()),
		}}
	}
	if s.GetGroupID() > 0 {
		// If the submission is for a group, create a new grade for each user in the group.
		// Get all the user IDs in the group
		userIDs := []uint64{}
		tx.Model(&Enrollment{}).Where("group_id = ?", s.GetGroupID()).Pluck("user_id", &userIDs)

		if len(userIDs) == 0 {
			return errors.New("group has no users")
		}

		s.Grades = make([]*Grade, len(userIDs))
		for idx, id := range userIDs {
			// Create a grade for each user in the group
			s.Grades[idx] = &Grade{
				UserID:       id,
				SubmissionID: s.GetID(),
				Status:       s.GetStatusByUser(id),
			}
		}
	}
	return nil
}
