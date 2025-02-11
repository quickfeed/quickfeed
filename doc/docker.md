# Notes on Using Docker

Code submitted by students is being built and run inside Docker containers.
After a container exits, the output is parsed and saved as a new submission entry in the database.

To be able to run docker on linux, and WSL or macOS without docker desktop, the user that is running docker either has to be running as sudo user, execute commands with `sudo`, or must be a part of the docker group. See [Configure a docker group](#configure-a-docker-group) for details.

## Freeing up space on Docker

If suddenly out of space on the production server, there are few Docker-related steps that can be taken:

- Check if there are containers running for too long with `docker ps`; if necessary, kill them with `docker rm <name/id>`
- Check if there are too many stopped containers waiting to be removed
- Show all running and stopped containers with `docker ps -a`
- To remove all stopped containers, use `docker container prune`
- Restart Docker daemon with `sudo service docker restart`
- Clean up all unused Docker objects with `docker system prune` (warning: can take a few minutes)

## Useful Docker Commands

To display images:

```sh
% docker images
```

To display all containers:

```sh
% docker ps -a
```

To stop (or delete) **all containers**, use one of these commands:

```sh
% docker stop $(docker ps -a -q)
% docker rm $(docker ps -a -q)
```

## Helpful Tools for dealing with Docker containers and too many open file descriptors

```sh
% pgrep quickfeed | ls /proc/$(xargs)/fd | wc -l
% docker ps -a
% docker stats
% docker rm $(docker ps -q -f status=exited)
```

## Configure a Docker group

There are two sections for setting up a Docker group, a guide and a setup script, choose whichever works best for you.

Once a group is configured and a user gains access to Docker daemon, deleting the Docker group or removing the user from the group will revoke the access. However, the changes will only take effect after the user terminates their session.

### Setup script

To setup Docker daemon move to root directory and run:

```sh
% bash ./doc/scripts/setup-docker-group.sh
```

For details, see [setup-script](../setup-daemon.sh)

### Guide

First check that the `docker` group exists:

```sh
% cat /etc/group | grep docker
```

or:

```sh
% getent group docker
```

These commands can also be used to check which users are in the Docker group, and thus can run Docker containers.

#### Missing Docker group

If there isn't a Docker group, add it manually with the command:

```sh
% sudo groupadd docker
```

After this command is executed please restart your machine or restart the Docker daemon with the commands:

```sh
% sudo systemctl restart docker.service
```

or

```sh
% sudo service docker restart
```

#### Docker group exists

If it does add the user that should be running Docker to this group with the given command

```sh
% sudo usermod -aG docker [username]
```

Also make sure that the Docker daemon is running,

```sh
% systemctl status docker.service
```

or with the command

```sh
% service docker status
```
