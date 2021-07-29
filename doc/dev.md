# QuickFeed Developer Guide

This developer guide assumes that you have [installed and configured QuickFeed](./install.md) and its dependent components.

## Technology stack

- [Go](https://golang.org/doc/code.html)
- [TypeScript](https://www.typescriptlang.org/)
- [gRPC](https://grpc.io/)
- [Protocol Buffers](https://developers.google.com/protocol-buffers/docs/proto3)
- [gRPC-Web](https://github.com/grpc/grpc-web)
- [Envoy](https://www.envoyproxy.io/)
- [NGINX](https://www.nginx.com/resources/wiki/)

## GitHub Integration

- [GitHub Application Setup](./github.md)
- [Setting up Course Organization](./teacher.md)

## Tools

QuickFeed provides a few command line tools.
See [cmd/scm/README.md](cmd/scm/README.md) for documentation of the SCM tool.

## Makefile

The Makefile provides various targets that aims to simplify running various tasks, such as compiling, updating and starting the server.
The Makefile defines variables, that can be altered depending on your environment, such as port numbers, organization name, and URIs.

### Compiling Targets

Following changes to `ag/ag.proto`, you need to recompile both the frontend and backend code by running:

```sh
make proto
```

To recompile and install the QuickFeed server, run:

```sh
make install
```

To recompile a new `bundle.js` for the QuickFeed browser-based client, run:

```sh
make ui
```

### Proxy

Currently, QuickFeed depends on two different proxies.
Envoy is used as a reverse proxy to facilitate gRPC invocations on the server-side.
NGINX acts as a web server endpoint and

<!-- TODO: Do we always want to purge before build? If so, merge the two targets in Makefile. -->
To rebuild the Envoy docker container:

```sh
make envoy-purge
make envoy-build
```

<!-- TODO: Do we need this target, since QuickFeed starts its own instance? -->
The following should not be needed, since the QuickFeed server starts its own Envoy instance.
However, you can also run the Envoy proxy manually:

```sh
make envoy-run
```

To reload NGINX after updating its configuration:

```sh
make nginx
```

### Testing

To run all tests that does not require remote interactions.

```sh
make test
```

To run specific tests that requires remote interactions with GitHub you must create a personal access token and assign it to `GITHUB_ACCESS_TOKEN`:

```sh
export GITHUB_ACCESS_TOKEN=<your-personal-access-token>
```

Some other tests may also require access to a specific test course organization; for these tests use `QF_TEST_ORG`:

```sh
export QF_TEST_ORG=<your-test-course-organization>
```

Here are some examples of such tests:

```sh
cd assignments
go test -v -run TestFetchAssignments
cd web
go test -v -run XXX
cd ci
QF_TEST_ORG=qf101 go test -v -run TestRunTests
TEST_TMPL=1 go test -v -run TestParseScript
TEST_IMAGE=1 go test -v -run TestParseScript
cd scm
go test -v -run TestGetOrganization
go test -v -run TestListHooks
QF_WEBHOOK=1 go test -v -run TestCreateHook
go test -v -run TestListHooks
```

### Utility

`make local` and `make remote` will switch where and how the gRPC client is being run, then will recompile the frontend. Use `make local` when running server on localhost with port forwarding, otherwise use `make remote`.

**Warning:** never push code with local gRPC client settings to the `quickfeed` repository, it will cause the server to stop responding to client requests. If this happens, just run `make remote` on the server location.

## Server architecture

### Default setup

By default, the gRPC server will be started at port **:9090**.
A docker container with Envoy proxy will listen on port **:8080** and redirect gRPC traffic to the gRPC server.
Webserver is running on one of internal ports, and NGINX, serving the static content, is set up to redirect HTTP traffic to that port, and all gRPC traffic to the port **:8080** (same port Envoy proxy is listening on).
NGINX and Envoy take care of all the relevant headers for gRPC traffic.

## Envoy

Envoy proxy allows making gRPC calls from a browser application.

### Basic configuration

[Default configuration from grpc-web repository](https://github.com/grpc/grpc-web/blob/master/net/grpc/gateway/examples/echo/envoy.yaml)
The main difference in [our configuration](https://github.com/autograde/quickfeed/blob/grpc-web-merge/envoy/envoy.yaml) is `http_protocol_options: { accept_http_10: true }` line inside HTTP filters list, and an additional header name.

## Errors and logging

Application errors can be classified into several groups and handled in different ways.

1. Database errors
Return generic "not found/failed" error message to user, log the original error.

2. SCM errors
Some of these can only be fixed by the user who is calling the method by interacting with UI elements (usually course teacher).
Example: if a github organization cannot be found, one of the possible issues causing this behavior is disabled third party access. As a result, the requested organization cannot be seen by the application. If a GitHub repo or team cannot be found, they could have been manually deleted from GitHub. Only the current user can remedy the situation, and it is most useful to inform them about the issue in detail and offer a solution.
Another example: sometimes GitHub interactions take too long and the request context can be canceled by GitHub. This is not an application failure, and neither developer team, nor user can actively do anything to correct it. But it is still useful to inform the user that the action must be repeated at later time to succeed.
Return a custom error with method name and original error for logging, and a custom message string to be displayed to user.

3. Access control errors
Return generic "access denied" error message to user.

4. API errors (invalid requests)
Return generic "malformed request" message to user.

5. GitHub API errors (request struct has missing/malformed fields)
Return a custom error with detailed information for logging and generic "action failed" message to user.

6. GitHub Cancelled Context errors
A sporadic error happening on the GitHub side, known to disappear on its own. The only solution is to wait and repeat the action later. Check for this type of errors and inform the user.

[GRPC status codes](https://github.com/grpc/grpc/blob/master/doc/statuscodes.md) are used to allow the client to check whether the error message should be displayed for user, or just logged for developers.

When the client is supposed to show error message to the user, error from the server will have status 9 (gRPC status "failed precondition")

### Backend

Errors are being logged at `Autograder Service` level. All other methods called from there (including database and scm methods) will just wrap and return all error messages directly. Introduce logging on layers deeper than `Autograder Service` only if necessary.

Errors returned to user interface must be few and informative, yet should not provide too many information about server routines. User must only be informed about details he or she can do something about.

### Frontend

When receiving response from the server, response status code is checked on the frontend. Any message with code different from 0 (0 is gRPC status code `OK`) will be logged to console. Relevant error messages will be displayed to user on course and group creation, and user and group enrollment updates.

[gRPC status codes](https://github.com/grpc/grpc/blob/master/doc/statuscodes.md)

## GitHub API

For GitHub integration we are using [Go implementation](https://github.com/google/go-github/tree/master/github) of [GitHub API](https://developer.github.com/v3/)

### Webhooks

- GitHub [Webhooks API](https://developer.github.com/webhooks/) is used for building and testing of code submitted by students.
- webhook is created automatically on course creation. It will react to every push event to any of course organization's repositories.
- depending on the repository the push event is coming from, assignment information will be updated in the QuickFeed's database, or a docker container with a student solution code will be built
- `name` field for any GitHub webhook is always "web"
- webhook will be using the same callback URL you have provided to the QuickFeed OAuth2 application and in the server startup command

### User roles/access levels for organization / team / repository

- GitHub API name for organization owner is `admin`
- Repository access levels for any organization member in GitHub API calls are: `read`/`write`/`admin`/`none`
- Individual repository permission levels in GitHub API are: `pull`/`push`/`admin`

### Slugs

When retrieving team, organization or repository by name, GitHub expects a slugified string instead of a full name as displayed on the organization page.
For example, organization with a name like `QuickFeed Test Org` will have slugified name `quickfeed-test-org`.

[URL slugs explained](http://patterns.dataincubator.org/book/url-slug.html)

### Repositories

- `owner` field for any organization repository is a slugified name for that organization
- access policy:
  - on course creation - default repository access across the whole organization is set to `none`, which means that only the organization owners can see any private repository on that organization
  - when students enroll, they receive read/pull access to `assignments` repository and write/push access to a personal student repository as GitHub invitations to their registered GitHub email

### Teams

Student groups will have GitHub teams with the same name created in the course organization.
Group records in the QuickFeed's database will have references to the corresponding GitHub team ID's.

## Simple testing of UI with dummy data

Since this is a single page application, then you would only be able to navigate to the default index page, every other navigation is handle by javascript. The way this works at the server, is that every request to /app/ returns the index.html page.
Now to get around the problem, we have to use the built in navigation manager, navMan. To do this, go to the main page of the application and open developer tools in the browser. In the command line type in the following command debugData.navMan.navigateTo("/app/admin/courses/new"). You could replace the url with any other url as you like.

## Docker

QuickFeed application will build code submitted by students, and run tests provided by teachers inside docker containers.
An often encountered problem is Docker being unable to resolve DNS due to disabled public DNS.
If you get a build error like that:

```log
Docker execution failed{error 25 0  Error response from daemon: Get https://registry-1.docker.io/v2/: dial tcp: lookup registry-1.docker.io on [::1]:53: read udp [::1]:50111->[::1]:53: read: connection refused}
```

then it must be a DNS problem.

One of the solutions is to uncomment or change `DOCKER_OPTS` line in `/etc/default/docker` file, then restart Docker daemon with `service docker restart`.

[Problem description and possible solutions](https://development.robinwinslow.uk/2016/06/23/fix-docker-networking-dns/)

## npm

`npm install` (or `npm i`) no longer installs all dependencies with versions stated in `package-lock.json`, but will also attempt to load latest versions for all root packages. If you just want to install the package dependencies without altering your `package-lock.json`, run `npm ci` instead.

## Repairing database from backups

Given a current database `ag.db` and a backup `bak.db`, and we want to replace records in a table `users` of the `ag.db` with entries from the same table in `bak.db`.
The database you open first will be under the alias `main`.

```sql
sqlite3 ag.db
delete from users;
attach database '/full/path/bak.db' as backup;
insert into main.users select * from backup.users;
detach database backup;
```
