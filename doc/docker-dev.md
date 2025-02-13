# Setup Docker development environment

## Create github app (required)

**IMPORTANT**: The following is required for running Quickfeed directly or with a Docker container

To create github app for local development create .env files for root and public folder, then run:

```sh
quickfeed -dev -new -secret
```

## Docker

View [Docker docs](<https://docs.docker.com/>) to learn about Docker, and please view the [dockerfile](/dockerfile) to understand how the image is constructed.

The combination of [Air-verse](https://github.com/air-verse/air) and Quickfeed's `watch` flag, enables live-reload for both front- and backend. A volume or bind mount is required for running a container since only `go.mod` and `go.sum` in root and kit folder is copied into the image.

The environment created in the image has every dependency Quickfeed requires, and will improve your development experience.

## Run Quickfeed with Docker

### Docker Desktop

**It is important** to map the Docker port to 443 - default https port, a different https port causes issues on the callback from the github app.

**To run a container with volume**, please use one of the following methods to create the container. Note that Docker Desktop doesn't provide a way to start a container with a volume, but you can configure a bind mount as an alternative.

### Docker-compose

To setup environment, run:

```sh
docker-compose up
```

Note `docker-compose up` will create an image if it does not exist, and create a container with a volume.

To create/update image:

```sh
docker-compose build
```

View [docker-compose file](/docker-compose.yml), and run `docker-compose --help` for more details

### Docker CLI

To create an image, run:

```sh
docker build -t quickfeed-web .
```

#### Volume

To create a volume bound to the Quickfeed directory, run:

```sh
docker volume create --driver local --opt type=none --opt device=*/path/to/quickfeed* --opt o=bind quickfeed_vol
```

Note that the [volume](https://docs.docker.com/engine/storage/volumes/) is configured to operate like a [bind mount](https://docs.docker.com/engine/storage/bind-mounts/), with the primary difference being the indirect connection; **container** -> **volume folder** -> **Quickfeed folder**. Additionally, the volume is persistent and can be reused, whereas a bind mount is typically one-time use.

#### Create a container

To create with a volume, run:

```sh
docker run -p 443:443 --mount src=quickfeed_vol,dst=/quickfeed quickfeed-web
```

To create with bind mount, run:

```sh
docker run -p 443:443 --mount type=bind,src=*/path/to/quickfeed*,dst=/quickfeed quickfeed-web
```

#### Useful Commands

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

There are more CLI commands, please view `docker --help` or [Notes on using docker](./docker.md)

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

Docker desktop is not running on your computer. `Docker-compose` requires the Docker engine to be active.

### Please go back to [step one](#create-github-app-required) if you got any of the following issues

- open /app/.env: no such file or directory
- Required QUICKFEED_AUTH_SECRET is not set
- missing application ID for provider github
