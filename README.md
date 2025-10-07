# QuickFeed: Instant Feedback on Programming Assignments

[![Go Test](https://github.com/quickfeed/quickfeed/workflows/Go%20Test/badge.svg)](https://github.com/quickfeed/quickfeed/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/quickfeed/quickfeed)](https://goreportcard.com/report/github.com/quickfeed/quickfeed)
[![Codecov](https://codecov.io/gh/quickfeed/quickfeed/branch/master/graph/badge.svg)](https://codecov.io/gh/quickfeed/quickfeed)
[![golangci-lint](https://github.com/quickfeed/quickfeed/workflows/golangci-lint/badge.svg)](https://github.com/quickfeed/quickfeed/actions)

QuickFeed is a Go/TypeScript web application that automates feedback on programming assignments with tight GitHub integration.
It provisions course repositories, runs tests in Docker, tracks progress, and helps teachers review and release results quickly.

## Who is this for?

- Teachers: Set up a course organization, publish assignments and tests, review submissions, and release results.
- Students: Get instant feedback on submissions and see progress for each assignment.
- Developers: Contribute to a modern Go backend with gRPC/Connect and a React/TypeScript frontend.

## Contributing

[Details about contributing](/doc/contributing.md)

## Documentation

[Start here: QuickFeed documentation](/doc/main-doc.md)

Quick links:

- [Overview of the code base](/doc/qf-overview.md)
- [Teacher user manual](/doc/teacher.md)
- [Developer guide](/doc/dev.md)
- [Deployment](/doc/deploy-prod.md)
- [Metrics](/doc/metrics.md)
- [License](/LICENSE)

## Quick start (development)

Prerequisites: Go, Node.js, Docker.

Common tasks (see Makefile for more):

```sh
# Download Go dependencies
make download

# Build backend
make install

# Build frontend
make ui

# Run all tests
make test
```

Local development server:

```sh
# One-time setup
cp .env-template .env
# Edit .env for localhost development

# Start the server (defaults to :8080 in dev mode)
PORT=8080 quickfeed -dev
```
