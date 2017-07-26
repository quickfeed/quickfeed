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
