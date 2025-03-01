package qf

import (
	"crypto/sha256"
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/quickfeed/quickfeed/internal/env"
	"github.com/quickfeed/quickfeed/internal/rand"
)

// Cached Dockerfile for each course.
var courseDockerfileCache = make(map[uint64]string)

// UpdateDockerfile updates the course's dockerfile cache and returns true
// if the given dockerfile was different from the course's previous Dockerfile.
// This method will also update the course's DigestDockerfile field so that
// changes to the dockerfile is reflected in the database.
func (course *Course) UpdateDockerfile(dockerfile string) bool {
	if dockerfile == "" {
		return false
	}
	// Always cache the dockerfile even if it has not been updated.
	// This ensures that the calls to GetDockerfile() can return it
	// even after a restart of the server.
	courseDockerfileCache[course.ID] = dockerfile
	dockerDigest := digest(dockerfile)
	updated := course.DockerfileDigest != dockerDigest
	if updated {
		course.DockerfileDigest = dockerDigest
	}
	return updated
}

// Mutex for each course
var (
	courseMuMap = make(map[uint64]*sync.Mutex)
	mapMu       = sync.Mutex{}
)

// Lock indexes the course mutex map with the course ID and locks the mutex.
// The mutex is initialized if it does not exist.
// This method is called when concurrently accessing the course.
func (course *Course) Lock() {
	mapMu.Lock()
	if _, ok := courseMuMap[course.ID]; !ok {
		courseMuMap[course.ID] = &sync.Mutex{}
	}
	mu := courseMuMap[course.ID]
	mapMu.Unlock()

	mu.Lock()
}

// Unlock indexes the course mutex map with the course ID and unlocks the mutex.
// This method is called when concurrently accessing the course.
func (course *Course) Unlock() {
	mapMu.Lock()
	mu, ok := courseMuMap[course.ID]
	mapMu.Unlock()
	if ok { // Will always be true if Lock() has been called.
		mu.Unlock()
	}
}

func (course *Course) GetDockerfile() string {
	return courseDockerfileCache[course.ID]
}

func (course *Course) DockerImage() string {
	return strings.ToLower(course.GetCode())
}

func (course *Course) JobName() string {
	return course.GetCode() + "-" + rand.String()
}

// digest returns a SHA256 digest of the given file.
func digest(file string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(file)))
}

func (course *Course) CloneDir() string {
	return filepath.Join(env.RepositoryPath(), course.GetScmOrganizationName())
}

func (course *Course) TeacherEnrollments() []*Enrollment {
	enrolledTeachers := []*Enrollment{}
	for _, enrollment := range course.Enrollments {
		if enrollment.IsTeacher() {
			enrolledTeachers = append(enrolledTeachers, enrollment)
		}
	}
	return enrolledTeachers
}

// PopulateSlipDays populates the slip days for all enrollments in the course.
func (course *Course) PopulateSlipDays() {
	// Set number of remaining slip days for each course enrollment
	for _, enrollment := range course.GetEnrollments() {
		enrollment.SetSlipDays(course)
	}
	for _, group := range course.GetGroups() {
		// Set number of remaining slip days for each group enrollment
		for _, enrollment := range group.GetEnrollments() {
			enrollment.SetSlipDays(course)
		}
	}
}

// Dummy implementation of the interceptor.userIDs interface.
// Marks this message type to be evaluated for token refresh.
func (*Course) UserIDs() []uint64 {
	return []uint64{}
}
