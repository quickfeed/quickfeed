#image/qf101

start=$SECONDS
printf "*** Preparing for Test Execution ***\n"

# Move to folder for the current assignment to test.
cd "$ASSIGNMENTS/$CURRENT"

# Remove student written tests to avoid interference
find . -name '*_test.go' -exec rm -rf {} \;

# Copy tests into student assignments folder for running tests
cp -r "$TESTS"/* "$ASSIGNMENTS"/

# Needed to ensure that dependencies from both $TESTS and $ASSIGNMENTS are blessed
go mod tidy

printf "\n*** Finished Test Setup in %s seconds ***\n" "$(( SECONDS - start ))"
start=$SECONDS
printf "\n*** Running Tests ***\n\n"
go test -v -timeout 30s ./... 2>&1
printf "\n*** Finished Running Tests in %s seconds ***\n" "$(( SECONDS - start ))"
