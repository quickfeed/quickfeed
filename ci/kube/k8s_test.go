package kube_test

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/autograde/aguis/ci"
	"github.com/autograde/aguis/ci/kube"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

//dummy comment
//var KUBERNTES_HOSTNMAE + PORT nr string
var (
	home       = homeDir()
	kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
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

type test struct {
	script, wantOut, out string
}

func TestK8sZero(t *testing.T) {
	const (
		//script  = `cat /root/work/secreting/aa; echo -n "hello world 0"`
		script  = `echo -n "hello world 0"`
		wantOut = "hello world 0"
	)

	job := &ci.Job{
		Image:    "golang",
		Commands: []string{script},
	}

	k := newKubeCI()
	out, err := k.RunKubeJob(context.Background(), job, "agcicd", time.Now().Format("20060102-150405-99999999"), kubeconfig)
	if err != nil {
		t.Fatal(err)
	}

	if out != wantOut {
		t.Errorf("have %#v want %#v", out, wantOut)
	}
}

<<<<<<< HEAD
func TestSequentielK8s(t *testing.T) {
=======
func TestK8sOne(t *testing.T) {
	const (
		//script  = `cat /root/work/secreting/aa; echo -n "hello world 0"`
		script  = `echo -n "hello world 1"`
		wantOut = "hello world 1"
	)

	job := &ci.Job{
		Image:    "golang",
		Commands: []string{script},
	}

	k := newKubeCI()
	out, err := k.RunKubeJob(context.Background(), job, "agcicd", time.Now().Format("20060102-150405-99999999"), kubeconfig)
	if err != nil {
		t.Fatal(err)
	}

	if out != wantOut {
		t.Errorf("have %#v want %#v", out, wantOut)
	}
}
func TestK8sTwo(t *testing.T) {
	const (
		//script  = `cat /root/work/secreting/aa; echo -n "hello world 0"`
		script  = `echo -n "hello world 2"`
		wantOut = "hello world 2"
	)

	job := &ci.Job{
		Image:    "golang",
		Commands: []string{script},
	}

	k := newKubeCI()
	out, err := k.RunKubeJob(context.Background(), job, "agcicd", time.Now().Format("20060102-150405-99999999"), kubeconfig)
	if err != nil {
		t.Fatal(err)
	}

	if out != wantOut {
		t.Errorf("have %#v want %#v", out, wantOut)
	}
}

func TestOneA(t *testing.T) {
>>>>>>> ag_dev
	numberOfPods := 10
	tests := make([]test, numberOfPods)

	for i := 0; i < numberOfPods; i++ {
		t := newTest(`echo -n "`+strconv.Itoa(i)+`"`, strconv.Itoa(i))
		tests[i] = *t
	}
<<<<<<< HEAD

	fmt.Println(tests)

	for i := 0; i < numberOfPods; i++ {
		tm := "ci" + strconv.Itoa(i)
		//tm := "ci-" + getTimeNow() + "-" + strconv.Itoa(i)

		k := newKubeCI()

		fmt.Println(tests[i])
		//err := newError()
		s := tests[i].script
		out, _ := k.RunKubeJob(context.Background(),
			&ci.Job{
				Image:    "golang",
				Commands: []string{s},
			},
			tm, kubeconfig)

		tests[i].out = out
		fmt.Println("Input value: ", s)
	}

=======
	var wg sync.WaitGroup
	for i := 0; i < numberOfPods; i++ {
		wg.Add(1)
		go func(i int) {
			tm := "cia" + strconv.Itoa(i)
			k := newKubeCI()
			m.Lock()
			s := tests[i].script
			out, _ := k.RunKubeJob(context.Background(),
				&ci.Job{
					Image:    "golang",
					Commands: []string{s},
				}, "agcicd",
				tm, kubeconfig)
			tests[i].out = out
			m.Unlock()
			wg.Done()
		}(i)
	}
	wg.Wait()
>>>>>>> ag_dev
	for i := 0; i < numberOfPods; i++ {
		tst := tests[i]
		if tst.out != tst.wantOut {
			t.Errorf("have %#v want %#v", tst.out, tst.wantOut)
		}
	}
}

func getTimeNow() string {
	return time.Now().Format("20060102-150405")
}

func TestDelete(t *testing.T) {
<<<<<<< HEAD
	namespace := getTimeNow() + "-delete"
	cs, k := setupEnv(t, namespace)
=======
	jobId := time.Now().Format("20060102-150405") + "-delete"
	cs, k := setupEnv(t, jobId)
>>>>>>> ag_dev
	k.DeleteObject(*cs, "agcicd")
}

func setupEnv(t *testing.T, jobId string) (*kubernetes.Clientset, *kube.K8s) {
	const (
		script  = `echo -n "hello world"`
		wantOut = "hello world"
	)

	job := &ci.Job{
		Image:    "golang",
		Commands: []string{script},
	}

	k := newKubeCI()
	out, err := k.RunKubeJob(context.Background(), job, "agcicd", jobId, kubeconfig)
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
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}
