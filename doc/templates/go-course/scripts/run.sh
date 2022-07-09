#image/qf101

start=$SECONDS
printf "*** Preparing for Test Execution ***\n"

ASSIGNMENTS=/quickfeed/assignments
TESTS=/quickfeed/tests
ASSIGNDIR=$ASSIGNMENTS/{{ .AssignmentName }}/

# Move to folder for assignment to test.
cd "$ASSIGNDIR"

# Remove student written tests to avoid interference
find . -name '*_test.go' -exec rm -rf {} \;

# Copy tests into student assignments folder for running tests
cp -r $TESTS/* $ASSIGNMENTS/

printf "\n*** Finished Test Setup in %s seconds ***\n" "$(( SECONDS - start ))"
start=$SECONDS
printf "\n*** Running Tests ***\n\n"
QUICKFEED_SESSION_SECRET={{ .RandomSecret }} go test -v -timeout 30s ./... 2>&1
printf "\n*** Finished Running Tests in %s seconds ***\n" "$(( SECONDS - start ))"
