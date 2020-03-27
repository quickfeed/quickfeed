package kube_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/autograde/aguis/ci/kube"
	"go.uber.org/zap"
)

/* var (
	home       = homeDir()
	kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	course     = "agcicd"
	m          sync.Mutex
) */
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

	container := &kube.Container{
		Image:    "golang",
		Commands: []string{script},
	}
	jobName := time.Now().Format("20060102-150405-") + echo
	k := newKubeCI()
	out, err := k.KRun(context.Background(), container, course, jobName, jobName /* , kubeconfig */)
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

	//jobName := tea.Format("20060102-150405")
	jobName := "jobName"
	ass := &kube.AssignmentInfo{
		AssignmentName:     "lab5",
		Language:           "go",
		CreatorAccessToken: "c0e4b71f27145d0653225d6415f65e39e1ab0f7f",
		GetURL:             cloneURL,
		TestURL:            getURLTest,
		RawGetURL:          strings.TrimPrefix(strings.TrimSuffix(cloneURL, ".git"), "https://"),
		RawTestURL:         strings.TrimPrefix(strings.TrimSuffix(getURLTest, ".git"), "https://"),
		RandomSecret:       "59fd5fe1c4f741604c1beeab875b9c789d2a7c73", //jobName,
	}
	//jobdock, err := ci.ParseScriptTemplate("", ass)         ///root/work/aguisforYannic/aguis/ci/scripts
	jobdock, err := kube.ParseKubeScriptTemplate("", ass) ///root/work/aguisforYannic/aguis/ci/scripts
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
	out, err := k.KRun(context.Background(), container, jobName, "agcicd", "59fd5fe1c4f741604c1beeab875b9c789d2a7c73" /* , kubeconfig */)
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

func TestK8sFPSecret(t *testing.T) {
	tea := time.Now()
	fmt.Println(tea.Format("20060102-150405"))
	cloneURL := "https://github.com/dat320-2019/assignments.git"
	getURLTest := "https://github.com/dat320-2019/tests.git"

	//jobName := tea.Format("20060102-150405")
	jobName := "jobName"

	ass := &kube.AssignmentInfo{
		AssignmentName:     "lab5",
		Language:           "go",
		CreatorAccessToken: "c0e4b71f27145d0653225d6415f65e39e1ab0f7f",
		GetURL:             cloneURL,
		TestURL:            getURLTest,
		RawGetURL:          strings.TrimPrefix(strings.TrimSuffix(cloneURL, ".git"), "https://"),
		RawTestURL:         strings.TrimPrefix(strings.TrimSuffix(getURLTest, ".git"), "https://"),
		RandomSecret:       "59fd5fe1c4f741604c1beeab875b9c789d2a7c73", //jobName,
	}
	jobdock, err := kube.ParseKubeScriptTemplate("", ass) ///root/work/aguisforYannic/aguis/ci/scripts
	if err != nil {
		panic(err)
	}
	script := jobdock.Commands
	container := &kube.Container{
		Image:    "golang",
		Commands: script,
	}

	k := newKubeCI()
	out, err := k.KRun(context.Background(), container, jobName, "agcicd", "59fd5fe1c4f741604c1beeab875b9c789d2a7c73")
	if err != nil {
		t.Fatal(err)
	}

	res, err := kube.ExtractKubeResult(zap.NewNop().Sugar(), out, "59fd5fe1c4f741604c1beeab875b9c789d2a7c73", 10)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(res.BuildInfo.BuildLog, "59fd5fe1c4f741604c1beeab875b9c789d2a7c73") {
		fmt.Println(out)
		t.Fatal("build log contains secret")
		t.Logf("res %+v", res.BuildInfo)
	}
	fmt.Println(out)
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}
