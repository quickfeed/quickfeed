# Overview of Quickfeed's code base

## Quickfeed's code base

Quickfeed has been in development for over 8 years and has formed into a quite complex system, which is highly integrated with github. Quickfeed eases the github experience for teachers and thereby also students by automating the process of configuring repositories for their course. Teachers can easily add labs and related tests, which gives students rapid response on their submissions.

### Abbreviations

- **ci** - Continuos Integration, closely related is CD - Continuos Deployment/Delivery.
  - Would recommend learning about DevOps, view Microsoft's article [What is DevOps](<https://learn.microsoft.com/en-us/devops/what-is-devops>)
- **cmd** - command
- **qcm** - quickfeed course manager
- **vercheck** - version check
- **doc** - documentation
- **dev** - developer
- **src** - source
- **qf** - quickfeed
- **rpc** - Remote procedure call
- **scm** - source code management
- **bh** - baseHook [related-file](web/bh.go)
- **os** - operation system
- **db** - database
- **CRUD** - Create, Read, Update and Delete

## About each secondary folder

Following sections will have a brief explanation of each subfolder from the root folder; quickfeed.

- [assignments](#assignments)
- [Methods used in web folder](#methods-used-in-web-folder)
- [ci](#ci---continuos-integration)
- [cmd](#cmd---commands)
- [database](#database)
- [internal](#internal)
- [kit](#kit)
- [metrics](#metrics)
- [patch](#patch)
- [public](#public)
- [qf](#qf---quickfeed)
- [scm](#scm---source-code-management)
- [testdata](#testdata)
- [web](#web)

## assignments

Contains methods executed on push and pull requests events for test repositories. Essentially automates processes like assigning reviewers for pull requests and synchronizing tasks by creating or updating issues.

Related documentation: [pr-feedback](design-docs/pr-feedback.md), [github-enhancement](design-docs/github-enhancement.md)

This code is currently not in use.

### Methods used in web folder

- [AssignReviewers()](assignments/pull_requests.go#30)
  - Called from [handlePullRequestPush()](web/hooks/pull_request.go#L47)
- [UpdateFromTestsRepo()](assignments/assignments.go#32)
  - Called from:
    - [handlePush()](web/hooks/push.go#L55)
    - [UpdateAssignments()](web/quickfeed_service.go#491)

Rest of the methods in the assignments folder are a dependency of either or both previously mentioned methods

## ci - Continuos Integration

The ci folder clone repositories, builds docker image and creates a container to run tests on assignments.

View [Notes on Using Docker](docker.md) for more information related to docker in quickfeed's context

## cmd - Commands

cmd folder contains different executable go and python programs:

- **anonymize**: creates a new database which filters out sensitive information
- **approvelist**: query the quickfeed's database to retrieve an overview over approved assignments
- **qcm**: clone repository and run tests locally, `go run qcm clone --help` gives a list of filter values
- **vercheck**: checks the version of protobuf

## database

The database folder contains methods which does CRUD operations on a [gorm](https://gorm.io/index.html) database, [Methods Overview](/database/database.go)

## internal

Internal interacts with the [os](#L26) to get sensitive information from the environment file: ".env", and has methods which add certificates.

## kit

Kit contains helper functions to be used in course specific test cases so that quickfeed can compute a score for the code submitted by students.
It also contains code to help score multiple choice exercises and command line execution that returns a given expected output.

## metrics

Metrics are useful when developing an application and quickfeed utilize the open source system called prometheus, and "Metrics are numerical measurements in layperson terms" - [documentation](https://prometheus.io/docs/introduction/overview/).

The metrics server is initialized in both development and production server, with: [metricsServer()](/web/server.go#L109)

View [Metrics Collection](metrics.md) for more information and guidance

## patch

It has something to do with protobuf.. but im uncertain of what this does..

## public

The public directory is where all the frontend code is stored, the frontend is built with [react](https://react.dev/), and is a single page application - [Beginner's guide](https://dev.to/hiteshtech/a-beginners-guide-to-create-spa-with-react-js-491c).

## qf - Quickfeed

The qf folder is mainly used to define types, which standardizes the types throughout the application. They are defined using protobuf, [documentation](https://protobuf.dev/getting-started/gotutorial/)

## scm - Source code management

scm folder contains methods which preforms [CRUD](#L31) operations on github, from creating repositories, demoting teacher to student, and managing issues, groups, enrollments and etc, please view [SCM interface](/scm/scm.go#L13).

## testdata

testData is used in this test [TestDockerBindDir](/ci/docker_test.go#L116)

## web

Web serves as an interface between the frontend and backend, and contains all the API endpoints that integrate business logic from the following folders: [qf](qf), [scm](scm), [database](database), [ci](ci) and [assignments](assignments). Web also acts on events triggered on github through [web hooks](https://docs.github.com/en/webhooks).
