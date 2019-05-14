package envoy

import (
	"context"
	"log"
	"os/exec"

	"github.com/docker/docker/api/types/filters"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

/* creates a Docker API client. If envoy container is not running, starts it from image.
If no image exists, pulls Envoy image from docker and builds it with envoy.yaml options */
func StartEnvoy() {

	ctx := context.Background()
	cli, err := client.NewEnvClient()
	if err != nil {
		log.Panicln("Envoy: docker client failed to start: ", err.Error())
	}

	// removes all stopped containers
	_, er := cli.ContainersPrune(ctx, filters.Args{})
	if er != nil {
		log.Println("Envoy: error attempting to prune unused containers: ", err.Error())
	}
	log.Println("Envoy: prunning unused containers. ")

	// looks at existing containers to check whether Envoy is already running
	containerRuns := false
	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		log.Println("Envoy: cannot retrieve docker container list: ", err.Error())
	}
	for i, container := range containers {
		if container.Names[0] == "/envoy" {
			log.Println("Envoy container is already running", i)
			containerRuns = true
		}
	}

	log.Println("Envoy: no container found, starting build...")
	if !containerRuns {
		images, err := cli.ImageList(ctx, types.ImageListOptions{})
		if err != nil {
			log.Panicln("Envoy: cannot retrieve docker image list: ", err.Error())
		}
		log.Println("Envoy: checking images")
		imgExists := false
		for _, img := range images {
			log.Println("Found image: ", img.RepoTags)
			if img.RepoTags[0] == "ag_envoy:latest" {
				log.Println("Envoy image already exists")
				imgExists = true
			}
		}
		// if there is no active Envoy image
		if !imgExists {
			log.Println("Envoy image building... ")
			out, err := exec.Command("/bin/sh", "./envoy/envoy.sh", "build").Output()
			log.Println("Envoy: started bash script with argument to build Envoy image, result: ", out)
			if err != nil {
				log.Println("Envoy: error when executing bash script: ", err.Error())
			}

		}
		log.Println("Envoy: starting container... ")
		out, err := exec.Command("/bin/sh", "./envoy/envoy.sh").Output()
		log.Println("Envoy: script resulted in: ", out)
		if err != nil {
			log.Println("Envoy: error when executing bash script: ", err.Error())
		}

	} else {
		log.Println("Envoy: done")
	}

	/*
		if !imgExists {
			img, err := cli.ImagePull(ctx, "envoyproxy/envoy:latest", types.ImagePullOptions{})
			if err != nil {
				log.Panicln("Envoy: cannot pull image: ", err.Error())
			}
			io.Copy(os.Stdout, img)

		}

		envoy, err := cli.ContainerCreate(ctx, &container.Config{
			Image:        "envoyproxy/envoy",
			ExposedPorts: nat.PortSet{"8080": struct{}{}},
		}, &container.HostConfig{
			PortBindings: map[nat.Port][]nat.PortBinding{nat.Port("8080"): {{HostIP: "127.0.0.1", HostPort: "8080"}}},
		}, nil, "ag_envoy")
		if err != nil {
			log.Panicln("Envoy: cannot create container: ", err.Error())
		}

		if err := cli.ContainerStart(ctx, envoy.ID, types.ContainerStartOptions{}); err != nil {
			log.Panicln("Envoy: container failed to start: ", err.Error())
		}
	*/

}
