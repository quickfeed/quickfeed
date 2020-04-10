package kube_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/autograde/aguis/ci"
	"github.com/autograde/aguis/ci/kube"
)

var (
	course = "agcicd"
	sec    = "59fd5fe1c4f741604c1beeab875b9c789d2a7c73"
)

//var scriptPath = "kube/kube_scripts"

func newKubeCI() *kube.K8s {
	return &kube.K8s{}
}

func newPodContainer(baseImage string, script []string) *kube.Container {
	return &kube.Container{
		Image:    baseImage,
		Commands: script,
	}
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

	container := &kube.Container{
		Image:    "golang",
		Commands: []string{script},
	}
	k := newKubeCI()
	out, err := k.KRun(context.Background(), container, jobName, course, sec)
	if err != nil {
		t.Fatal(err)

	}
	if out != wantOut {
		t.Errorf("have %#v want %#v", out, wantOut)
	}
}

func TestK8sFP(t *testing.T) {
	startTime := time.Now()
	fmt.Println(startTime.Format("20060102-150405"))
	jobName := startTime.Format("20060102-150405")
	info := getAssignmentInfo()
	jobdock, err := ci.ParseScriptTemplate("", info)
	if err != nil {
		panic(err)
	}
	wantOut := ""
	script := jobdock.Commands
	container := newPodContainer("golang", script)
	k := newKubeCI()

	out, err := k.KRun(context.Background(), container, jobName, course, sec)
	if err != nil {
		t.Fatal(err)
	}

	if out != wantOut {
		t.Errorf("have %#v want %#v", out, wantOut)
	} else {
		fmt.Println(wantOut)
	}
	fmt.Println(time.Since(startTime))
}

func getAssignmentInfo() *ci.AssignmentInfo {
	cloneURL := "https://github.com/dat320-2019/assignments.git"
	getURLTest := "https://github.com/dat320-2019/tests.git"

	info := &ci.AssignmentInfo{
		AssignmentName:     "lab5",
		Language:           "go",
		CreatorAccessToken: "",
		GetURL:             cloneURL,
		TestURL:            getURLTest,
		RawGetURL:          strings.TrimPrefix(strings.TrimSuffix(cloneURL, ".git"), "https://"),
		RawTestURL:         strings.TrimPrefix(strings.TrimSuffix(getURLTest, ".git"), "https://"),
		RandomSecret:       sec,
	}
	return info
}
