# Running in docker

```
    Port: 4243
    Address: 0.0.0.0
```

To be able to run in docker, the docker deamon has to be started on a TCP port instead of a unix socket. First the docker deamon has to be stopped if it is running, and could be done with.

```
service docker stop
```

Then start the docker deamon in a terminal instance manully.

```
dockerd -H localhost:4243
```

## Connecting the docker client to the deamon

If you want to now connect the docker client `docker` to the docker deamon, the extra flag `-H localhost:4243` has to be configured as well.


```
docker -H localhost:4243 ps
```