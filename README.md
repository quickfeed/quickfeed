# QuickFeed: Instant Feedback on Programming Assignments

[![Go Test](https://github.com/autograde/quickfeed/workflows/Go%20Test/badge.svg)](https://github.com/autograde/quickfeed/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/autograde/quickfeed)](https://goreportcard.com/report/github.com/autograde/quickfeed)
[![Coverall Status](https://coveralls.io/repos/github/autograde/quickfeed/badge.svg?branch=master)](https://coveralls.io/github/autograde/quickfeed?branch=master)
[![Codecov](https://codecov.io/gh/autograde/quickfeed/branch/master/graph/badge.svg)](https://codecov.io/gh/autograde/quickfeed)

## Roles and Concepts

The system has three **user** roles.

- **Administrators** can create new courses and promote other users to become administrator. It is common that all teachers that are responsible for one or more courses be an administrator. The administrator role is system-wide.

- **Teachers** are associated with one or more courses. A teacher can view results for all students in his course(s). A course can have many teachers. A teacher is anyone associated with the course that are not students, such as professors and teaching assistants.

- **Students** are associated with one or more courses. A student can view his own results and progress on individual assignments.

The following concepts are important to understand.

- **Assignments** are organized into folders in a git repository.
  - **Individual assignments** are solved by one student. There is one repository for individual assignments.
  - **Group assignments** are solved by a group of students. There is one repository for group assignments.
- **Submissions** are made by a student submitting his code to a supported git service provider (e.g. github or gitlab).

## Download and install

   ```sh
   go get -u github.com/autograde/quickfeed
   ```

## Running the server

   ```sh
   # Server listening on port 8080 serving static files from /public at https://example.com/.
   quickfeed -service.url example.com -http.addr :8080 -http.public /public
   ```

*As a bootstrap mechanism, the first user to sign in is automatically made administrator for the system.*

## Install for React web development

   ```sh
   cd public
   npm install
   webpack
   ```

## Development

### We accept pull requests

We are happy to accept pull requests from anyone that want to help out in our
effort to implement a strong autograder platform. To create a PR, simply fork
our repo, or create a new branch, and then follow the usual guidelines for
creating a PR.

### Style guidelines

We chose to implement Autograder in Go and Typescript because these langauges
offer simplicity and type safety. We therefore require that certain style guidelines
are followed when creating pull requests.

For Go, we expect code to follow these style guidelines and list of common mistakes:

- The `gofmt` should always be used. This is usually handled automatically in VSCode
  when the `formatOnSave` option is set to true; see below.

- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments).
  In particular, note the section on how to
  [Handle Errors](https://github.com/golang/go/wiki/CodeReviewComments#handle-errors),
  [Mixed Caps](https://github.com/golang/go/wiki/CodeReviewComments#mixed-caps),
  [Variable Names](https://github.com/golang/go/wiki/CodeReviewComments#variable-names).

For Typescript, we think these [style guidelines](https://github.com/basarat/typescript-book/blob/master/docs/styleguide/styleguide.md)
look reasonable. Moreover, the `formatOnSave` and `tslint.run` options (see below)
should help maintain reasonable style.

Note that we currently violate the [interface naming](https://github.com/basarat/typescript-book/blob/master/docs/styleguide/styleguide.md#interface)
guideline by using the `I` prefix on interfaces, and several of the other guidelines.
We should refactor these, to the extent possible.

### Working with webpack

To ensure that webpack bundle files are updated when you pull in changes or
rebase from the repository you can add the following script to the files
`post-merge` (invoked on git pull) and `post-rewrite` (invoked on git rebase)
in the `.git/hooks/` folder.

   ```sh
   #!/bin/sh
   cd $GOPATH/src/github.com/autograde/quickfeed/public
   webpack
   ```

If you don't want to run `webpack` to create the bundle files on git pull/rebase,
you will need to manually run `webpack` in the `public` folder.

### Visual Studio Code Configuration

The development team has mainly used VSCode and we recommend using the `tslint`
plugin and the Go plugin together with the following configuration settings.

```json
{
    "go.inferGopath": true,
    "go.lintTool": "megacheck",
    "editor.formatOnSave": true,
    "tslint.run": "onSave",
}
```
