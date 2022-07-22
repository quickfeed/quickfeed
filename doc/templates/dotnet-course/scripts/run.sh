#image/mcr.microsoft.com/dotnet/sdk:5.0

start=$SECONDS
printf "*** Preparing for Test Execution ***\n"

cd "$TESTS/$CURRENT"

# Perform lab specific setup
if [ -f "setup.sh" ]; then
    bash setup.sh
fi

printf "\n*** Finished Test Setup in %s seconds ***\n" "$(( SECONDS - start ))"
start=$SECONDS
printf "\n*** Running Tests ***\n\n"
dotnet test "--logger:console;verbosity=detailed" 2>&1
printf "\n*** Finished Running Tests in %s seconds ***\n" "$(( SECONDS - start ))"
