# Deployment Notes for Production Environments

These are some notes related to the running and maintenance of a QuickFeed server.
To set up a production server follow these [instructions](./deploy.md).

## Accessing the Server Machine

To get access to the production server at UiS for maintenance, please send your `.ssh/id_ed25519.pub` public key to meling.
Once your public key has been added to the `authorized_keys` file on the server machine, you can access the machine more easily if you add these entries to your `.ssh/config` file.
Make sure to replace `meling` with your user name.

```text
Host uis
  User meling
  HostName ssh1.ux.uis.no

Host qf qf2
  HostName %h.ux.uis.no
  User quickfeed
  ProxyJump uis
```

With this configuration, you can reach the `qf2` test machine with:

```sh
% ssh qf2
```

## Maintaining the QuickFeed Server

To upgrade the server with new code from the master branch, use the script:

```sh
% ./doc/misc/update.sh
```

This will fetch code from GitHub, recompile the server, stop the server, backup the database, and restart the server.

To stop the server:

```sh
% killall quickfeed
```

To update the server manually:

```sh
% git fetch
% git status
```

Ensure that there are no local changes and the branch can be fast-forwarded.
Otherwise, resolve local changes, preferably as new commits or pull requests to the main repository.
Then run these commands to recompile the server and frontend:

```sh
% git pull
% make install
% make ui
```

To start the server, follow the [instructions herein](./deploy.md).

## Server Logs

We use `logrotate` to maintain server logs.
Configuration file is `/etc/logrotate.d/quickfeed`.
Example configuration is:

```text
/home/quickfeed/quickfeed/qf.log {
        size 5M
        copytruncate
        dateext
        rotate 2
        compress
        maxage 14
}
```

This configuration will rotate the `qf.log` file when its size reaches 5 MB, and start to log to a new file.
The rotated file will be renamed with the current date.
Logrotate will keep the two latest log files in compressed form and will delete them two weeks after the rotation.

For additional information, see the [logrotate manual](https://www.digitalocean.com/community/tutorials/how-to-manage-logfiles-with-logrotate-on-ubuntu-16-04).

## Cron Jobs

TODO(meling) We have not prepared any cron jobs.

Cron is a Linux utility to schedule running of scripts or commands automatically at a specified time.

[Minimal Cron tutorial](https://www.ostechnix.com/a-beginners-guide-to-cron-jobs/).

To add, edit or remove a cron job in the user specific cron table, run `crontab -e`.

**Important:** Cron will send the job outputs an email to every email address provided for the user.
To disable emails, discard job output by adding `>/dev/null 2>&1` at the end of the job description.
