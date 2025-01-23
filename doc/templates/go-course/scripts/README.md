# Debugging and Testing the Dockerfile (for dat520)

The main command is this one:

```console
docker run -it --rm -v /tmp/go-mod-cache:/quickfeed-go-mod-cache -v ~/work/distsys/2025/meling-labs:/quickfeed dat520:latest /bin/bash
```

This provides a temporary directory for the go mod cache and mounts a course repo into the image.
These are mapped to these locations in the image:

- `/quickfeed-go-mod-cache`
- `/quickfeed`

Note that the quickfeed server will set up these mounts automatically when running the tests.

Below are some relevant commands to test the image.

```console
# Build the image
$ cd scripts
$ docker build -t dat520:latest .
# Check that the image is built
$ docker images
# Make a temporary directory for the go mod cache
$ mkdir /tmp/go-mod-cache
# Run the image interactively and remove it when done
# Mount the go mod cache and the course repo (~/work/distsys/2025/meling-labs)
$ docker run -it --rm -v /tmp/go-mod-cache:/quickfeed-go-mod-cache -v ~/work/distsys/2025/meling-labs:/quickfeed dat520:latest /bin/bash
# Inside the container
$ alias l='ls -la'
# Poke around in the image
$ l /quickfeed
$ l /quickfeed-go-mod-cache
# Check the go version and golangci-lint version
$ which golangci-lint
$ golangci-lint --version
$ which go
$ go version
# Move to the uecho lab and run the tests
$ cd /quickfeed/lab1/uecho
$ go test -v
go: downloading golang.org/x/text v0.21.0
go: downloading golang.org/x/sync v0.10.0
=== RUN   TestEchoServerProtocol
2025/01/17 21:41:48 waiting connection...
--- PASS: TestEchoServerProtocol (0.00s)
=== RUN   TestPreAndPostSetup
2025/01/17 21:41:48 waiting connection...
--- PASS: TestPreAndPostSetup (0.00s)
=== RUN   TestMalformedRequest
2025/01/17 21:41:48 waiting connection...
--- PASS: TestMalformedRequest (0.00s)
PASS
ok  	dat520/lab1/uecho	0.004s
# Check that the go mod cache is populated
$ ls /quickfeed-go-mod-cache/
cache       golang.org
$ exit
```
