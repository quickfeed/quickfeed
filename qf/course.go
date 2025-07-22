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
	courseDockerfileCache[course.GetID()] = dockerfile
	dockerDigest := digest(dockerfile)
	updated := course.GetDockerfileDigest() != dockerDigest
	if updated {
		course.DockerfileDigest = dockerDigest
	}
	return updated
}

var (
	// courseMutexMap contains a single mutex for each course.
	courseMutexMap = make(map[uint64]*sync.Mutex)
	// mapMutex protects the courseMutexMap.
	mapMutex = sync.Mutex{}
)

// Lock locks the course to prevent concurrent updates to the course.
// It returns a corresponding unlock function which must be called when the update is done.
// Specifically, this method is used to prevent concurrent database updates
// derived from the test repository. See [assignments.UpdateFromTestsRepo].
func (course *Course) Lock() func() {
	mapMutex.Lock()
	if _, ok := courseMutexMap[course.GetID()]; !ok {
		courseMutexMap[course.GetID()] = &sync.Mutex{}
	}
	mu := courseMutexMap[course.GetID()]
	mapMutex.Unlock()

	mu.Lock()
	return mu.Unlock
}

func (course *Course) GetDockerfile() string {
	return courseDockerfileCache[course.GetID()]
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
	var enrolledTeachers []*Enrollment
	for _, enrollment := range course.GetEnrollments() {
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
