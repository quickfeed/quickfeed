#!/bin/bash

if sysctl -q -n net.ipv4.ip_forward | grep -q '0'; then
	echo "IP packet forwarding disabled on host machine; run these commands before restarting again:"
	echo ""
	echo "sudo sysctl net.ipv4.ip_forward=1"
	echo "sudo sysctl -p"
	echo "sudo service docker restart"
	exit 1
fi

QUICKFEED=$HOME/quickfeed
GOBIN=$HOME/go/bin
DATABASE=qf.db
LOGFILE=quickfeed.log
BACKUP=$QUICKFEED/backups

cd $QUICKFEED

echo "Backing up executables"
cp $GOBIN/quickfeed $BACKUP/quickfeed.$(date +"%m-%d-%y").bak

echo "Backing up the database file"
cp $QUICKFEED/$DATABASE $BACKUP/$DATABASE.$(date +"%m-%d-%y").bak

echo "Backing up the $LOGFILE file"
cp $HOME/logs/$LOGFILE $BACKUP/$LOGFILE.$(date +"%m-%d-%y").bak

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

echo "Fetching changes"
git fetch

echo "Applying changes"
if ! git pull; then
	echo "Failed to apply changes from the remote repository"
	exit 1
fi

echo "Running webpack"
cd $QUICKFEED/public
if ! webpack; then
	echo "Failed to compile the client"
	exit 1
fi

echo "Running go install"
cd $QUICKFEED
if ! go install; then
	echo "Failed to compile the server"
	exit 1
fi

if pgrep quickfeed &> /dev/null; then
	echo "Server running; shutting down server..."
	if ! killall quickfeed; then
		echo "Failed to stop the server; trying to restart..."
	fi
fi

echo "Starting the QuickFeed server"
source quickfeed-env.sh
quickfeed -service.url uis.itest.run &> $HOME/logs/$LOGFILE &

if ! pgrep quickfeed &> /dev/null; then
    echo "Failed to start the server"
    exit 1
fi

echo "All done. QuickFeed server restarted and running"
exit 0
