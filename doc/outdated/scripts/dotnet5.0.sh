#image/mcr.microsoft.com/dotnet/sdk:5.0

start=$SECONDS
printf "*** Preparing for Test Execution ***\n"

ASSIGNDIR=/quickfeed/assignments/{{ .AssignmentName }}/
TESTDIR=/quickfeed/tests/{{ .AssignmentName }}/

if [ ! -d "$ASSIGNDIR" ]; then
  printf "Folder $ASSIGNDIR not found"
  exit
fi

# Move to folder for assignment to test.
cd $ASSIGNDIR

# Fail student code that attempts to access secret
if grep -r -e QUICKFEED_SESSION_SECRET * ; then
  printf "\n=== Misbehavior Detected: Failed ===\n"
  exit
fi

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
