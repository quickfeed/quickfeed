package kube_test

import (
	"context"
	"crypto/rand"
	"crypto/sha1"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/autograde/aguis/ci"
	"github.com/autograde/aguis/ci/kube"
	apiv1 "k8s.io/api/core/v1"
)

var (
	home       = homeDir()
	kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	course     = "agcicd"
	m          sync.Mutex
)

func init() { //TODO kube clinet
}

func newKubeCI() *kube.K8s {
	return &kube.K8s{}
}

func newTest(script, wantOut string) *test {
	t := &test{}
	t.script = script
	t.wantOut = wantOut
	return t
}

func newPod() *apiv1.Pod {
	return &apiv1.Pod{}
}

type test struct {
	script, wantOut, out string
}

func TestK8s(t *testing.T) {
	testK8s(t, "Hallo World")
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
		BaseImage:    "golang",
		ContainerCmd: []string{script},
	}

	k := newKubeCI()
	out, err := k.RunKubeJob(context.Background(), container, course, time.Now().Format("20060102-150405-")+echo, kubeconfig)
	if err != nil {
		t.Fatal(err)
	}

	if out != wantOut {
		t.Errorf("have %#v want %#v", out, wantOut)
	}
}
func TestK8sZero(t *testing.T) {
	const (
		script  = `echo -n "hello world 0"`
		wantOut = "hello world 0"
	)

	container := &kube.PodContainer{
		BaseImage:    "golang",
		ContainerCmd: []string{script},
	}

	k := newKubeCI()
	out, err := k.RunKubeJob(context.Background(), container, "agcicd", time.Now().Format("20060102-150405-99999999"), kubeconfig)
	if err != nil {
		t.Fatal(err)
	}

	if out != wantOut {
		t.Errorf("have %#v want %#v", out, wantOut)
	}
}
func randomSecret() string {
	randomness := make([]byte, 10)
	_, err := rand.Read(randomness)
	if err != nil {
		log.Fatal("couldn't generate randomness")
	}
	return fmt.Sprintf("%x", sha1.Sum(randomness))
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
		CreatorAccessToken: "",
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
		BaseImage:    "golang",
		ContainerCmd: script,
	}

	k := newKubeCI()
	out, err := k.RunKubeJob(context.Background(), container, "agcicd", jobName, kubeconfig)
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

func TestSequentials1(t *testing.T) {
	testSequentialK8s(t, 1)
}
func TestSequentials2(t *testing.T) {
	testSequentialK8s(t, 2)
}
func TestSequentials3(t *testing.T) {
	testSequentialK8s(t, 3)
}
func TestSequentials4(t *testing.T) {
	testSequentialK8s(t, 4)
}

func testSequentialK8s(t *testing.T, j int) {
	numberOfPods := 10
	tests := make([]test, numberOfPods)

	for i := 0; i < numberOfPods; i++ {
		t := newTest(`echo -n "`+strconv.Itoa(i)+`"`, strconv.Itoa(i))
		tests[i] = *t
	}

	for i := 0; i < numberOfPods; i++ {
		tm := "ci" + time.Now().Format("20060102-150405-") + strconv.Itoa(i) + strconv.Itoa(j)

		k := newKubeCI()
		s := tests[i].script
		out, _ := k.RunKubeJob(context.Background(),
			&kube.PodContainer{
				BaseImage:    "golang",
				ContainerCmd: []string{s},
			},
			course,
			tm, kubeconfig)

		tests[i].out = out
		fmt.Println(out)
	}

	for i := 0; i < numberOfPods; i++ {
		tst := tests[i]
		if tst.out != tst.wantOut {
			t.Errorf("have %#v want %#v", tst.out, tst.wantOut)
		}
	}
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}
