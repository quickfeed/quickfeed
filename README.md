# aguis (name pending...) [![Build Status](https://travis-ci.org/autograde/aguis.svg?branch=master)](https://travis-ci.org/autograde/aguis) [![Go Report Card](https://goreportcard.com/badge/github.com/autograde/aguis)](https://goreportcard.com/report/github.com/autograde/aguis)
## Download
   ```sh
   go get -u github.com/autograde/aguis/...
   ```
## Install
   ```sh
   # Debug build
   go install -tags debug github.com/autograde/aguis/...
   # Release build
   go install github.com/autograde/aguis/...
   ```
## Run
   ```sh
   # Server listening on port 8080 serving static files from /public at https://example.com/.
   aguis -service.url example.com -http.addr :8080 -http.public /public
   ```
