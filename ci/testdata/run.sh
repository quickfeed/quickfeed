#image/quickfeed:go

start=$SECONDS
printf "*** Preparing for Test Execution ***\n"

ASSIGNMENTS=/quickfeed/assignments
TESTDIR=/quickfeed/tests
ASSIGNDIR=$ASSIGNMENTS/{{ .AssignmentName }}/

if [ ! -d "$ASSIGNDIR" ]; then
  printf "Folder %s not found" "$ASSIGNDIR"
  exit
fi

# Move to folder for assignment to test.
cd "$ASSIGNDIR"

# Fail student code that attempts to access secret
if grep -r -e QUICKFEED_SESSION_SECRET ./* ; then
  printf "\n=== Misbehavior Detected: Failed ===\n"
  exit
fi

# Remove student written tests to avoid interference
find . -name '*_test.go' -exec rm -rf {} \;

# Copy tests into student assignments folder for running tests
cp -r $TESTDIR/* $ASSIGNMENTS/

# (ensure) Move to folder for assignment to test.
cd "$ASSIGNDIR"

printf "\n*** Finished Test Setup in %s seconds ***\n" "$(( SECONDS - start ))"

start=$SECONDS
printf "\n*** Running Tests ***\n\n"
QUICKFEED_SESSION_SECRET={{ .RandomSecret }} go test -v -timeout 30s ./... 2>&1
printf "\n*** Finished Running Tests in %s seconds ***\n" "$(( SECONDS - start ))"
