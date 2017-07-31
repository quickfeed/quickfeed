# agserver [![Build Status](https://travis-ci.org/autograde/aguis.svg?branch=master)](https://travis-ci.org/autograde/aguis) [![Go Report Card](https://goreportcard.com/badge/github.com/autograde/aguis)](https://goreportcard.com/report/github.com/autograde/aguis) [![Coverage Status](https://coveralls.io/repos/github/autograde/aguis/badge.svg?branch=master)](https://coveralls.io/github/autograde/aguis?branch=master)

## Download and install

   ```sh
   go get -u github.com/autograde/aguis
   ```

## Run

   ```sh
   # Server listening on port 8080 serving static files from /public at https://example.com/.
   aguis -service.url example.com -http.addr :8080 -http.public /public
   # Set the admin user for github provider
   agctl set admin -id 1 -provider github -username <githubusername>
   ```

## Install for React web development

   ```sh
   cd public
   npm install
   ```

## Development

To ensure that webpack bundle files are updated when you pull in changes or rebase from the repository you can add the following script to the files `post-merge` (invoked on git pull) and `post-rewrite` (invoked on git rebase) in the `.git/hooks/` folder.
   ```sh
   #!/bin/sh
   cd $GOPATH/src/github.com/autograde/aguis/public
   webpack
   ```
If you don't want to run `webpack` to create the bundle files on git pull/rebase, you will need to manually run `webpack` in the `public` folder.
