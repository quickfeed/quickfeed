package kube_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/autograde/aguis/ci"
	"github.com/autograde/aguis/ci/kube"
	"github.com/docker/docker/client"
)

var docker bool
var host, version string

func init() {
	host = envString("DOCKER_HOST", "http://localhost:4243")
	version = envString("DOCKER_VERSION", "1.39")

	 dt:= os.Getenv("DOCKER_TESTS")
	if dt != "" {
		 docker = true
		fmt.Println(dt)
	}

	fmt.Println("dt "+ dt)

	cli, err := client.NewClient(host, version, nil, nil)
	if err != nil {
		docker = false
		fmt.Println("false 1")

	}
	if _, err := cli.Ping(context.Background()); err != nil {
		docker = false
		panic(err)
	}
}

func newKubeCI() *kube.K8s {
	return &kube.K8s{
		Endpoint: host,
		Version:  version,
	}
}

func TestK8s(t *testing.T) {
	if !docker {
		t.SkipNow()
	}
	fmt.Println("testiii")
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

func envString(env, fallback string) string {
	e := os.Getenv(env)
	if e == "" {
		return fallback
	}
	return e
}
