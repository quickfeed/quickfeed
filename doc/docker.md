# Notes on Using Docker

QuickFeed builds and runs student submitted code within a Docker container.
After a container exits, the output is parsed and saved as a new submission entry in the database.
The following provides a brief overview of Docker and how it is used in QuickFeed.

To run Docker commands the user must be a part of the docker group.
See the section [Adding a User to the Docker Group](#adding-a-user-to-the-docker-group) for details.
This does not apply to the Docker Desktop application, available for Windows and macOS.

## Freeing up Space Consumed by Docker Containers

If we run out of space on the production server, there are a few Docker-related steps that can be taken:

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

The following commands are useful for dealing with Docker containers and too many open file descriptors:

```sh
# Check the number of open file descriptors for the quickfeed process (linux only)
% pgrep quickfeed | ls /proc/$(xargs)/fd | wc -l
% docker ps -a
% docker stats
# Delete all stopped containers
% docker rm $(docker ps -q -f status=exited)
```

## Adding a User to the Docker Group

We provide a script to automate the process of adding a user to the Docker group on Linux with systemd.
This script is located in the `scripts` directory.

To run the script, execute the following command:

```sh
% bash ./doc/scripts/setup-docker-group.sh
```

For details, see [setup-docker-group.sh](./scripts/setup-docker-group.sh)

### Manual Steps to Add a User to the Docker Group

First check that the `docker` group exists:

```sh
% cat /etc/group | grep docker
# or
% getent group docker
```

These commands can also be used to check which users are in the `docker` group, and can run Docker containers.

#### Missing Docker Group

If there isn't a `docker` group, add it manually with the command:

```sh
% sudo groupadd docker
```

After executing this command, you will need to restart the Docker daemon with the command:

```sh
% sudo systemctl restart docker.service
# or
% sudo service docker restart
```

#### Docker Group Exists

If the `docker` group exists, check if the user is already in the `docker` group with the command:

```sh
% groups [username]
```

If the `docker` group is not listed, the user is not in the `docker` group.
To add the user to the `docker` group, use the command:

```sh
% sudo usermod -aG docker [username]
```

Also make sure that the Docker daemon is running,

```sh
% sudo systemctl status docker.service
# or
% sudo service docker status
```

Check that you can run Docker commands without `sudo`:

```sh
% docker run hello-world
```
