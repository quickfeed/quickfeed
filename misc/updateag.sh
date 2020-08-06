#!/bin/bash

AG_BIN="/home/autograder/go/bin"
AG_ROOT="/home/autograder/quickfeed"

# if a branch name provided, switch to the branch
if [ "$1" != "" ]; then
	echo "Switching to the" $1 "branch"
	if ! git checkout $1; then
		echo "Failed to switch branches"
		exit 1
	fi
fi

# fail if there are some local changes
if ! git diff-index --quiet HEAD --; then
	echo "There are uncommited changes, make sure they are in the main codebase or in a PR before removing."
	exit 1
fi

# git fetch and pull
echo "Fetching changes"
git fetch

echo "Applying changes"
if ! git pull; then
	echo "Failed to apply changes from the remote repository"
	exit 1
fi

# remove old exec backup, then backup the actual one
echo "Backing up executables"
cp $AG_BIN/quickfeed $AG_BIN/quickfeed.bak

# recompile: go and ts
echo "Compiling changes"
cd $AG_ROOT/public
if ! webpack; then
	echo "Failed to compile the client"
	exit 1
fi
cd $AG_ROOT
if ! go install; then
	echo "Failed to compile the server"
	exit 1
fi

# stop the server
echo "Bringing the server down"
if ! killall quickfeed; then
	echo "Failed to stop the server"
	exit 1
fi

# backup the database
echo "Backing up the database"
cp ag.db ./backups/ag.db.$(date +"%m-%d-%y").bak

# start the server
echo "Starting the server"
source ag-env.sh
quickfeed -service.url uis.itest.run -database.file ./ag.db -http.addr :3005 &> ag.log &

# done
echo "All done. Server restarted and running"
