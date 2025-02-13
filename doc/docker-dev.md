# Setup development environment

## Create github app (required)

**IMPORTANT**: The following is required for running Quickfeed directly or with a docker container

To create github app for local development create .env files for root and public folder, then run:

```sh
quickfeed -dev -new -secret
```

## Docker

Useful sources: [Docker docs](<https://docs.docker.com/>), [dockerfile](/dockerfile), [setup docker desktop for WSL 2](<https://docs.docker.com/desktop/features/wsl/>), [Guide for installing WSL 2](<https://learn.microsoft.com/en-us/windows/wsl/install>)

Docker is executed with [air-verse](https://github.com/air-verse/air) and [volumes](https://docs.docker.com/engine/storage/volumes/), utilizing the `-watch` flag for Quickfeed, enabling live-reload.

## Run Quickfeed with Docker

### Use docker desktop

**It is important** to map the docker port to 443 - default https port, a different https port causes issues on the callback from the github app.

**To run a container with volume**, please use one of the following methods to create the container. The docker desktop doesn't provide a way of starting a container with a volume.

### Use docker-compose

To create/update image:

```sh
docker-compose build
```

To create a container, run:

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

#### Volume

To create a volume bound to the Quickfeed directory, run:

```sh
docker volume create --driver local --opt type=none --opt device=$(pwd) --opt o=bind quickfeed_vol
```

#### Create a container

With a volume, run:

```sh
docker run -p 443:443 --mount src=quickfeed_vol,dst=/quickfeed quickfeed-web
```

With bind mount, run:

```sh
docker run -p 443:443 --mount type=bind,src=$(pwd),dst=/quickfeed quickfeed-web
```

#### Commands

To show all containers, run:

```sh
docker ps -a
```

To sh into a container, run:

```sh
docker exec -it *ID/Name* sh
```

To view details, run:

```sh
docker inspect *ID/Name*
```

There are more useful CLI commands, please view `docker --help` or [Notes on using docker](./docker.md)

### Issues and warnings

#### To suppress warning: **The "QUICKFEED" variable is not set. Defaulting to a blank string**, run

```sh
export QUICKFEED=/path/to/quickfeed-repository
```

The variable can be any string to suppress it, but must be a valid path to the repository for Quickfeed to function.

To persist the suppress, run:

```sh
echo 'export QUICKFEED=' >> $HOME/.bashrc
```

Set the environment variable to an empty string.

### The command 'docker-compose' could not be found

Docker desktop is not running on your computer. `Docker-compose` requires the docker engine to be active.

### Please go back to [step one](#create-github-app-required) if you got any of the following issues

- open /app/.env: no such file or directory
- Required QUICKFEED_AUTH_SECRET is not set
- missing application ID for provider github
