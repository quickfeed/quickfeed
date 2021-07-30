# Test Database

The `test.db` database uses the updated Quickfeed schema without any json-encoded string fields.
Hence, the `DisableForeignKeyConstraintWhenMigrating` (`gormdb.go:48`) should not be set to true.

- The easiest way to inspect or edit the database is to use a sqlite database browser GUI, for example the [DB Browser](https://sqlitebrowser.org/) or the `sqlite3` command-line tool.

- User with ID 1 is a super user registered as `CourseCreator` of all courses in the database.
  The `RemoteIdentity` record for this user can be replaced with GitHub's remote ID and access token of your GitHub user to login as super user. Additionally, replace `login` field of the `User` record with your GitHub login.
  
  To update with the command-line tool:

  ```sql
  sqlite3 test.db
  update remote_identities
  set remote_id = {id: int}, access_token = {token: string}
  where id = 1;

  update users
  set login = {login: string}
  where id = 1;
  ```

- Course with ID 1 is based on the [qf406](https://github.com/qf406) organization created specifically for testing purposes.
  This organization can be used for any kind of testing: adding or removing users and groups/teams, adding assignments and tests, pushing submissions to student repos and so on.
  Request an invite to the repository if you want to use it for testing.
