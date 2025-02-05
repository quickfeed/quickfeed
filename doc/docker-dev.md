# Setup development environment

## Create github app (required)

**IMPORTANT**: The following is required for running quickfeed in DEV mode directly or through a docker container

To create github app for local development, run:

```bash
make new-githubApp
```

## Run quickfeed in development mode

```bash
quickfeed -dev
```

## Docker

Useful sources: [Docker docs](<https://docs.docker.com/>), [View the dockerfile](/dockerfile)

Docker is executed with [air-verse](https://github.com/air-verse/air) and [volumes](https://docs.docker.com/engine/storage/volumes/), enabling live-reload.

To install air, run:

```sh
go install github.com/air-verse/air@latest
```

### Run quickfeed with Docker

#### Use docker desktop through the App/interface

**It is important** to map the docker port to 443 (default https port, different https port causes issues on the callback from the github app). The app does internally use 443 by default which forces you to use the port 443, you have to set the flag -http.addr manually if you'd like to run the container on a different port (btw will cause issues on callback), and keep in mind that you have to create a new image, and remember which port the application runs on (the -http.addr flag in the CMD, [here](/dockerfile#L27))

#### Use docker desktop through CLI

[Setup docker desktop for WSL 2](<https://docs.docker.com/desktop/features/wsl/>) - [Guide for installing WSL 2](<https://learn.microsoft.com/en-us/windows/wsl/install>)

To create image and create container, run:

```bash
docker-compose up --build
```

[View docker-compose file](/docker-compose.yml)

To create container of already created image, run:

```bash
docker-compose up
```

Run `docker-compose --help` for more details

### Issues and warnings

### To suppress warning: **The "QUICKFEED" variable is not set. Defaulting to a blank string**

Run:

```bash
export QUICKFEED=/path/to/quickfeed-repository
```

The variable can be any string to suppress it, but needs to be a valid path for the repository if the quickfeed binary is in a different directory

### 'The command 'docker-compose' could not be found in this WSL 2 distro'

You are most likely getting this message because docker isn't running on your computer.

### '/usr/local/bin/docker-compose: line 1: Not: command not found'

Weird issue of most likely broken binary or non existent one.. advise you to restart wsl/linux.
