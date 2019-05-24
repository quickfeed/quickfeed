# Notes to developers

To run some of our unit and integration tests, as well as the `scm` tool,
you will need to set up a personal access token.
This is done by on GitHub's web page:

1. Navigate to Settings (in the personal menu accessible from your avatar picture)
2. Select _Developer settings_ from the menu on the left.
3. Select _Personal access tokens_ and on the next page,
4. Select _Generate new token_. Name the token, e.g. `Autograder Test Token`.
5. Select _Scopes_ as needed; currently I have enabled `admin:org, admin:org_hook, admin:repo_hook, delete_repo, repo, user`, but you may be able to get away with fewer access scopes. It depends on your needs.
6. Copy the generated token string to the `GITHUB_ACCESS_TOKEN` environment variable. You may wish to add this token to your local `ag-setup.sh` script file.

```sh
  export GITHUB_ACCESS_TOKEN=<your token>
```
