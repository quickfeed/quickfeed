package kube_test

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/autograde/aguis/ci"
	"github.com/autograde/aguis/ci/kube"
	apiv1 "k8s.io/api/core/v1"
)

//dummy comment
//var KUBERNTES_HOSTNMAE + PORT nr string
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

	job := &ci.Job{
		Image:    "golang",
		Commands: []string{script},
	}

	k := newKubeCI()
	out, err := k.RunKubeJob(context.Background(), job, course, time.Now().Format("20060102-150405"), kubeconfig)
	if err != nil {
		t.Fatal(err)
	}

	if out != wantOut {
		t.Errorf("have %#v want %#v", out, wantOut)
	} else {
		fmt.Println(wantOut)
	}

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
			&ci.Job{
				Image:    "golang",
				Commands: []string{s},
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

/*func setupEnv(t *testing.T, namespace string) (*kubernetes.Clientset, *kube.K8s) {
	const (
		script  = `echo -n "hello world"`
		wantOut = "hello world"
	)

	job := &ci.Job{
		Image:    "golang",
		Commands: []string{script},
	}

	k := newKubeCI()
	out, err := k.RunKubeJob(context.Background(), job, namespace, kubeconfig)
	if err != nil {
		t.Fatal(err)
		fmt.Println(out)
	}

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		t.Errorf(err.Error())
		return nil, nil
	}
	//K8s clinet
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		t.Errorf(err.Error())
		return nil, nil
	}
	return clientset, k
}*/

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}
