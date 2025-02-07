#image/qf104

start=$SECONDS
printf "*** Initializing Tests for %s ***\n" "$CURRENT"

# Move to folder with assignment handout code for the current assignment to test.
cd "$ASSIGNMENTS/$CURRENT"
# Remove assignment handout tests, if any, to avoid interference, but keep quickfeed tests.
find . \( -name '*_test.go' -and -not -name '*_ag_test.go' \) -exec rm -rf {} \;

# Copy tests into the base assignments folder for initializing test scores
cp -r "$TESTS"/* "$ASSIGNMENTS"/

# $TESTS does not contain go.mod and go.sum: make sure to get the kit/score package
go get -t github.com/quickfeed/quickfeed/kit/score
go mod tidy
# Initialize test scores
SCORE_INIT=1 go test -v ./... 2>&1 | grep TestName

printf "*** Preparing Test Execution for %s ***\n" "$CURRENT"

# Move to folder with submitted code for the current assignment to test.
cd "$SUBMITTED/$CURRENT"
# Remove student written tests to avoid interference
find . -name '*_test.go' -exec rm -rf {} \;

# Copy tests into student assignments folder for running tests
cp -r "$TESTS"/* "$SUBMITTED"/

# $TESTS does not contain go.mod and go.sum: make sure to get the kit/score package
go get -t github.com/quickfeed/quickfeed/kit/score
go mod tidy

printf "\n*** Finished Test Setup in %s seconds ***\n" "$(( SECONDS - start ))"
start=$SECONDS
printf "\n*** Running Tests ***\n\n"
go test -v -timeout 30s ./... 2>&1
printf "\n*** Finished Running Tests in %s seconds ***\n" "$(( SECONDS - start ))"
