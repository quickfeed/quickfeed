# Overview of QuickFeed's code base

## QuickFeed's code base

QuickFeed has been in development since 2014 and is tightly integrated with GitHub.
QuickFeed eases the GitHub experience for teachers and thereby also students by automating the process of configuring repositories for their courses.
Teachers creates lab assignments and related tests, which gives students rapid response on their submissions.

### Abbreviations

- **ci** - Continuous Integration, closely related is CD - Continuous Deployment/Delivery.
  - Would recommend learning about DevOps, view Microsoft's article [What is DevOps](<https://learn.microsoft.com/en-us/devops/what-is-devops>)
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

The following sections gives a brief explanation of each subfolder in the root directory of the repository.

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
Essentially automates processes like assigning reviewers for pull requests and synchronizing tasks by creating or updating issues.
See the related documentation: [pr-feedback](design-docs/pr-feedback.md), [github-enhancement](design-docs/github-enhancement.md).
Currently, no courses are using the pull request and review assignment functionality.

### Methods used in web folder

- [AssignReviewers()](assignments/pull_requests.go#30)
  - Called from [handlePullRequestPush()](web/hooks/pull_request.go#L47)
- [UpdateFromTestsRepo()](assignments/assignments.go#32)
  - Called from:
    - [handlePush()](web/hooks/push.go#L55)
    - [UpdateAssignments()](web/quickfeed_service.go#491)

Rest of the methods in the assignments folder are a dependency of either or both previously mentioned methods

## ci - Continuous Integration

The `ci` folder clone repositories, builds docker image and creates a container to run tests on assignments.

View [Notes on Using Docker](docker.md) for more information related to docker in QuickFeed's context

## cmd - Commands

cmd folder contains different executable go and python programs:

- **anonymize**: creates a new database which filters out sensitive information
- **approvelist**: query the QuickFeed's database to retrieve an overview over approved assignments
- **qcm**: clone repository and run tests locally, `go run qcm clone --help` gives a list of filter values
- **vercheck**: checks the version of protobuf

## database

The database folder contains methods which does CRUD operations on a [gorm](https://gorm.io/index.html) database, [Methods Overview](/database/database.go)

## internal

Internal interacts with the [os](#L26) to get sensitive information from the environment file: ".env", and has methods which add certificates.

## kit

Kit contains helper functions to be used in course specific test cases so that QuickFeed can compute a score for the code submitted by students.
It also contains code to help score multiple choice exercises and command line execution that returns a given expected output.

## metrics

Metrics are useful when developing an application and QuickFeed utilize the open source system called prometheus, and "Metrics are numerical measurements in layperson terms" - [documentation](https://prometheus.io/docs/introduction/overview/).

The metrics server is initialized in both development and production server, with: [metricsServer()](/web/server.go#L109)

View [Metrics Collection](metrics.md) for more information and guidance

## patch

It has something to do with protobuf.. but im uncertain of what this does..

## public

The public directory is where all the frontend code is stored, the frontend is built with [react](https://react.dev/), and is a single page application - [Beginner's guide](https://dev.to/hiteshtech/a-beginners-guide-to-create-spa-with-react-js-491c).

## qf - QuickFeed

The qf folder is mainly used to define types, which standardizes the types throughout the application.
They are defined using protobuf, [documentation](https://protobuf.dev/getting-started/gotutorial/)

## scm - Source code management

scm folder contains methods which preforms [CRUD](#L31) operations on GitHub, from creating repositories, demoting teacher to student, and managing issues, groups, enrollments and etc, please view [SCM interface](/scm/scm.go#L13).

## testdata

The `testdata` folder is used in this test [TestDockerBindDir](/ci/docker_test.go#L116).
This is a convention in Go, and the folder is ignored by `go test` command.

## web

Web serves as an interface between the frontend and backend, and contains all the API endpoints that integrate business logic from the following folders: [qf](qf), [scm](scm), [database](database), [ci](ci) and [assignments](assignments). Web also acts on events triggered on GitHub through [web hooks](https://docs.github.com/en/webhooks).
