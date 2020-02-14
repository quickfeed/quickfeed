package kube_test

import (
	"context"
	"testing"

	"github.com/autograde/aguis/ci"
	"github.com/autograde/aguis/ci/kube"
)

//var docker bool
//var host, version string

func init() {
	//TODO kube clinet 
}

func newKubeCI() *kube.K8s {
	return &kube.K8s{}
}

func TestK8s(t *testing.T) {
	const (
		script  = `echo -n "hello world"`
		wantOut = "hello world"
	)

	k := newKubeCI()
	out, err := k.RunKubeJob(context.Background(),
		&ci.Job{
			Image:    "golang",
			Commands: []string{script},
		},
		"")

	if err != nil {
		t.Fatal(err)
	}

	if out != wantOut {
		t.Errorf("have %#v want %#v", out, wantOut)
	}
}
