package kube_test

import (
	"context"
	"strconv"
	"testing"

	"github.com/autograde/aguis/ci"
	"github.com/autograde/aguis/ci/kube"
)

//var KUBERNTES_HOSTNMAE + PORT nr string

func init() {
	//TODO kube clinet
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

func TestK8s(t *testing.T) {
	const (
		script  = `echo -n "hello world"`
		wantOut = "hello world"
	)

	job := &ci.Job{
		Image:    "golang",
		Commands: []string{script},
	}

	k := newKubeCI()
	out, err := k.RunKubeJob(context.Background(), job, "")
	if err != nil {
		t.Fatal(err)
	}

	if out != wantOut {
		t.Errorf("have %#v want %#v", out, wantOut)
	}
}

func TestParalellK8s(t *testing.T) {
	numberOfPods := 10
	tests := make([]test, numberOfPods)

	for i := 0; i < numberOfPods; i++ {
		t := newTest(`echo -n "`+strconv.Itoa(i)+`"`, strconv.Itoa(i))
		tests[i] = *t
	}

	for i := 0; i < numberOfPods; i++ {
		tst := tests[i]
		k := newKubeCI()
		out, err := k.RunKubeJob(context.Background(),
			&ci.Job{
				Image:    "golang",
				Commands: []string{tst.script},
			},
			"")

		if err != nil {
			t.Fatal(err)
		}
		tst.out = out
	}

	for i := 0; i < numberOfPods; i++ {
		tst := tests[i]
		if tst.out != tst.wantOut {
			t.Errorf("have %#v want %#v", tst.out, tst.wantOut)
		}
	}

}
