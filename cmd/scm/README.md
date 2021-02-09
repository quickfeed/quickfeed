# Source Control Management Tool

The SCM tool is designed for working with GitHub organizations from the command line.
For example, you can list or remove all repositories and teams from your test organization, without having to interact manually with GitHub's web user interface.

You must be an owner of the GitHub organization to be able to access its repositories and teams with SCM tool.

## Installation

SCM tool must be compiled before it can be used.
To compile the tool from the projects root folder:

```sh
make scm
```

Or from the `cmd/scm` folder:

```sh
go install
```

This will compile and install the tool in your `$GOPATH/bin` or `$GOBIN` folder; this path should also be added to your `$PATH` variable.

## GitHub Access Token

To use SCM tool, you need to create a personal GitHub access token.
This is done on GitHub's web page:

1. Click your profile picture and select Settings.
2. Select _Developer settings_ from the menu on the left.
3. Select _Personal access tokens_ from the menu on the left.
4. Select _Generate new token_.
5. Name the token, e.g. `QuickFeed SCM Token`.
6. Select _Scopes_ as needed.
   Currently I have enabled `admin:org, admin:org_hook, admin:repo_hook, delete_repo, repo, user`, but you may be able to get away with fewer access scopes.
   It depends on your needs.
7. Copy the generated token string to the `GITHUB_ACCESS_TOKEN` environment variable.
   You may wish to add this token to your local `quickfeed-env.sh` script.

   ```sh
   export GITHUB_ACCESS_TOKEN="your token"
   ```

## Example Usage

Assuming you are an owner of the `qf101` GitHub organization you can perform the several commands.
To print information about all repositories under the `qf101` organization, you can run:

```sh
scm --provider github get repo -all -namespace qf101
```

To delete all teams under the `qf101` organization, you can run:

```sh
scm delete team -all -namespace qf101
```

For additional examples and instructions please see the comments in `cmd/scm/main.go`.
