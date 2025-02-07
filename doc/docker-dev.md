# Setup development environment

## Create github app (required)

**IMPORTANT**: The following is required for running quickfeed directly or with a docker container

To create github app for local development create .env files for root and public folder, then run:

```sh
quickfeed -dev -new -secret
```

## Docker

Useful sources: [Docker docs](<https://docs.docker.com/>), [dockerfile](/dockerfile), [setup docker desktop for WSL 2](<https://docs.docker.com/desktop/features/wsl/>), [Guide for installing WSL 2](<https://learn.microsoft.com/en-us/windows/wsl/install>)

Docker is executed with [air-verse](https://github.com/air-verse/air) and [volumes](https://docs.docker.com/engine/storage/volumes/), enabling live-reload.

To install air, run:

```sh
go install github.com/air-verse/air@latest
```

PS: required to run air directly, air is installed in the docker image

## Run quickfeed with Docker

### Use docker desktop

**It is important** to map the docker port to 443 (default https port, different https port causes issues on the callback from the github app). The app internally use port 443 by default which forces you to use the port 443, you have to set the flag -http.addr manually if you'd like to run the container on a different port (will cause issues on callback). TODO (joachim): Update when <https://github.com/quickfeed/quickfeed/issues/1195> is closed

**To run a container with volume**, please use one of the following methods to create the container. The docker desktop doesn't provide a way of starting a container with a volume.

### Use docker-compose

To create an image and/or create a container, run:

```sh
docker-compose up
```

Docker will build the image if it doesn't exist.

View [docker-compose file](/docker-compose.yml), and run `docker-compose --help` for more details

### Use docker CLI

To create an image, run:

```sh
docker build -t quickfeed-web .
```

To create a volume bound to the quickfeed directory, run:

```sh
docker volume create --driver local --opt type=none --opt device=. --opt o=bind vol
```

Vol is a partial name of the volume, which results in full name: quickfeed_vol, and `device` value: . is equal to current directory - `pwd`

To create a container with a volume, run:

```sh
docker run -p 443:443 --mount src=quickfeed_vol,dst=/app quickfeed-web
```

To create a container with bind mount, run:

```sh
docker run -p 443:443 --mount type=bind,src=$(pwd),dst=/app quickfeed-web
```

To show all containers, run:

```sh
docker container ls -a
```

To sh into a container, run:

```sh
docker exec -it *container ID* sh
```

Theres many more useful CLI commands, please view `docker --help`

### Issues and warnings

#### To suppress warning: **The "QUICKFEED" variable is not set. Defaulting to a blank string**, run

```sh
export QUICKFEED=/path/to/quickfeed-repository
```

The variable can be any string to suppress it, but needs to be a valid path for the repository if the quickfeed binary is in a different directory

To persist the suppress, run:

```sh
echo 'export QUICKFEED=' >> $HOME/.bashrc
```

### The command 'docker-compose' could not be found in this WSL 2 distro

You are most likely getting this message because docker isn't running on your computer.

### /usr/local/bin/docker-compose: line 1: Not: command not found

Weird issue of most likely broken binary or non existent one.. advise you to restart wsl/linux.

### Please go back to [step one](#create-github-app-required) if you got any of the following issues

- open /app/.env: no such file or directory
- Required QUICKFEED_AUTH_SECRET is not set
- missing application ID for provider github
