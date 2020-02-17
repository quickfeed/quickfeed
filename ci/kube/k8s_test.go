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
	host = envString("DOCKER_HOST", "http://localhost:4242")
	version = envString("DOCKER_VERSION", "1.39")

	docker = true
	if os.Getenv("DOCKER_TESTS") != "" {
		docker = true
		fmt.Println("true")
	}

	cli, err := client.NewClient(host, version, nil, nil)
	if err != nil {
		docker = false
		fmt.Println("false 1")
	}

	fmt.Println("host: " + host + "\tversion: " + version)

	fmt.Println(cli)

	if _, err := cli.Ping(context.Background()); err != nil {
		docker = false
		fmt.Println("false 2")
		fmt.Println(err)
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
		//t.Fatal(err)
	}

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

func envString(env, fallback string) string {
	e := os.Getenv(env)
	if e == "" {
		return fallback
	}
	return e
}
