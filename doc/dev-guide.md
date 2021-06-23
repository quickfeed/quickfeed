# QuickFeed Developer Guide

This developer guide assumes that you have [installed and configured QuickFeed](./deploy.md) and its dependent components.

## Recommended VSCode Plugins

- Go
- vscode-proto3
- Code Spell Checker
- TSLint
- markdownlint
- Better Comments
- GitLens
- Git History Diff
- SQLite

## Shipping Code

In the following instructions we assume you have installed the [GitHub CLI](https://github.com/cli/cli).
On systems with homebrew:

```shell
% brew install gh
% gh help
```

To create a pull request on the main repository follow these steps.

```shell
% gh repo clone autograde/quickfeed
% cd quickfeed
# Create and switch to your new feature branch
% git switch -C <feature-branch>
# Add/edit files
% git add <files>
% git commit
# When done and ready to share
% gh pr create --title "Short description of the feature or fix"
# Use --draft if you want to share your code, but want to continue developing
% gh pr create --draft --title "Short description of the feature or fix"
```

To continue development on a pull request (same branch as before):

```shell
# Only necessary if you previously switched away from the feature-branch
% git switch <feature-branch>
# Add/edit files
% git add <files>
% git commit
% git push
```

To fetch an existing pull request to your local machine.

```shell
% gh pr checkout <PR#>
```

For details about the `gh pr` and `gh pr create` commands:

```shell
% gh help pr
% gh help pr create
```

## GitHub Issues and Pull Requests

When creating a pull request, it is always nice to connect it to a GitHub issue describing the feature or problem you are fixing.
If there is an issue that is fixed by your pull request then you must remember to add one of the following lines at the end of the pull request description.

```text
Closes <Issue#>.
Fixes <Issue#>.
Resolves <Issue#>.
```
