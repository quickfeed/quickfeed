#!/bin/sh

#   This script backs up the database.
#   It is run by cron every day at 03:00 (0 3 * * *).
#   Use `crontab -e` to add or edit the schedule.
#   Use `crontab -l` to review the schedule.

# Path to QuickFeed
QUICKFEED=$HOME/quickfeed

# Name of database file to backup
DATABASE=qf.db

# Folder to place backups in
# Be sure to create this folder before adding this script to crontab.
BACKUP=$QUICKFEED/backups

cp $QUICKFEED/$DATABASE $BACKUP/$DATABASE.$(date +"daily-%m-%d-%y").bak
