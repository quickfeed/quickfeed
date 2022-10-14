# Quickfeed Deployments

## Table of Contents

- [Technology Stack](#technology-stack)
- [Preparing the Environment](#preparing-the-environment)
  - [Configuring Docker](#configuring-docker)
  - [Setup Environment Variables](#setup-environment-variables)
- [Bare Metal Deployment](#bare-metal-deployment)
  - [Configure Fixed IP and Router](#configure-fixed-ip-and-router)
  - [Install Tools for Deployment](#install-tools-for-deployment)
  - [Install Tools for Development](#install-tools-for-development)
  - [Build and Run QuickFeed Server](#build-and-run-quickfeed-server)
    - [Troubleshooting](#troubleshooting)
  - [Running the QuickFeed Server Details](#running-the-quickfeed-server-details)
    - [Flags](#flags)

## Technology Stack

QuickFeed depends on these technologies.

- [Go](https://golang.org/doc/code.html)
- [TypeScript](https://www.typescriptlang.org/)
- Buf's [connect-go](https://buf.build/blog/connect-a-better-grpc) for gRPC
- Buf's [connect-web](https://buf.build/blog/connect-web-protobuf-grpc-in-the-browser) replaces gRPC-Web
- [Protocol Buffers](https://developers.google.com/protocol-buffers/docs/proto3)
- TODO remove when no longer used [gRPC-Web](https://github.com/grpc/grpc-web) (currently only used in the frontend)

### Recommended VSCode Plugins

- Go
- vscode-proto3
- Code Spell Checker
- ESLint
- markdownlint
- Better Comments
- GitLens
- Git History Diff
- SQLite

## Updated Deployment Instructions for new GitHub App and Autocert

### Configure .env

TODO(meling) add missing QUICKFEED_WEBHOOK_SECRET and

If your `.env` file is has no keys and your `quickfeed.pem` does not exist, you need to install QuickFeed's GitHub App.

```shell
# GitHub App IDs and secrets for localhost deployment
QUICKFEED_APP_ID=""
QUICKFEED_APP_KEY=$QUICKFEED/internal/config/github/quickfeed.pem
QUICKFEED_CLIENT_ID=""
QUICKFEED_CLIENT_SECRET=""

# Quickfeed server domain or ip
DOMAIN="example.com"

# Comma-separated list of domains to allow certificates for.
# IP addresses and "localhost" are *not* valid.
# The whitelist must also include the domain defined above.
QUICKFEED_WHITELIST="example.com"
```

### Starting server and installing QuickFeed's GitHub App

To start the server for first-time installation, use the `-new` flag.

```shell
% make install
% quickfeed -new
2022/09/11 16:45:22 running: go list -m -f {{.Dir}}
2022/09/11 16:45:22 Loading environment variables from /Users/meling/work/quickfeed/.env
2022/09/11 16:45:22 Important: The GitHub user that installs the QuickFeed App will become the server's admin user.
2022/09/11 16:45:22 Go to https://example.com/manifest to install the QuickFeed GitHub App.
2022/09/11 16:45:43 http: TLS handshake error from 192.168.86.1:52823: write tcp 192.168.86.32:443->192.168.86.1:52823: i/o timeout
2022/09/11 16:45:43 http: TLS handshake error from 192.168.86.1:52824: tls: client using inappropriate protocol fallback
2022/09/11 16:46:00 Successfully installed the QuickFeed GitHub App.
2022/09/11 16:46:00 Loading environment variables from /Users/meling/work/quickfeed/.env
2022/09/11 16:46:00 Starting QuickFeed in production mode on example.com
```

After starting the server you should see various configuration files saved for in `internal/config`:

```shell
% tree internal/config/
internal/config/
├── certs
│   ├── acme_account+key
│   └── example.com
└── github
    └── quickfeed.pem
```

In addition your `.env` file should be populated with important secrets that should be kept away from prying eyes.

```shell
% cat .env
# GitHub App IDs and secrets for localhost deployment
QUICKFEED_APP_ID=<6 digit ID>
QUICKFEED_APP_KEY=/Users/meling/work/quickfeed/internal/config/github/quickfeed.pem
QUICKFEED_CLIENT_ID=Iv1.<16 chars of identifying data>
QUICKFEED_CLIENT_SECRET=<40 chars of secret data>
```

## Localhost: Updated Deployment Instructions for new GitHub App and Autocert

### Configure .env for localhost deployment

If your `.env` file is has no keys and your `quickfeed.pem` does not exist, you need to install QuickFeed's GitHub App.
For localhost deployment, you need to specify the file names for the self-signed certificates.

```shell
# GitHub App IDs and secrets for localhost deployment
QUICKFEED_APP_ID=""
QUICKFEED_APP_KEY=$QUICKFEED/internal/config/github/quickfeed.pem
QUICKFEED_CLIENT_ID=""
QUICKFEED_CLIENT_SECRET=""
# Certificate chain and private key file
QUICKFEED_KEY_FILE=$QUICKFEED/internal/config/certs/privkey.pem
QUICKFEED_CERT_FILE=$QUICKFEED/internal/config/certs/fullchain.pem

# Quickfeed server domain or ip
DOMAIN="localhost"
```

### Starting server and installing QuickFeed's GitHub App

To start the server for first-time installation, use the `-new` flag.
Since this is a development server, you must also supply the `dev` flag.

```shell
% make install
% quickfeed -dev -new
2022/09/11 20:28:08 running: go list -m -f {{.Dir}}
2022/09/11 20:28:08 Loading environment variables from /Users/meling/work/quickfeed/.env
2022/09/11 20:28:08 Generating self-signed certificates.
2022/09/11 20:28:08 Certificates successfully generated at: internal/config/certs
2022/09/11 20:28:08 When running with self-signed certificates on localhost, browsers will complain that the connection is not private.
2022/09/11 20:28:08 To run the server from localhost, you will need to manually bypass the browser warning.
WARNING: You are creating an app on localhost. Only for development purposes. Continue? (Y/n) y
2022/09/11 20:28:11 Important: The GitHub user that installs the QuickFeed App will become the server's admin user.
2022/09/11 20:28:11 Go to https://localhost/manifest to install the QuickFeed GitHub App.
2022/09/11 20:28:51 http: TLS handshake error from [::1]:61014: EOF
2022/09/11 20:28:51 http: TLS handshake error from [::1]:61015: EOF
2022/09/11 20:29:09 Successfully installed the QuickFeed GitHub App.
2022/09/11 20:29:09 Loading environment variables from /Users/meling/work/quickfeed/.env
2022/09/11 20:29:09 Starting QuickFeed in development mode on :443
2022/09/11 20:29:09 Existing credentials successfully loaded.
2022/09/11 20:29:09 When running with self-signed certificates on localhost, browsers will complain that the connection is not private.
2022/09/11 20:29:09 To run the server from localhost, you will need to manually bypass the browser warning.
```

After starting the server you should see various configuration files saved for in `internal/config`:

```shell
% tree internal/config/
internal/config
├── certs
│   ├── fullchain.pem
│   └── privkey.pem
└── github
    └── quickfeed.pem
```

In addition your `.env` file should be populated with important secrets that should be kept away from prying eyes.

```shell
% cat .env
# GitHub App IDs and secrets for localhost deployment
QUICKFEED_APP_ID=<6 digit ID>
QUICKFEED_APP_KEY=/Users/meling/work/quickfeed/internal/config/github/quickfeed.pem
QUICKFEED_CLIENT_ID=Iv1.<16 chars of identifying data>
QUICKFEED_CLIENT_SECRET=<40 chars of secret data>
```

## Preparing the Environment

TODO(meling): Borrow instructions/scripts from ag server.

### Configuring Docker

To ensure that Docker containers has access to networking, you may need to set up IPv4 port forwarding on your server machine:

```sh
sudo sysctl net.ipv4.ip_forward=1
sudo sysctl -p
sudo service docker restart
```

### Setup Environment Variables

```sh
% cd $QUICKFEED
% cp .env-template .env
# Edit .env according to your domain and quickfeed ports
```

The `$DOMAIN` should be set to your public landing page for QuickFeed, e.g., `www.my-quickfeed.com`.

## Bare Metal Deployment

TODO(meling) move stuff here

### Configure Fixed IP and Router

In your domain name provider, configure your IP and domain name; for instance:

```text
Type         Host          Value                TTL
A Record     cyclone       92.221.105.172       5 min
```

Set up port forwarding on your router.
External ports 80/443 maps to internal ports 80/443 for TCP.

### Install Tools for Deployment

This assumes you have homebrew installed.
For systems without homebrew, the make target should list well-known packages available on most Unix distributions.

```sh
% make brew
```

### Install Tools for Development

The development tools are only needed for development, and can be skipped for deployment only.
To install:

```sh
% make devtools
```

The `devtools` make target will download and install various Protobuf compiler plugins and the grpcweb Protobuf compiler.

### Build and Run QuickFeed Server

After editing files in the `public` folder, run the following command.
This should also work while the application is running.

```bash
% make ui
```

Build and run the `quickfeed` server; here we use all default values:

```bash
% go install
% quickfeed &> quickfeed.log &
```

#### Troubleshooting

If `go install` fails with the following (on Ubuntu):

```sh
cgo: exec gcc-5: exec: "gcc-5": executable file not found in $PATH
```

Then run and retry `go install`:

```bash
% brew install gcc@5
% go install
```

### Running the QuickFeed Server Details

The following provides additional details for running QuickFeed.
Before running the QuickFeed server, you need to configure [GitHub](./github.md).

The command line arguments for the QuickFeed server looks roughly like this:

```sh
quickfeed -database.file <path to database> -http.addr <HTTP listener address>
```

To view the full usage details:

```sh
quickfeed -help
```

Here is an example with all default values:

```sh
quickfeed &> quickfeed.log &
```

_As a bootstrap mechanism, the first user to sign in, automatically becomes administrator for the system._

#### Flags

| **Flag**        | **Description**                                      | **Example** |
| --------------- | ---------------------------------------------------- | ----------- |
| `database.file` | Path to QuickFeed database                           | `qf.db`     |
| `http.addr`     | Listener address for HTTP service                    | `:8081`     |
| `dev`           | Run development server with self-signed certificates |             |
| `new`           | Create a new QuickFeed App                           |             |
