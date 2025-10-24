# Overview of QuickFeed's code base

## QuickFeed's code base

QuickFeed has been in development since 2014 and is tightly integrated with GitHub.
QuickFeed eases the GitHub experience for teachers and students by automating the process of configuring repositories for their courses.
Teachers create lab assignments and related tests, which gives students rapid responses on their submissions.

### Abbreviations

- **ci** - Continuous Integration, closely related is CD - Continuous Deployment/Delivery.
  - We recommend learning about DevOps.
  See Microsoft's article: [What is DevOps](https://learn.microsoft.com/en-us/devops/what-is-devops).
- **cmd** - command
- **qcm** - quickfeed course manager (planning to deprecate this)
- **vercheck** - version check (planning to deprecate this)
- **doc** - documentation
- **dev** - developer
- **src** - source
- **qf** - quickfeed
- **rpc** - Remote procedure call
- **scm** - source code management
- **bh** - baseHook [related-file](web/bh.go)
- **os** - operating system
- **db** - database
- **CRUD** - Create, Read, Update and Delete

## About each secondary folder

The following sections give a brief explanation of each subfolder in the root directory of the repository.

- [Overview of QuickFeed's code base](#overview-of-quickfeeds-code-base)
  - [QuickFeed's code base](#quickfeeds-code-base)
    - [Abbreviations](#abbreviations)
  - [About each secondary folder](#about-each-secondary-folder)
  - [assignments](#assignments)
    - [Methods used in web folder](#methods-used-in-web-folder)
  - [ci - Continuous Integration](#ci---continuous-integration)
  - [cmd - Commands](#cmd---commands)
  - [database](#database)
  - [internal](#internal)
  - [kit](#kit)
  - [metrics](#metrics)
  - [patch](#patch)
  - [public](#public)
  - [qf - QuickFeed](#qf---quickfeed)
  - [scm - Source code management](#scm---source-code-management)
  - [testdata](#testdata)
  - [web](#web)

## assignments

The `assignments` package contains functionality triggered by push events from the `tests` repository, such as:

- cloning and scanning the `tests` repository for new assignments
  - `assignments.json`
  - `criteria.json`
  - `tests.json`
  - `Dockerfile`
  - `tasks-*.md` files
- parsing assignment information from `assignment.json` files
- triggering rebuild of the course's docker image
- creating and updating assignments in the database

The package also contains methods executed on pull requests on student repositories.
It automates processes like assigning reviewers for pull requests and synchronizing tasks by creating or updating issues.
See the related documentation: [pr-feedback](design-docs/pr-feedback.md), [github-enhancement](design-docs/github-enhancement.md).
Currently, no courses are using the pull request and review assignment functionality.

### Methods used in web folder

- [AssignReviewers()](assignments/pull_requests.go#L30)
  - Called from [handlePullRequestPush()](web/hooks/pull_request.go#L47)
- [UpdateFromTestsRepo()](assignments/assignments.go#L32)
  - Called from:
    - [handlePush()](web/hooks/push.go#L55)
    - [UpdateAssignments()](web/quickfeed_service.go#L491)

The rest of the methods in the `assignments` folder are dependencies of either or both of the previously mentioned methods.

## ci - Continuous Integration

The `ci` folder clones repositories, builds Docker images, and creates containers to run tests on assignments.

See [Notes on Using Docker](docker.md) for more information related to Docker in QuickFeed's context.

## cmd - Commands

The `cmd` folder contains different executable Go and Python programs:

- **anonymize**: creates a new database which filters out sensitive information
- **approvelist**: query the QuickFeed's database to retrieve an overview over approved assignments
- **qcm**: clone repository and run tests locally, `go run qcm clone --help` gives a list of filter values
- **vercheck**: checks the version of protobuf

## database

The `database` folder contains methods that perform CRUD operations on a [GORM](https://gorm.io/index.html) database.
See [Methods Overview](/database/database.go).

## internal

The `internal` packages interact with the OS to get sensitive information from the environment file `.env` and include methods that add certificates.

## kit

`kit` contains helper functions to be used in course specific test cases so that QuickFeed can compute a score for the code submitted by students.
It also contains code to help score multiple choice exercises and command line execution that returns a given expected output.

## metrics

Metrics are useful when developing an application and QuickFeed utilizes the open source system called Prometheus.
"Metrics are numerical measurements in layperson terms." See the [Prometheus documentation](https://prometheus.io/docs/introduction/overview/).

The metrics server is initialized in both development and production servers with: [metricsServer()](/web/server.go#L109).

See [Metrics Collection](metrics.md) for more information and guidance.

## patch

This folder contains files related to protobuf patching and generation.
If you are unsure about its purpose, consult the protobuf build configuration files in the repository and the Makefile targets.

## public

The `public` directory contains all the frontend code.
The frontend is built with [React](https://react.dev/), and is a single-page application.

## qf - QuickFeed

The `qf` folder is mainly used to define types, which standardizes types throughout the application.
They are defined using protobuf; see the [documentation](https://protobuf.dev/getting-started/gotutorial/).

## scm - Source code management

The `scm` folder contains methods that perform CRUD operations on GitHub, from creating repositories and managing roles, to managing issues, groups, and enrollments.
See the [SCM interface](/scm/scm.go#L13).

## testdata

The `testdata` folder is used in tests, for example [TestDockerBindDir](/ci/docker_test.go#L116).
This is a convention in Go, and the folder is ignored by the `go test` command.

## web

The `web` package serves as an interface between the frontend and backend and contains all the API endpoints that integrate business logic from the following folders: [qf](../qf), [scm](../scm), [database](../database), [ci](../ci), and [assignments](../assignments).
The `web` package also acts on events triggered on GitHub through [webhooks](https://docs.github.com/en/webhooks).
