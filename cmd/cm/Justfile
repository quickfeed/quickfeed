set dotenv-load

# Variables
dest  := "../$YEAR"
username := `gh auth status 2>/dev/null | awk '/Logged in to/ {print $(NF-1)}'`
rsync := "rsync --prune-empty-dirs -av --itemize-changes --delete --exclude=.git"
bsync := "rsync --prune-empty-dirs -av --itemize-changes"

[private]
@default:
    just --list

# Create .env file for the course. Do only once and commit the file to git and edit it as needed.
@env: install-cm
    cm init-env --year 2025 --course dat520 --name "Distributed Systems" --discord-join-url "https://discord.gg/abc123" --bot-user "dat520-helpbot"

# Initialize the course repositories based on the .env file.
@init: install-cm
    cm init-repos

# Update the README.md files for all labs based on readme_tmpl.md files.
@readme: install-cm
    cm gen-readme

# Clone the username-labs repository to the $dest directory, where username is
# the GitHub username of the user running the command.
# This is a one-time operation to clone the repository to the correct location.
# This requires that the username-labs repository exists on GitHub; it should be
# created by the user on QuickFeed first and left empty.
@clone: install-cm
    cm clone-repo -pull

# Install the cm command to simplify course management.
@install-cm:
    go install dat520/internal/cmd/cm

# Install the required tools for the course repository.
@tools:
    brew install protobuf gh just tree golangci-lint
    go install tool

[private]
@msg target:
    tree {{dest}}/{{target}}
    echo ""
    echo "Manual steps that may be performed next:"
    echo "  cd {{dest}}/{{target}}"
    echo "  git status"
    echo "  git diff"
    echo "  git add ."
    echo "  git commit"
    echo "  git push -u origin main"

# Sync changes to the info repository.
@info: install-cm && (msg "info")
    echo "Syncing info repo to {{dest}}/info"
    go mod tidy
    {{rsync}} --filter='dir-merge /.rsync-filter-info' info/ {{dest}}/info
    cm update-doc-tags -repo info
    (cd {{dest}}/info && go mod tidy && git status)

# Sync changes to the assignments repository.
@assignments +lab: install-cm && (msg "assignments")
    echo "Syncing {{lab}} assignments to {{dest}}/assignments/"
    go mod tidy
    # Sync shared files from internal to assignments (using bsync to avoid deleting lab folders)
    {{bsync}} --filter='dir-merge /.rsync-filter-assignments' . {{dest}}/assignments
    # Sync lab-specific files (using rsync which will delete files no longer in the source folder)
    {{rsync}} --filter='dir-merge /.rsync-filter-assignments' {{lab}} {{dest}}/assignments
    (cp go.mod {{dest}}/assignments && cd {{dest}}/assignments && go mod tidy)
    cm update-doc-tags -repo assignments
    cm remove-solution-tags -repo assignments
    (cd {{dest}}/assignments && go mod tidy && git status)

# Sync changes to the tests repository.
@tests +lab: install-cm && (msg "tests")
    echo "Syncing {{lab}} tests to {{dest}}/tests/"
    go mod tidy
    cm add-main-tests
    cm add-lint-checkers -labs "{{lab}}"
    # Sync shared files from internal to tests (using bsync to avoid deleting lab folders)
    {{bsync}} --filter='dir-merge /.rsync-filter-tests' . {{dest}}/tests
    # Sync lab-specific files (using rsync which will delete files no longer in the source folder)
    {{rsync}} --filter='dir-merge /.rsync-filter-tests' {{lab}} {{dest}}/tests
    (cd {{dest}}/tests && git status)

# Sync changes from the per-year info repository back to the info directory.
@back-info:
    # Warning: This will overwrite some content in the info directory (use with caution)
    # TODO: avoid overwriting files with placeholders like COURSE_ORG, COURSE_NAME etc.
    echo "Syncing {{dest}}/info back to info"
    {{bsync}} --filter='dir-merge /.rsync-filter-info' {{dest}}/info/ info
    git status

# Sync changes from the per-year assignments repository back to the lab folders.
@back-assignments +lab:
    # TODO: should probably only back sync .md files
    echo "Syncing {{dest}}/assignments/{{lab}} back to {{lab}}"
    {{bsync}} --filter='dir-merge /.rsync-filter-assignments' {{dest}}/assignments/{{lab}}/ {{lab}}
    git status

# Sync changes from the per-year tests repository back to the tests directory.
@back-tests +lab:
    # TODO: what do we need to sync back?
    echo "Syncing {{dest}}/tests/{{lab}} back to {{lab}}"
    {{bsync}} --filter='dir-merge /.rsync-filter-tests' {{dest}}/tests/{{lab}}/ {{lab}}
    git status

# Check the last updated date in all readme_tmpl.md files.
@last-updated:
    find . -name readme_tmpl.md | xargs grep "Last.updated"
