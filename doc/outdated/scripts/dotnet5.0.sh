#image/mcr.microsoft.com/dotnet/sdk:5.0

start=$SECONDS
printf "*** Preparing for Test Execution ***\n"

git config --global url."https://{{ .CreatorAccessToken }}:x-oauth-basic@github.com/".insteadOf "https://github.com/"

ASSIGNMENTS=/quickfeed/assignments
TESTS=/quickfeed/tests
ASSIGNDIR=$ASSIGNMENTS/{{ .AssignmentName }}/
TESTDIR=$TESTS/{{ .AssignmentName }}/

# Fetch student and test repos
git clone {{ .GetURL }} $ASSIGNMENTS
git clone {{ .TestURL }} $TESTS

if [ ! -d "$ASSIGNDIR" ]; then
  printf "Folder $ASSIGNDIR not found in {{ .GetURL }}"
  exit
fi

# Move to folder for assignment to test.
cd $ASSIGNDIR

# Fail student code that attempts to access secret
if grep -r -e QUICKFEED_SESSION_SECRET * ; then
  printf "\n=== Misbehavior Detected: Failed ===\n"
  exit
fi

# Clear access token and the shell history to avoid leaking information to student test code.
git config --global url."https://0:x-oauth-basic@github.com/".insteadOf "https://github.com/"
history -c

# (ensure) Move to folder for assignment to test.
cd $TESTDIR

# Perform lab specific setup
if [ -f "setup.sh" ]; then
    bash setup.sh
fi

printf "\n*** Finished Test Setup in $(( SECONDS - start )) seconds ***\n"

start=$SECONDS
printf "\n*** Running Tests ***\n\n"
QUICKFEED_SESSION_SECRET={{ .RandomSecret }} dotnet test "--logger:console;verbosity=detailed" 2>&1
printf "\n*** Finished Running Tests in $(( SECONDS - start )) seconds ***\n"
