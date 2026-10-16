# Preparing a new Release of QuickFeed's kit Module

## Testing the kit module before a release

Before releasing the kit module, you may wish to test it locally on a course that uses the various APIs; to do this, use go mod's `replace` directive:

```shell
% go mod edit -replace=github.com/quickfeed/quickfeed/kit=../../quickfeed/kit
```

## Install tools needed for release

To cut a release you will need additional tools:

```shell
% go install golang.org/x/exp/cmd/gorelease@latest
% brew install gh
```

## Steps necessary to cut a release

Below are the steps needed to prepare a new release of QuickFeed's kit module.

1. Make sure `master` is synchronized with `origin/master`, then run `gorelease` from the `kit` directory to get the suggested new version number, e.g.:

   ```text
   ... (list of compatibility changes) ...
   Inferred base version: v0.2.0
   Suggested version: v0.3.0 (with tag kit/v0.3.0)
   ```

2. If there are uncommitted changes due to upgrades and recompilation, commit and push them.
   Skip this step if the working tree is clean:

   ```shell
   % git add -u
   % git commit -m "QuickFeed's kit module release v0.3.0"
   % git push
   ```

3. Publish the release with release notes:

   ```shell
   # Prepare release notes in release-notes.md
   # (base them on the gorelease output and the commits since the previous release)
   % gh release create kit/v0.3.0 --prerelease -F release-notes.md --title "Main changes in release"
   ```

   Without release notes file (select `Write my own` when asked about release notes; a non-interactive run needs `--notes` or `-F`):

   ```shell
   % gh release create kit/v0.3.0 --prerelease --title "Revised MultipleChoice API; rerelease"
   ```

   Without the `gh` tool:

   ```shell
   % git tag kit/v0.3.0
   % git push origin kit/v0.3.0
   ```

   Now other projects can depend on `v0.3.0` of `github.com/quickfeed/quickfeed/kit`.

4. To check that the new version is available (after a bit of time):

    ```shell
    % go list -m github.com/quickfeed/quickfeed/kit@v0.3.0
    ```

5. From your course that depend on new features of the kit module:

   ```shell
   # if you have used a replace directive
   % go mod edit -dropreplace=github.com/quickfeed/quickfeed/kit
   % go get -u github.com/quickfeed/quickfeed/kit
   % go mod tidy
   % git add go.mod go.sum
   % git commit -m "Upgraded to latest version of kit module"
   % git push
   ```
