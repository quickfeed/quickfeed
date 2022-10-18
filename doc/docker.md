# Notes on Using Docker

Code submitted by students is being built and run inside Docker containers.
After a container exits, the output is parsed and saved as a new submission entry in the database.

If suddenly out of space on the production server, there are few Docker-related steps that can be taken:

- Check if there are containers running for too long with `docker ps`; if necessary, kill them with `docker rm <name/id>`
- Check if there are too many stopped containers waiting to be removed
- Show all running and stopped containers with `docker ps -a`
- To remove all stopped containers, use `docker container prune`
- Restart Docker daemon with `sudo service docker restart`
- Clean up all unused Docker objects with `docker system prune` (warning: can take a few minutes)

To be able to run docker, the user that is running docker either has to be running as sudo user, or must be a part of the docker group.

## Linux

First check that the `docker` group exists:

```console
% cat /etc/group | grep docker
```

This command can also be used to check which users are in the docker group, and thus can run docker containers.

To stop (or delete) **all containers**, use one of these commands:

```console
% docker stop $(docker ps -a -q)
% docker rm $(docker ps -a -q)
```

### Useful Docker Commands

```console
% docker images
```

### Helpful Tools for dealing with docker containers and too many open file descriptors

```console
% pgrep quickfeed | ls /proc/$(xargs)/fd | wc -l
% docker ps -a
% docker stats
% docker rm $(docker ps -q -f status=exited)
```

### Missing docker group

If there is no docker group, add it manually with the command:

```console
% sudo groupadd docker
```

After this command is executed please restart your machine or restart the docker daemon with the commands:

```console
% sudo systemctl restart docker.service
```

or

```console
% sudo service docker restart
```

### Docker group exists

If it does add the user that should be running docker to this group with the given command

```console
% sudo usermod -aG docker [username]
```

Also make sure that the docker daemon is running,

```console
% systemctl status docker.service
```

or with the command

```console
% service docker status
```
