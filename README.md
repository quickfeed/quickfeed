# agserver [![Build Status](https://travis-ci.org/autograde/aguis.svg?branch=master)](https://travis-ci.org/autograde/aguis) [![Go Report Card](https://goreportcard.com/badge/github.com/autograde/aguis)](https://goreportcard.com/report/github.com/autograde/aguis) [![Coverage Status](https://coveralls.io/repos/github/autograde/aguis/badge.svg?branch=master)](https://coveralls.io/github/autograde/aguis?branch=master)

## Roles and Concepts

The system has three **user** roles.
- **Administrators** can create new courses and promote other users to become administrator. It is common that all teachers that are responsible for one or more courses be an administrator. The administrator role is system-wide.
- **Teachers** are associated with one or more courses. A teacher can view results for all students in his course(s). A course can have many teachers. A teacher is anyone associated with the course that are not students, such as professors and teaching assistants.
- **Students** are associated with one or more courses. A student can view his own results and progress on individual assignments.

The following concepts are important to understand.
- **Assignments** are organized into folders in a git repository.
  - **Individual assignments** are solved by one student. There is one repository for individual assignments.
  - **Group assignments** are solved by a group of students. There is one repository for group assignments.
- **Submissions** are made by a student submitting his code to a supported git service provider (e.g. github or gitlab). 

## Download and install

   ```sh
   go get -u github.com/autograde/aguis
   ```

## Run

   ```sh
   # Server listening on port 8080 serving static files from /public at https://example.com/.
   aguis -service.url example.com -http.addr :8080 -http.public /public
   ```
*As a bootstrap mechanism, the first user to sign in is automatically made administrator for the system.*

## Install for React web development

   ```sh
   cd public
   npm install
   webpack
   ```

## Development

To ensure that webpack bundle files are updated when you pull in changes or rebase from the repository you can add the following script to the files `post-merge` (invoked on git pull) and `post-rewrite` (invoked on git rebase) in the `.git/hooks/` folder.
   ```sh
   #!/bin/sh
   cd $GOPATH/src/github.com/autograde/aguis/public
   webpack
   ```
If you don't want to run `webpack` to create the bundle files on git pull/rebase, you will need to manually run `webpack` in the `public` folder.
