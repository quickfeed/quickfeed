package qf_test

import (
	"sync"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/quickfeed/quickfeed/qf"
)

func TestGetDockerfileEmptyCache(t *testing.T) {
	course := &qf.Course{}
	got := course.GetDockerfile()
	if got != "" {
		t.Errorf("GetDockerfile() = %s, want empty string", got)
	}
}

func TestUpdateDockerfile(t *testing.T) {
	course := &qf.Course{ID: 1}
	want := false
	dockerfile := ""
	got := course.UpdateDockerfile(dockerfile)
	if got != want {
		t.Errorf("UpdateDockerfile(%q) = %t, want %t", dockerfile, got, want)
	}

	want = true
	dockerfile = "FROM ubuntu:latest"
	got = course.UpdateDockerfile(dockerfile)
	if got != want {
		t.Errorf("UpdateDockerfile(%q) = %t, want %t", dockerfile, got, want)
	}

	want = false
	got = course.UpdateDockerfile(dockerfile)
	if got != want {
		t.Errorf("UpdateDockerfile(%q) = %t, want %t", dockerfile, got, want)
	}

	want = true
	dockerfile = "FROM golang:latest"
	got = course.UpdateDockerfile(dockerfile)
	if got != want {
		t.Errorf("UpdateDockerfile(%q) = %t, want %t", dockerfile, got, want)
	}
}

func TestLock(t *testing.T) {
	var wg sync.WaitGroup

	course := &qf.Course{ID: 1, CourseCreatorID: 0}
	want := uint64(10)
	rang := 5

	for i := range rang {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			unlock := course.Lock()
			defer unlock()

			course.CourseCreatorID += uint64(i)
		}(i)
	}

	wg.Wait()

	// Asserts the course is accessed concurrently and the course creator ID is updated correctly.
	if !cmp.Equal(course.GetCourseCreatorID(), want) {
		t.Errorf("CourseCreatorID = %v, want %v", course.GetCourseCreatorID(), want)
	}

	// Ensure the lock works as expected
	// Run two goroutines, expects only one update, because the lock is not released.

	want = uint64(0)
	for i := range 2 {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_ = course.Lock()
			course.CourseCreatorID = uint64(i)
		}(i)
	}

	ch := make(chan struct{})
	go func() {
		wg.Wait()
		ch <- struct{}{}
	}()

	select {
	case <-ch:
		t.Errorf("Lock failed to prevent concurrent access")
	case <-time.After(1 * time.Second):
		// Expected behavior
	}

	// Assert one of the goroutines updated the course creator ID
	if !cmp.Equal(course.GetCourseCreatorID(), want) {
		t.Errorf("CourseCreatorID = %v, want %v", course.GetCourseCreatorID(), want)
	}
}

func TestDockerfileForCourse(t *testing.T) {
	course := &qf.Course{ID: 1}
	want := "FROM ubuntu:latest"
	updated := course.UpdateDockerfile(want)
	if !updated {
		t.Errorf("UpdateDockerfile(%q) = %t, want %t", want, updated, true)
	}
	got := course.GetDockerfile()
	if got != want {
		t.Errorf("GetDockerfile() = %q, want %q", got, want)
	}

	want2 := "FROM golang:latest"
	got2 := course.GetDockerfile()
	if got2 == want2 {
		// They should be different, since the cache is not updated yet.
		t.Errorf("GetDockerfile() = %q, want %q", got2, want)
	}

	updated = course.UpdateDockerfile(want2)
	if !updated {
		t.Errorf("UpdateDockerfile(%q) = %t, want %t", want2, updated, true)
	}
	got2 = course.GetDockerfile()
	if got2 != want2 {
		// Now they should be the same since the cache is updated.
		t.Errorf("GetDockerfile() = %q, want %q", got2, want2)
	}
}
