# Quickfeed Deployments

- [Quickfeed Deployments](#quickfeed-deployments)
  - [Setup](#setup)
    - [Install Tools for Deployment](#install-tools-for-deployment)
    - [Install Tools for Development](#install-tools-for-development)
    - [The PORT environment variable](#the-port-environment-variable)
    - [Preparing the Environment for Production](#preparing-the-environment-for-production)
    - [Preparing the Environment for Testing](#preparing-the-environment-for-testing)
    - [First-time Installation](#first-time-installation)
    - [Configuring Docker](#configuring-docker)
    - [Configuring Fixed IP and Router](#configuring-fixed-ip-and-router)
  - [Building QuickFeed Server](#building-quickfeed-server)
  - [Running QuickFeed Server](#running-quickfeed-server)
    - [Flags](#flags)
    - [Running Server on a Privileged Port](#running-server-on-a-privileged-port)
    - [Using GitHub Webhooks When Running Server On Localhost](#using-github-webhooks-when-running-server-on-localhost)
  - [Authentication Secret Handling in QuickFeed](#authentication-secret-handling-in-quickfeed)
  - [Troubleshooting](#troubleshooting)

## Setup

### Install Tools for Deployment

This assumes you have [homebrew](https://brew.sh/?gad_source=1&gclid=Cj0KCQiA-aK8BhCDARIsAL_-H9nbD3jqSHwIYOpn3hzxCU-pmItDSXIroK-fhUm7KYkxUNdJwzbrElkaAglyEALw_wcB&gclsrc=aw.ds) and [make](https://www.gnu.org/software/make/) installed.

**For systems without homebrew**, the make target should list well-known packages available on most Unix distributions. View [makefile](../Makefile)

```sh
% make brew
```

### Install Tools for Development

Tools for development are managed via `go.mod` and does not need to be installed separately.
They are used via the `go tool` command, which are invoked via:

```sh
% make proto
```

For additional details, see the file `buf.gen.yaml` in the repository's root folder:

```yaml
version: v2
plugins:
  - local: ["go", "tool", "protoc-gen-go-patch"]
    out: ./
    opt:
      - plugin=go
      - paths=source_relative
  - local: ["go", "tool", "protoc-gen-connect-go"]
    out: ./
    opt:
      - paths=source_relative
```

### The PORT environment variable

The `PORT` environment variable can be changed to any valid port, and we strongly advise running Quickfeed a secure port in production.
It can be set to an unprivileged port and prevent permission issues when [running on a privileged port](#running-server-on-a-privileged-port).
This is helpful for debugging, as it allows you to run Quickfeed directly by setting program to `"${workspaceFolder}/main.go"`, in the launch profile.
Alternatively, you can grant the Quickfeed binary access to the privileged port and run it as an executable.
The launch profile program should then be set to `"${env:GOPATH}/bin/quickfeed"`.

### Preparing the Environment for Production

QuickFeed expects the `.env` file to contain certain environment variables.
For a first-time installation, the `.env` file is not present.
However, the `.env-template` file contains a template that can be copied and modified.
The following is an example production deployment on the `example.com` domain.

```shell
# GitHub App IDs and secrets for deployment
QUICKFEED_APP_ID=""
QUICKFEED_APP_KEY=$QUICKFEED/internal/config/github/quickfeed.pem
QUICKFEED_CLIENT_ID=""
QUICKFEED_CLIENT_SECRET=""
QUICKFEED_WEBHOOK_SECRET=""

# QuickFeed server domain or ip
DOMAIN="example.com"
# Quickfeed server port
PORT="443"

# Comma-separated list of domains to allow certificates for.
# IP addresses and "localhost" are *not* valid.
# The whitelist must also include the domain defined above.
QUICKFEED_WHITELIST="example.com"
```

You only need to edit the `$DOMAIN` environment variable to point to your public landing page for QuickFeed.
The [QuickFeed App installation process](#first-time-installation) will guide you through the rest of the setup,
setting the environment variables in your `.env` file and saving the `quickfeed.pem` file.

### Preparing the Environment for Testing

For a localhost test deployment, you can use the default certificate path in `~/.quickfeed/certs/`.
Generate self-signed certificates using the `-gencert` flag before starting the server:

```shell
% quickfeed -gencert
```

This will generate self-signed certificates and add them to your system's trust store.
The certificates are stored in `~/.quickfeed/certs/<domain>/` by default.

Alternatively, you can specify custom certificate file paths by setting the following environment variables in your `.env` file:

```shell
# Optional: Custom certificate paths (defaults to ~/.quickfeed/certs/<domain>/)
# QUICKFEED_KEY_FILE=$HOME/.quickfeed/certs/127.0.0.1/privkey.pem
# QUICKFEED_CERT_FILE=$HOME/.quickfeed/certs/127.0.0.1/fullchain.pem

# QuickFeed server domain or ip
DOMAIN="127.0.0.1"
# Quickfeed server port
PORT="443"
```

The `QUICKFEED_WHITELIST` must be removed from your `.env` file for localhost deployments.

### First-time Installation

To start the server for first-time installation, use the `-new` flag.

```shell
% make install
% quickfeed -new
2022/09/11 16:45:22 running: go list -m -f {{.Dir}}
2022/09/11 16:45:22 Loading environment variables from /Users/meling/work/quickfeed/.env
2022/09/11 16:45:22 Important: The GitHub user that installs the QuickFeed App will become the server's admin user.
2022/09/11 16:45:22 Go to https://example.com/manifest to install the QuickFeed GitHub App.
2022/09/11 16:46:00 Successfully installed the QuickFeed GitHub App.
2022/09/11 16:46:00 Loading environment variables from /Users/meling/work/quickfeed/.env
2022/09/11 16:46:00 Starting QuickFeed in production mode on example.com
```

After starting the server you should see various configuration files saved to `internal/config`:

```shell
% tree internal/config/
internal/config/
├── certs
│   ├── acme_account+key
│   └── example.com
└── github
    └── quickfeed.pem
```

In addition, your `.env` file should be populated with important secrets that should be kept away from prying eyes.

```shell
% cat .env
# GitHub App IDs and secrets for deployment
QUICKFEED_APP_ID=<6 digit ID>
QUICKFEED_APP_KEY=/Users/meling/work/quickfeed/internal/config/github/quickfeed.pem
QUICKFEED_CLIENT_ID=Iv1.<16 chars of identifying data>
QUICKFEED_CLIENT_SECRET=<40 chars of secret data>
QUICKFEED_WEBHOOK_SECRET=<40 chars of secret data>
```

### Configuring Docker

To ensure that Docker containers has access to networking, you may need to set up IPv4 port forwarding on your server machine:

```sh
sudo sysctl net.ipv4.ip_forward=1
sudo sysctl -p
sudo service docker restart
```

### Configuring Fixed IP and Router

In your domain name provider, configure your IP and domain name; for instance:

```text
Type         Host          Value                TTL
A Record     cyclone       92.221.105.172       5 min
```

Set up port forwarding on your router.
External ports 80/443 maps to internal ports 80/443 for TCP.

## Building QuickFeed Server

After editing files in the `public` folder, run the following command.
This should also work while the application is running.

```sh
% make ui
```

Build the `quickfeed` server.

```sh
% make install
```

After editing any of the `.proto` files you will need to recompile the protobuf files, run the following command.

```sh
% make proto
```

This may require you to run both `make install` and `make ui`.

## Running QuickFeed Server

To run in production mode on `$DOMAIN` using default values:

```sh
% quickfeed &> quickfeed.log &
```

To run in development mode on localhost:

```sh
% quickfeed -dev &> quickfeed.log &
```

To view the full usage details:

```sh
% quickfeed -help
```

### Flags

| **Flag**        | **Description**                                                           | **Example** |
| --------------- | ------------------------------------------------------------------------- | ----------- |
| `database.file` | Path to QuickFeed database                                                | `qf.db`     |
| `http.public`   | Path to content to serve                                                  |             |
| `dev`           | Run local development server with self-signed certificates and watch mode |             |
| `gencert`       | Generate self-signed certificates for development                         |             |
| `new`           | Create a new QuickFeed App                                                |             |
| `secret`        | Create new secret for JWT signing                                         |             |

### Running Server on a Privileged Port

It is possible to run server in development mode on different ports by updating the environment variable `PORT`, which is by default the standard HTTPS port: `443`.
If the quickfeed binary cannot access port `:443` on your Linux system, you can enable it by running:

```sh
sudo setcap CAP_NET_BIND_SERVICE=+eip /path/to/binary/quickfeed
```

Note that you will need to repeat this step each time you recompile the server.

### Using GitHub Webhooks When Running Server On Localhost

GitHub webhooks cannot send events directly to your server if it runs on localhost.
However, it is possible to setup a tunneling service that will be listening to the events coming from webhooks and redirecting them to the locally deployed server.

One of the many options is [ngrok](https://ngrok.com/). To use ngrok you have to create a free account and download ngrok.
After that it will be possible to receive webhook events on QuickFeed server running on localhost by performing a few steps.

1. Start ngrok: `ngrok http 443` - assuming the server runs on port `:443`.
2. ngrok will generate a new endpoint URL.
   Copy the urls an update webhook callback information in your GitHub app to point to this URL.
   E.g., `https://de08-2a01-799-4df-d900-b5af-5adc-a42a-bcf.eu.ngrok.io/hook/`.

After that any webhook events your GitHub app is subscribed to will send payload to this URL, and ngrok will redirect them to the `/hooks` endpoint of the QuickFeed server running on the given port number.

Note that ngrok generates a new URL every time it is restarted and you will need to update webhook callback details unless you want to subscribe to the paid version of ngrok that supports static callback URLs.

## Authentication Secret Handling in QuickFeed

QuickFeed uses the `QUICKFEED_AUTH_SECRET` environment variable to sign JWT tokens for user authentication.
Two modes are supported for handling the authentication secret:

1. Starting the server with the `-secret` flag; this will
   - Generate a new random authentication secret.
   - Save the new secret in the `.env` file, making it the default for future server restarts.
   - Log out all currently logged-in users, requiring them to sign in again.
   - This mode is useful for deployments where periodic secret rotation is necessary.
2. Starting the server without the `-secret` flag; this will
   - Use the previously saved secret from the `.env` file.
   - Allow server restarts without logging users out.
   - Ensure that existing JWT tokens remain valid across restarts.

For custom secret management, users can manually set the `QUICKFEED_AUTH_SECRET` environment variable.
This will override the secret saved in the `.env` file and is useful for deployments where the secret is managed externally.

**Security Warning:**
It is important that this secret is kept secure, as exposure can lead to compromised JWT tokens.
Always use a long randomly generated secret value to maintain security.

## Troubleshooting

If `go install` fails with the following (on Ubuntu):

```sh
cgo: exec gcc-5: exec: "gcc-5": executable file not found in $PATH
```

Then run and retry `go install`:

```sh
% brew install gcc@5
% go install
```
