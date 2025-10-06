# QuickFeed Documentation

## Learn more about the QuickFeed system

Read [Overview of QuickFeed's code base](qf-overview.md)

## For Teachers

Please review the [User Manual](teacher.md) to setup QuickFeed.
You may also want to copy the [sign up instructions](templates/signup.md) and [lab submission instructions](templates/lab-submission.md), and make the necessary adjustments for your course.

## Technology Stack

QuickFeed depends on these technologies.

- [Go](https://golang.org/doc/code.html)
- [TypeScript](https://www.typescriptlang.org/)
- Buf's [connect-go](https://buf.build/blog/connect-a-better-grpc) for gRPC
- Buf's [connect-web](https://buf.build/blog/connect-web-protobuf-grpc-in-the-browser) replaces gRPC-Web
- [Protocol Buffers](https://developers.google.com/protocol-buffers/docs/proto3)

## Recommended VSCode Plugins

View [Recommended extensions](../.vscode/extensions.json)

## For developers

View [QuickFeed Developer Guide](dev.md)

## Kit - release guide

The kit package contains helper functions to be used in course specific test cases so that QuickFeed can compute a score for the code submitted by students.
It also contains code to help score multiple choice exercises and command line execution that returns a given expected output.

View [Preparing a new Release of QuickFeed's kit Module](release-guide.md)

## Deploy to production

View [Deployment Notes for Production Environments](deploy-prod.md)

## Third party access

View [Accessing QuickFeed with Third-party Applications](third-party-access.md)
