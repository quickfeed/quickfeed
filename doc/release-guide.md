# Preparing a new Release of QuickFeed's kit Module

Below are the steps to prepare a new release of QuickFeed's kit Module.

To cut a release you will need additional tools:

```shell
% go install golang.org/x/exp/cmd/gorelease@latest
% brew install gh
```

1. Run `gorelease` to suggested new version number, e.g.:

   ```text
   ... (list of compatability changes) ...
   Inferred base version: v0.2.0
   Suggested version: v0.3.0 (with tag kit/v0.3.0)
   ```

2. Add and commit changes due to upgrades and recompilation:

   ```shell
   % git add
   % git commit -m "QuickFeed's kit module release v0.3.0"
   # Synchronize master branch
   % git push
   ```

3. Publish the release with release notes:

   ```shell
   # Prepare release notes in release-notes.md
   % gh release create kit/v0.3.0 --prerelease -F release-notes.md --title "Main changes in release"
   ```

   Without release notes file (select `Write my own` when asked about release notes):

   ```shell
   % gh release create kit/v0.3.0 --prerelease --title "Revised MultipleChoice API; rerelease"
   ```

   Without the `gh` tool:

   ```shell
   % git tag kit/v0.3.0
   % git push origin kit/v0.3.0
   ```

   Now other projects can depend on `v0.3.0` of `github.com/autograde/quickfeed/kit`.

4. To check that the new version is available (after a bit of time):

    ```shell
    % go list -m github.com/autograde/quickfeed/kit@v0.3.0
    ```

5. From your course that depend on new features of the kit module:

   ```shell
   % go get -u github.com/autograde/quickfeed/kit
   % go mod tidy
   % git add go.mod go.sum
   % git commit -m "Upgraded to latest version of kit module"
   % git push
   ```
