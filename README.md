# QuickFeed: Instant Feedback on Programming Assignments

[![Go Test](https://github.com/quickfeed/quickfeed/workflows/Go%20Test/badge.svg)](https://github.com/quickfeed/quickfeed/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/quickfeed/quickfeed)](https://goreportcard.com/report/github.com/quickfeed/quickfeed)
[![Codecov](https://codecov.io/gh/quickfeed/quickfeed/branch/master/graph/badge.svg)](https://codecov.io/gh/quickfeed/quickfeed)
[![golangci-lint](https://github.com/quickfeed/quickfeed/workflows/golangci-lint/badge.svg)](https://github.com/quickfeed/quickfeed/actions)

## Documentation

- Teachers that wants to use QuickFeed may wish to review the [User Manual](doc/teacher.md).
- Teachers may also want to copy the [sign up instructions](doc/templates/signup.md) and [lab submission instructions](doc/templates/lab-submission.md), and make the necessary adjustments for your course.
- [Installation instructions for QuickFeed](doc/deploy.md).

### Install Development Tools

On Unix systems with homebrew you should be able to install development tools using:

```shell
% make brew
% make devtools
# Make sure the bin folder with our tools is in your PATH.
% export PATH=$PATH:$PWD/bin
```

For non-brew users, please inspect the `Makefile` to determine instructions for your system.

## Contributing

The following instructions assume you have installed the [GitHub CLI](https://github.com/cli/cli).
See here for [installation instructions](https://github.com/cli/cli#installation) for your platform.

Further, we require that code is formatted according to the rules and extensions that have been configured for VSCode.
When opening VSCode, please install the recommended extensions for QuickFeed; [see also style guidelines below](#style-guidelines).
Specifically, you will need to install the `clang-format` tool to edit `.proto` files, and the `golangci-lint` tool to edit `.go` files.

### Create Issue First

Before you implement some feature or bug fix, you should open an issue first.
This issue should then be linked in the corresponding pull request.

### Create Pull Request

Before starting a new pull request, either clone the repo:

```shell
% gh repo clone quickfeed/quickfeed
% cd quickfeed
```

Or if you have already cloned, make sure to start from an up-to-date master branch:

```shell
# Make sure to start from master branch
% git checkout master
# Make sure your master branch is up-to-date
% git pull
```

To create a pull request on the main repository follow these steps.

```shell
# Create and switch to your new feature branch
% git switch -C <feature-branch>
# Edit and stage files
% git add <files>
% git commit
# When done and ready to share
% gh pr create --title "Short description of the feature or fix"
# Alternatively: Use --draft if you want to share your code, but want to continue developing
% gh pr create --draft --title "Short description of the feature or fix"
```

To continue development on a pull request (same branch as before):

```shell
# Only necessary if you previously switched away from the feature-branch
% git switch <feature-branch>
# Edit and stage files
% git add <files>
% git commit
% git push
```

To fetch an existing pull request to your local machine.

```shell
% gh pr checkout <PR#>
```

For additional details on the `gh pr` and `gh pr create` commands:

```shell
% gh help pr
% gh help pr create
```

### GitHub Issues and Pull Requests

When creating a pull request, it is always nice to connect it to a GitHub issue describing the feature or problem you are fixing.
If there is an issue that is fixed by your pull request please remember to add one of the following lines at the end of the pull request description.

```text
Closes <Issue#>.
Fixes <Issue#>.
Resolves <Issue#>.
```

For detailed instructions on configuring QuickFeed for development, please see our [Developer Guide](doc/dev.md).

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
