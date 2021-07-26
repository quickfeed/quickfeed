# Testing database

The database (`testing.db`) is using the updated Quickfeed schema without any json-encoded string fields, `DisableForeignKeyConstraintWhenMigrating` (`gormdb.go:48`) should not be set to true.

- The easiest way to inspect or edit the database is to use a sqlite database browser GUI, for example the [DB Browser](https://sqlitebrowser.org/) or any other alternative.

- User with ID 1 is a super user registered as `CourseCreator` of all courses in the database. The `RemoteIdentity` record for this user can be replaced with Github's remote ID and access token of your Github user to login as super user.

- Course with ID 1 is based on the [qf406](https://github.com/qf406) created specifically for testing purposes. This repository can be used for any kind of testing: adding or removing users and groups/teams, adding assignments and tests, pushing submissions to student repos and so on. Request an invite to the repository if you want to use it for testing.
