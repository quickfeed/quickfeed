# Running in docker

To be able to run docker, the user that is running docker either has to be running as sudo user, or the current user has to be a part of the docker group.

## Linux

First check that the `docker` group exists:

```console
cat /etc/group | grep docker
```

This command can also be used to check which users are in the docker group, and thus can run docker containers.

To stop (or delete) **all containers**, use one of these commands:

```console
docker stop $(docker ps -a -q)
docker rm $(docker ps -a -q)
```

### Helpful Tools for dealing with docker containers and too many open file descriptors

```console
pgrep quickfeed | ls /proc/$(xargs)/fd | wc -l
docker ps -a
docker stats
docker rm $(docker ps -q -f status=exited)
```

### Missing docker group

If there is no docker group, add it manually with the command:

```console
sudo groupadd docker
```

After this command is executed please restart your machine or restart the docker daemon with the commands:

```console
sudo systemctl restart docker.service
```

or

```console
sudo service docker restart
```

### Docker group exists

If it does add the user that should be running docker to this group with the given command

```
sudo usermod -aG docker [username]
```

Also make sure that the docker daemon is running,

```
systemctl status docker.service
```

or with the command

```
service docker status
```

# Old method of running docker in autograder

```
    Port: 4243
    Address: 0.0.0.0
```

To be able to run in docker, the docker deamon has to be started on a TCP port instead of a unix socket. First the docker deamon has to be stopped if it is running.

## Linux with systemd (most distros)

Check status:

```
service docker status
```

Stopping the docker deamon

```
service docker stop
```

Then start the docker deamon in a terminal instance manully.

```
dockerd -H localhost:4243
```

## Connecting the docker client to the deamon

If you want to now connect the docker client `docker` to the docker deamon, the extra flag `-H localhost:4243` has to be configured as well.

The command:

```
docker ps
```

Has to add the extra `-H` flag to be able to run with the docker deamon running on port 4243

```
docker -H localhost:4243 ps
```
