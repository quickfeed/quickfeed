package kube_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/autograde/aguis/ci/kube"
)

var course = "agcicd"

//var scriptPath = "kube/kube_scripts"

func newKubeCI() *kube.K8s {
	return &kube.K8s{}
}

type test struct {
	script, wantOut, out string
}

func TestK8s1(t *testing.T) {
	testK8s(t, "1")
}

func TestK8s2(t *testing.T) {
	testK8s(t, "2")
}

func TestK8s3(t *testing.T) {
	testK8s(t, "3")
}

func TestK8s4(t *testing.T) {
	testK8s(t, "4")
}

func testK8s(t *testing.T, echo string) {
	script := `echo -n ` + echo
	wantOut := echo
	jobName := time.Now().Format("20060102-150405-") + echo
	sec := "59fd5fe1c4f741604c1beeab875b9c789d2a7c73"

	err := kube.Jobsecrets(jobName, "agcicd", sec)
	if err != nil {
		panic(err)
	}
	container := &kube.Container{
		Image:    "golang",
		Commands: []string{script},
	}
	k := newKubeCI()
	out, err := k.KRun(context.Background(), container, jobName, "agcicd")
	if err != nil {
		t.Fatal(err)
	}

	if out != wantOut {
		t.Errorf("have %#v want %#v", out, wantOut)
	}
}

func TestK8sFP(t *testing.T) {
	tea := time.Now()
	fmt.Println(tea.Format("20060102-150405"))
	cloneURL := "https://github.com/dat320-2019/assignments.git"
	getURLTest := "https://github.com/dat320-2019/tests.git"

	sec := "59fd5fe1c4f741604c1beeab875b9c789d2a7c73"
	jobName := "jobName"
	err := kube.Jobsecrets(jobName, "agcicd", sec)
	if err != nil {
		panic(err)
	}
	//jobName := tea.Format("20060102-150405")
	ass := &kube.AssignmentInfo{
		AssignmentName:     "lab5",
		Language:           "go",
		CreatorAccessToken: "c0e4b71f27145d0653225d6415f65e39e1ab0f7f",
		GetURL:             cloneURL,
		TestURL:            getURLTest,
		RawGetURL:          strings.TrimPrefix(strings.TrimSuffix(cloneURL, ".git"), "https://"),
		RawTestURL:         strings.TrimPrefix(strings.TrimSuffix(getURLTest, ".git"), "https://"),
		RandomSecret:       sec,
	}
	jobdock, err := kube.ParseKubeScriptTemplate("", ass)
	if err != nil {
		panic(err)
	}
	wantOut := ""
	script := jobdock.Commands

	container := &kube.Container{
		Image:    "golang",
		Commands: script,
	}

	k := newKubeCI()
	out, err := k.KRun(context.Background(), container, jobName, "agcicd")
	if err != nil {
		t.Fatal(err)
	}

	if out != wantOut {
		t.Errorf("have %#v want %#v", out, wantOut)
	} else {
		fmt.Println(wantOut)
	}
	fmt.Println(time.Since(tea))
}
