# QuickFeed: Instant Feedback on Programming Assignments

[![Go Test](https://github.com/autograde/quickfeed/workflows/Go%20Test/badge.svg)](https://github.com/autograde/quickfeed/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/autograde/quickfeed)](https://goreportcard.com/report/github.com/autograde/quickfeed)
[![Coverall Status](https://coveralls.io/repos/github/autograde/quickfeed/badge.svg?branch=master)](https://coveralls.io/github/autograde/quickfeed?branch=master)
[![Codecov](https://codecov.io/gh/autograde/quickfeed/branch/master/graph/badge.svg)](https://codecov.io/gh/autograde/quickfeed)
[![golangci-lint](https://github.com/autograde/quickfeed/workflows/golangci-lint/badge.svg)](https://github.com/autograde/quickfeed/actions)

## Documentation

- Teachers that wants to use QuickFeed may wish to review the [User Manual](doc/teacher.md).
- Teachers may also want to copy the [sign up instructions](doc/templates/signup.md) and [lab submission instructions](doc/templates/lab-submission.md), and make the necessary adjustments for your course.
- For detailed instructions on installation of QuickFeed for a production environment, please see our [Server Installation and Configuration](doc/install.md).

## Development

For detailed instructions on configuring QuickFeed for development, please see our [Developer Guide](doc/dev.md).

### We Accept Pull Requests

We are happy to accept pull requests from anyone that want to help out in our effort to implement our QuickFeed platform.
To avoid wasted work and duplication of efforts, feel free to first open an issue to discuss the feature or bug you want to fix.

### Style Guidelines

We chose to implement QuickFeed in Go and Typescript because these languages offer simplicity and type safety.
We therefore require that certain style guidelines are followed when creating pull requests.

For Go, we expect code to follow these style guidelines and list of common mistakes:

- We use the `golangci-lint` linter in VSCode.

- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments).
  In particular, note the section on how to
  [Handle Errors](https://github.com/golang/go/wiki/CodeReviewComments#handle-errors),
  [Mixed Caps](https://github.com/golang/go/wiki/CodeReviewComments#mixed-caps),
  [Variable Names](https://github.com/golang/go/wiki/CodeReviewComments#variable-names).

For Typescript, we think these [style guidelines](https://github.com/basarat/typescript-book/blob/master/docs/styleguide/styleguide.md) look reasonable.
Moreover, the `formatOnSave` and `tslint.run` options in VSCode should help maintain reasonable style.

Note that we currently violate the [interface naming](https://github.com/basarat/typescript-book/blob/master/docs/styleguide/styleguide.md#interface)
guideline by using the `I` prefix on some interfaces, and several of the other guidelines.
We have started to rename these interfaces, and will eventually rename all such interfaces.
