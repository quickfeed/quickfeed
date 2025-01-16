# Quickfeed Documentation

- [Quickfeed Documentation](#quickfeed-documentation)
  - [Learn more about the quickfeed system](#learn-more-about-the-quickfeed-system)
  - [Technology Stack](#technology-stack)
  - [Recommended VSCode Plugins](#recommended-vscode-plugins)
  - [Installation](#installation)
  - [For Teachers](#for-teachers)
  - [For developers](#for-developers)
  - [Kit - release guide](#kit---release-guide)
  - [Deploy to production](#deploy-to-production)
  - [Third party access](#third-party-access)

## Learn more about the quickfeed system

Read [Overview of Quickfeed's code base](qf-overview.md)

## Technology Stack

QuickFeed depends on these technologies.

- [Go](https://golang.org/doc/code.html)
- [TypeScript](https://www.typescriptlang.org/)
- Buf's [connect-go](https://buf.build/blog/connect-a-better-grpc) for gRPC
- Buf's [connect-web](https://buf.build/blog/connect-web-protobuf-grpc-in-the-browser) replaces gRPC-Web
- [Protocol Buffers](https://developers.google.com/protocol-buffers/docs/proto3)

## Recommended VSCode Plugins

View [Recommended extensions](../.vscode/extensions.json)

## Installation

View [Installation instructions for QuickFeed](deploy.md).

## For Teachers

Please review the [User Manual](teacher.md) to setup quickfeed, and you'ed also want to copy the [sign up instructions](templates/signup.md) and [lab submission instructions](templates/lab-submission.md), and make the necessary adjustments for your course.

## For developers

View [QuickFeed Developer Guide](dev.md)

## Kit - release guide

View [Preparing a new Release of QuickFeed's kit Module](release-guide.md)

## Deploy to production

View [Deployment Notes for Production Environments](deploy-prod.md)

## Third party access

View [Accessing QuickFeed with Third-party Applications](third-party-access.md)
