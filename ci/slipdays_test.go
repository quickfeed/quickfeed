package ci

import (
	"testing"
	"time"

	pb "github.com/autograde/aguis/ag"
)

const (
	days = 24 * time.Hour
)

func TestSlipDays(t *testing.T) {
	course := &pb.Course{
		ID:       1,
		Name:     "opsys",
		SlipDays: 5,
	}

	repo := &pb.Repository{UserID: 2}
	enrol := &pb.Enrollment{
		CourseID:          course.ID,
		UserID:            repo.UserID,
		RemainingSlipDays: 5,
	}

	now := time.Now()
	lab1 := &pb.Assignment{
		Name:     "lab1",
		Deadline: now.Add(-2 * days).String(),
	}

	rData := &RunData{
		Course:     course,
		Assignment: lab1,
		Repo:       repo,
	}
	slipdays(rData, enrol)
}
