#image/quickfeed:go

start=$SECONDS
printf "*** Preparing for Test Execution ***\n"

git config --global url."https://{{ .CreatorAccessToken }}:x-oauth-basic@github.com/".insteadOf "https://github.com/"

ASSIGNMENTS=/quickfeed/assignments
TESTDIR=/quickfeed/tests
ASSIGNDIR=$ASSIGNMENTS/{{ .AssignmentName }}/

# Fetch student and test repos
git clone {{ .GetURL }} $ASSIGNMENTS
git clone {{ .TestURL }} $TESTDIR

if [ ! -d "$ASSIGNDIR" ]; then
  printf "No code to test for {{ .GetURL }}/{{ .AssignmentName }}"
  exit
fi

# Move to folder for assignment to test.
cd $ASSIGNDIR

# Fail student code that attempts to access secret
if grep -r -e QUICKFEED_SESSION_SECRET * ; then
  printf "\n=== Misbehavior Detected: Failed ===\n"
  exit
fi

# Remove student written tests to avoid interference
find . -name '*_test.go' -exec rm -rf {} \;
rm -f setup.sh

# Copy tests into student assignments folder for running tests
cp -r $TESTDIR/* $ASSIGNMENTS/

# Clear access token and the shell history to avoid leaking information to student test code.
git config --global url."https://0:x-oauth-basic@github.com/".insteadOf "https://github.com/"
history -c

# (ensure) Move to folder for assignment to test.
cd $ASSIGNDIR

# Perform lab specific setup
if [ -f "setup.sh" ]; then
    bash setup.sh
fi

printf "\n*** Finished Test Setup in $(( SECONDS - start )) seconds ***\n"

start=$SECONDS
printf "\n*** Running Tests ***\n\n"
QUICKFEED_SESSION_SECRET={{ .RandomSecret }} go test -v -timeout 30s ./... 2>&1
printf "\n*** Finished Running Tests in $(( SECONDS - start )) seconds ***\n"
