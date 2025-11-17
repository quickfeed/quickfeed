package qtest

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/google/go-cmp/cmp"
	"github.com/quickfeed/quickfeed/database"
	"github.com/quickfeed/quickfeed/internal/fileop"
	"github.com/quickfeed/quickfeed/qf"
)

// TestDB returns a test database and close function.
// This function should only be used as a test helper.
func TestDB(t *testing.T) (database.Database, func()) {
	t.Helper()

	f, err := os.CreateTemp(t.TempDir(), "test.db")
	if err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		os.Remove(f.Name())
		t.Fatal(err)
	}

	db, err := database.NewGormDB(f.Name(), Logger(t).Desugar())
	if err != nil {
		os.Remove(f.Name())
		t.Fatal(err)
	}

	return db, func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
		if err := os.Remove(f.Name()); err != nil {
			t.Error(err)
		}
	}
}

func SetupCourseAssignment(t *testing.T, db database.Database) (*qf.User, *qf.Course, *qf.Assignment) {
	// create a course and an assignment
	admin := CreateFakeUser(t, db)
	course := &qf.Course{}
	CreateCourse(t, db, admin, course)
	assignment := &qf.Assignment{
		CourseID: course.GetID(),
		Order:    1,
	}
	CreateAssignment(t, db, assignment)
	// create user and enroll as student
	user := CreateFakeUser(t, db)
	EnrollStudent(t, db, user, course)
	return user, course, assignment
}

// SetupCourseAssignmentTeacherStudent returns the admin (teacher) user, course, an assignment, and a student user.
func SetupCourseAssignmentTeacherStudent(t *testing.T, db database.Database) (*qf.User, *qf.Course, *qf.Assignment, *qf.User) {
	admin := CreateFakeUser(t, db)
	course := &qf.Course{}
	CreateCourse(t, db, admin, course)
	assignment := &qf.Assignment{
		CourseID: course.GetID(),
		Order:    1,
	}
	CreateAssignment(t, db, assignment)
	user := CreateFakeUser(t, db)
	EnrollStudent(t, db, user, course)
	return admin, course, assignment, user
}

// PrepareGitRepo creates copies src/repo folder to dst and initializes
// dst/repo as a git repository and adds a single file lab1/lab1.go.
func PrepareGitRepo(t *testing.T, src, dst, repo string) {
	if err := fileop.CopyDir(filepath.Join(src, repo), dst); err != nil {
		t.Fatal(err)
	}
	gitRepo := filepath.Join(dst, repo)
	r, err := git.PlainInit(gitRepo, false)
	if err != nil {
		t.Fatal(err)
	}
	w, err := r.Worktree()
	if err != nil {
		t.Fatal(err)
	}
	if _, err = w.Add("lab1"); err != nil {
		t.Fatal(err)
	}
	if _, err = w.Commit("added lab1", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test",
			Email: "test@itest.run",
			When:  time.Now(),
		},
	}); err != nil {
		t.Fatal(err)
	}
}

// CreateFakeUser is a test helper to create a user in the database.
func CreateFakeUser(t *testing.T, db database.Database) *qf.User {
	t.Helper()
	user := &qf.User{}
	if err := db.CreateUser(user); err != nil {
		t.Fatal(err)
	}
	return user
}

func CreateFakeCustomUser(t *testing.T, db database.Database, user *qf.User) *qf.User {
	t.Helper()
	if err := db.CreateUser(user); err != nil {
		t.Fatal(err)
	}
	return user
}

func UpdateUser(t *testing.T, db database.Database, user *qf.User) {
	t.Helper()
	if err := db.UpdateUser(user); err != nil {
		t.Fatal(err)
	}
}

// CreateCourse is a test helper to create a course in the database; it updates the course with the ID.
func CreateCourse(t *testing.T, db database.Database, user *qf.User, course *qf.Course) {
	t.Helper()
	if err := db.CreateCourse(user.GetID(), course); err != nil {
		t.Fatal(err)
	}
}

func GetCourse(t *testing.T, db database.Database, courseID uint64) *qf.Course {
	t.Helper()
	course, err := db.GetCourse(courseID)
	if err != nil {
		t.Fatal(err)
	}
	return course
}

func CreateRepository(t *testing.T, db database.Database, repo *qf.Repository) {
	t.Helper()
	if err := db.CreateRepository(repo); err != nil {
		t.Fatal(err)
	}
}

func CreateAssignment(t *testing.T, db database.Database, assignment *qf.Assignment) {
	t.Helper()
	if err := db.CreateAssignment(assignment); err != nil {
		t.Fatal(err)
	}
}

func UpdateAssignments(t *testing.T, db database.Database, assignment []*qf.Assignment) {
	t.Helper()
	if err := db.UpdateAssignments(assignment); err != nil {
		t.Fatal(err)
	}
}

func GetAssignment(t *testing.T, db database.Database, assignmentID uint64) *qf.Assignment {
	t.Helper()
	assignment, err := db.GetAssignment(&qf.Assignment{ID: assignmentID})
	if err != nil {
		t.Fatal(err)
	}
	return assignment
}

func GetAssignments(t *testing.T, db database.Database, courseID uint64) []*qf.Assignment {
	t.Helper()
	assignments, err := db.GetAssignmentsByCourse(courseID)
	if err != nil {
		t.Fatal(err)
	}
	return assignments
}

func CreateSubmission(t *testing.T, db database.Database, submission *qf.Submission) {
	t.Helper()
	if err := db.CreateSubmission(submission); err != nil {
		t.Fatal(err)
	}
}

func GetSubmission(t *testing.T, db database.Database, submission *qf.Submission) *qf.Submission {
	t.Helper()
	submission, err := db.GetSubmission(submission)
	if err != nil {
		t.Fatal(err)
	}
	return submission
}

func GetSubmissions(t *testing.T, db database.Database, submission *qf.Submission) []*qf.Submission {
	t.Helper()
	submissions, err := db.GetSubmissions(submission)
	if err != nil {
		t.Fatal(err)
	}
	return submissions
}

func CreateReview(t *testing.T, db database.Database, review *qf.Review) {
	t.Helper()
	if err := db.CreateReview(review); err != nil {
		t.Fatal(err)
	}
}

func GetReview(t *testing.T, db database.Database, id uint64) *qf.Review {
	t.Helper()
	review, err := db.GetReview(&qf.Review{ID: id})
	if err != nil {
		t.Fatal(err)
	}
	return review
}

func CreateBenchmark(t *testing.T, db database.Database, benchmark *qf.GradingBenchmark) {
	t.Helper()
	if err := db.CreateBenchmark(benchmark); err != nil {
		t.Fatal(err)
	}
}

func GetBenchmarks(t *testing.T, db database.Database, assignmentID uint64) []*qf.GradingBenchmark {
	t.Helper()
	benchmarks, err := db.GetBenchmarks(&qf.Assignment{ID: assignmentID})
	if err != nil {
		t.Fatal(err)
	}
	return benchmarks
}

func GetBenchmark(t *testing.T, db database.Database, assignmentID, benchmarkID uint64) *qf.GradingBenchmark {
	t.Helper()
	benchmarks, err := db.GetBenchmarks(&qf.Assignment{ID: assignmentID})
	if err != nil {
		t.Fatal(err)
	}
	for _, bm := range benchmarks {
		if bm.GetID() == benchmarkID {
			return bm
		}
	}
	t.Fatalf("benchmark %d not found for assignment %d", benchmarkID, assignmentID)
	return nil
}

func CreateCriterion(t *testing.T, db database.Database, criterion *qf.GradingCriterion) {
	t.Helper()
	if err := db.CreateCriterion(criterion); err != nil {
		t.Fatal(err)
	}
}

func CreateFakeGroup(t *testing.T, db database.Database, course *qf.Course, groupSize int) *qf.Group {
	t.Helper()
	var users []*qf.User
	for range groupSize {
		user := CreateFakeUser(t, db)
		users = append(users, user)
		EnrollStudent(t, db, user, course)
	}
	group := &qf.Group{
		CourseID: course.ID,
		Name:     "group " + RandomString(t),
		Users:    users,
	}
	if err := db.CreateGroup(group); err != nil {
		t.Fatal(err)
	}
	return group
}

func CreateGroup(t *testing.T, db database.Database, group *qf.Group) *qf.Group {
	t.Helper()
	if err := db.CreateGroup(group); err != nil {
		t.Fatal(err)
	}
	return group
}

func GetEnrollment(t *testing.T, db database.Database, userID, courseID uint64) *qf.Enrollment {
	t.Helper()
	enrollment, err := db.GetEnrollmentByCourseAndUser(courseID, userID)
	if err != nil {
		t.Fatal(err)
	}
	return enrollment
}

func EnrollStudent(t *testing.T, db database.Database, student *qf.User, course *qf.Course) {
	t.Helper()
	query := &qf.Enrollment{
		UserID:   student.GetID(),
		CourseID: course.GetID(),
	}
	if err := db.CreateEnrollment(query); err != nil {
		t.Fatal(err)
	}
	query.Status = qf.Enrollment_STUDENT
	if err := db.UpdateEnrollment(query); err != nil {
		t.Fatal(err)
	}
}

func EnrollTeacher(t *testing.T, db database.Database, student *qf.User, course *qf.Course) {
	t.Helper()
	query := &qf.Enrollment{
		UserID:   student.GetID(),
		CourseID: course.GetID(),
	}
	if err := db.CreateEnrollment(query); err != nil {
		t.Fatal(err)
	}
	query.Status = qf.Enrollment_TEACHER
	if err := db.UpdateEnrollment(query); err != nil {
		t.Fatal(err)
	}
}

func EnrollUser(t *testing.T, db database.Database, user *qf.User, course *qf.Course, status qf.Enrollment_UserStatus) *qf.Enrollment {
	t.Helper()
	enrollment := &qf.Enrollment{
		UserID:   user.GetID(),
		CourseID: course.GetID(),
	}
	if err := db.CreateEnrollment(enrollment); err != nil {
		t.Fatal(err)
	}
	enrollment.Status = status
	if err := db.UpdateEnrollment(enrollment); err != nil {
		t.Fatal(err)
	}
	return enrollment
}

// CreateTempFile creates a temporary file in the given directory.
// The file is automatically removed when the test ends.
func CreateTempFile(t *testing.T, dir string) string {
	t.Helper()
	tmpFile, err := os.CreateTemp(dir, "*")
	if err != nil {
		t.Fatal(err)
	}
	envFileName := tmpFile.Name()
	t.Cleanup(func() {
		if err := os.Remove(envFileName); err != nil {
			t.Error(err)
		}
	})
	return envFileName
}

func RandomString(t *testing.T) string {
	t.Helper()
	randomness := make([]byte, 10)
	if _, err := rand.Read(randomness); err != nil {
		t.Fatal(err)
	}
	return fmt.Sprintf("%x", sha256.Sum256(randomness))[:6]
}

func RequestWithCookie[T any](message *T, cookie string) *connect.Request[T] {
	request := connect.NewRequest(message)
	request.Header().Set("cookie", cookie)
	return request
}

// Ptr returns a pointer to the given value.
//
// How to use:
//   - Use this function to create a pointer to a value.
//   - This function is useful when initializing a struct with a pointer field.
//
// Example:
//
//	type MyStruct struct {
//		Field *int
//	    Src   *string
//	}
//	myStruct := MyStruct{
//		Field: Ptr(10),
//		Src:   Ptr("hello"),
//	}
func Ptr[T any](t T) *T {
	return &t
}

// Diff compares the got and want values and prints a diff with the given message.
func Diff(t *testing.T, msg string, got, want any, opts ...cmp.Option) {
	if diff := cmp.Diff(got, want, opts...); diff != "" {
		t.Errorf("%s: (-got +want)\n%s", msg, diff)
	}
}

// CheckError checks if the got error matches the want error and fails the test if not.
func CheckError(t *testing.T, got, want error) {
	if got != nil {
		if want == nil {
			t.Fatalf("Expected no error, got: %v", got)
		}
		if got.Error() != want.Error() {
			t.Fatalf("Expected error: %v, got: %v", want, got)
		}
	} else if want != nil {
		t.Fatalf("Expected error: %v, got: nil", want)
	}
}

// CheckCode checks if the got error matches the want error, and its code, and fails the test if not.
// It returns true if got is an error, which indicates that the test should stop.
func CheckCode(t *testing.T, got, want error) bool {
	if got != nil {
		if want == nil {
			t.Fatalf("Expected no error, got: %v", got)
		}
		if got.Error() != want.Error() {
			t.Errorf("Expected error: %v, got: %v", want, got)
		}
	} else if want != nil {
		t.Errorf("Expected error: %v, got: nil", want)
	}
	if connect.CodeOf(got) != connect.CodeOf(want) {
		t.Errorf("Expected error code: %v, got: %v", connect.CodeOf(want), connect.CodeOf(got))
	}
	return got != nil
}
