package kube_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/autograde/aguis/ci"
	"github.com/autograde/aguis/ci/kube"
)

/* var (
	home       = homeDir()
	kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	course     = "agcicd"
	m          sync.Mutex
) */
var course = "agcicd"

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

	container := &kube.PodContainer{
		Image:    "golang",
		Commands: []string{script},
	}

	k := newKubeCI()
	out, err := k.RunKubeJob(context.Background(), container, course, time.Now().Format("20060102-150405-")+echo /* , kubeconfig */)
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

	jobName := tea.Format("20060102-150405")

	ass := &ci.AssignmentInfo{
		AssignmentName:     "lab5",
		Language:           "go",
		CreatorAccessToken: "166f3712bbd32a6750c436244f74d031c0c91257",
		GetURL:             cloneURL,
		TestURL:            getURLTest,
		RawGetURL:          strings.TrimPrefix(strings.TrimSuffix(cloneURL, ".git"), "https://"),
		RawTestURL:         strings.TrimPrefix(strings.TrimSuffix(getURLTest, ".git"), "https://"),
		RandomSecret:       jobName,
	}
	jobdock, err := ci.ParseScriptTemplate("", ass) ///root/work/aguisforYannic/aguis/ci/scripts
	if err != nil {
		panic(err)
	}
	wantOut := ""
	script := jobdock.Commands

	container := &kube.PodContainer{
		Image:    "golang",
		Commands: script,
	}

	k := newKubeCI()
	out, err := k.RunKubeJob(context.Background(), container, "agcicd", jobName /* , kubeconfig */)
	fmt.Println(out)
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

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}
