# QuickFeed Developer Guide

This developer guide assumes that you have [installed and configured QuickFeed](./deploy.md) and its dependent components.

## GitHub Integration

- [GitHub Application Setup](./github.md)
- [Setting up Course Organization](./teacher.md)

## Tools

QuickFeed provides a few command line tools.
See [cmd/scm/README.md](cmd/scm/README.md) for documentation of the SCM tool.

## Makefile

The Makefile in QuickFeed simplifies various tasks like compiling, updating, and launching the server.

### Compiling Targets

After modifying qf/qf.proto, you need to recompile both frontend and backend. Use the following commands:

```sh
make proto
```

To recompile and install the QuickFeed server, run:

```sh
make install
```

To compile the browser client bundles:

```sh
make ui
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
cd ci
go test -v -run TestRunTests
cd scm
go test -v -run TestGetOrganization
go test -v -run TestListHooks
QF_WEBHOOK_SERVER=https://62b9b9c05ece.ngrok.io go test -v -run TestCreateHook
go test -v -run TestListHooks
cd web/hooks
QF_WEBHOOK_SERVER=https://62b9b9c05ece.ngrok.io go test -v -run TestGitHubWebHook
```

## Server architecture

### Default Configuration

TODO(meling) Update and improve this part. It is not correct anymore, I think.

- **Primary Server Port:** 
By default, the server runs on port **:443**, the standard port for HTTPS traffic. This ensures secure communication right out of the box.
- **Custom Port Configuration:** 
If you need to use a different port, you can easily change this by using the `-http.addr` flag when launching the server.
- **HTTP to HTTPS Redirection:** 
Alongside the main server, we also initiate a secondary server on port **:80**. Its sole purpose is to redirect all incoming HTTP requests to HTTPS. 
This ensures that even if someone attempts to connect via the unsecured HTTP protocol, their request will be automatically upgraded to a secure connection.

**Important Note:** 
When running servers on ports like **:80** or **:443**, some operating systems may require elevated permissions or specific configurations. 
This is because ports below 1024 are considered privileged ports, and running services on these ports might need administrative rights or special configurations.

In Linux, you can use the `setcap` command to allow a binary to bind to privileged ports without elevated permissions. 
For example, to allow the `quickfeed` binary to bind to port **:443**, you can run the following command:

```sh
sudo setcap CAP_NET_BIND_SERVICE=+eip /path/to/binary/quickfeed
```

Note that you will need to repeat this step each time you recompile the server.

Please make sure to check your operating system's documentation for the necessary steps to run services on these ports.

## Errors and logging

Application errors can be classified into several groups and handled in different ways:

**Database errors**:
  - Return generic "not found/failed" error message to the user, log the original error.

**SCM errors**:

- Some of these can only be fixed by the user who is calling the method by interacting with UI elements (usually course teacher).

  **Examples**:
  - If a GitHub organization cannot be found, one of the possible issues causing this behavior is not having installed the GitHub application on the organization.
  As a result, the requested organization cannot be seen by QuickFeed.
  - If a GitHub repository cannot be found, they could have been manually deleted from GitHub.
  Only the current user can remedy the situation, and it is most useful to inform them about the issue in detail and offer a solution.

- Sometimes GitHub interactions take too long and the request times out, or is otherwise cancelled by GitHub.
In these cases the error is usually ephemeral in nature, and the action should be repeated at later time. This should be communicated to the end user.

- Return a custom error with detailed information for logging, and a custom error message to the user.

**Access control errors**:

- Return generic "access denied" error message to the user.

**API errors (invalid requests)**:

- Return generic "malformed request" message to the user.

**GitHub API errors (request struct has missing/malformed fields)**

- Return a custom error with detailed information for logging and generic "failed precondition" message to the user.


[Connect Error Codes](https://connectrpc.com/docs/protocol#error-codes) are used to allow the client to check whether the error message should be displayed for user, or just logged for developers.

### Backend

Errors are being logged at `QuickFeed Service` level. 
All other methods called from there (including database and SCM methods) will just wrap and return all error messages directly. 
Introduce logging on layers deeper than `QuickFeed Service` only if necessary.

Errors returned to a user should be few and informative. 
They should not reveal internal details of the application.

### Frontend

When receiving a response from the server, the response status code is checked on the frontend. 
Any message with code different from 0 (0 is status code `OK`) will be logged to console. 
Error messages will be displayed to user where relevant, e.g. on course and group creation, and user and group enrollment updates.

[Connect Error Codes](https://connectrpc.com/docs/protocol#error-codes)

## GitHub API

For GitHub integration we are using [Go implementation](https://github.com/google/go-github) of [GitHub API](https://docs.github.com/en/rest)

### Webhooks

- GitHub [Webhooks API](https://docs.github.com/en/webhooks) is used for building and testing of code submitted by students.
- A webhook is created automatically when installing the GitHub App on a course organization. The webhook will be triggered by pushes to repositories in the organization.
- Push events from the `tests` repository may update
  - The assignment information in QuickFeed's database.
  - The Docker container and run.sh script used for building and testing student submitted code.
- Push events from the `username-labs` repositories may trigger text execution.
- The webhook will POST events to `$DOMAIN/hook/`, where `$DOMAIN` is the domain name of the server, as defined in your `.env` file.

### User roles/access levels for organization / repository

- GitHub API name for organization owner is `admin`
- Repository access levels for any organization member in GitHub API calls are: `read`/`write`/`admin`/`none`
- Individual repository permission levels in GitHub API are: `pull`/`push`/`admin`

### Slugs

When retrieving organization or repository by name, GitHub expects a slugified string instead of a full name as displayed on the organization page.
For example, organization with a name like `QuickFeed Test Org` will have slugified name `quickfeed-test-org`.

[URL slugs explained](http://patterns.dataincubator.org/book/url-slug.html)

### Repositories

- `owner` field for any organization repository is a slugified name for that organization
- access policy:
  - on course creation - default repository access across the whole organization is set to `none`, which means that only the organization owners can see any private repository on that organization
  - when students enroll, they receive read/pull access to `assignments` repository and write/push access to a personal student repository as GitHub invitations to their registered GitHub email

## Docker

QuickFeed will build code submitted by students, and run tests provided by teachers inside docker containers.
An often encountered problem is Docker being unable to resolve DNS due to disabled public DNS.
If you get a build error like that:

```log
Docker execution failed{error 25 0  Error response from daemon: Get https://registry-1.docker.io/v2/: dial tcp: lookup registry-1.docker.io on [::1]:53: read udp [::1]:50111->[::1]:53: read: connection refused}
```

then it must be a DNS problem.

One of the solutions is to uncomment or change `DOCKER_OPTS` line in `/etc/default/docker` file, then restart Docker daemon with `service docker restart`.

[Problem description and possible solutions](https://development.robinwinslow.uk/2016/06/23/fix-docker-networking-dns/)

## npm

`npm install` (or `npm i`) no longer installs all dependencies with versions stated in `package-lock.json`, but will also attempt to load latest versions for all root packages. 
If you just want to install the package dependencies without altering your `package-lock.json`, run `npm ci` instead.

## Repairing database from backups

Given a current database `qf.db` and a backup `bak.db`, and we want to replace records in a table `users` of the `qf.db` with entries from the same table in `bak.db`.
The database you open first will be under the alias `main`.

```sql
sqlite3 qf.db
delete from users;
attach database '/full/path/bak.db' as backup;
insert into main.users select * from backup.users;
detach database backup;
```
