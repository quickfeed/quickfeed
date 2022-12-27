package qf_test

import (
	"testing"

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
