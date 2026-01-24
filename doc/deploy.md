# Quickfeed Deployments

- [Quickfeed Deployments](#quickfeed-deployments)
  - [Setup](#setup)
    - [Install Tools for Deployment](#install-tools-for-deployment)
    - [Install Tools for Development](#install-tools-for-development)
    - [Preparing the Environment for Development](#preparing-the-environment-for-development)
    - [Preparing the Environment for Production](#preparing-the-environment-for-production)
    - [Installation](#installation)
  - [Building QuickFeed Server](#building-quickfeed-server)
  - [Running QuickFeed Server](#running-quickfeed-server)
    - [Flags](#flags)
    - [Using GitHub Webhooks When Running Server On Localhost](#using-github-webhooks-when-running-server-on-localhost)
  - [Advanced Configuration](#advanced-configuration)
    - [The PORT Environment Variable](#the-port-environment-variable)
    - [Custom Certificate Folder](#custom-certificate-folder)
    - [Custom Certificate File Paths (Development Mode)](#custom-certificate-file-paths-development-mode)
    - [Configuring Docker](#configuring-docker)
    - [Configuring Fixed IP and Router](#configuring-fixed-ip-and-router)
  - [Troubleshooting](#troubleshooting)

## Setup

### Install Tools for Deployment

This assumes you have [homebrew](https://brew.sh/) and [make](https://www.gnu.org/software/make/) installed.

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

### Preparing the Environment for Development

For localhost development, create a minimal `.env` file:

```shell
% echo "DOMAIN=localhost" > .env
```

QuickFeed will automatically generate self-signed certificates when started with the `-dev` flag.
The certificates are stored in `~/.config/quickfeed/certs/` and will be added to your system's trust store.

To reset your development environment and start fresh, run:

```shell
% make clean-dev
```

This removes the `.env` file and `~/.config/quickfeed/` directory, then creates a new `.env` with `DOMAIN=localhost`.
See [Advanced Configuration](#advanced-configuration) for optional certificate path customization.

### Preparing the Environment for Production

QuickFeed expects a `.env` file specifying environment variables.
To get started, copy the provided `.env-template` file:

```shell
% cp .env-template .env
```

And edit (only) the `DOMAIN` variable to point to your QuickFeed server's public domain name or IP address.
The [installation process](#installation) will populate the environment variables in your `.env` file.

Below is an example production deployment on the `example.com` domain.

```shell
# The URL for the installed QuickFeed GitHub app.
QUICKFEED_APP_URL=""
# GitHub App IDs and secrets for deployment
QUICKFEED_APP_ID=""
QUICKFEED_CLIENT_ID=""
QUICKFEED_CLIENT_SECRET=""
# Secret for validating incoming webhooks from GitHub.
QUICKFEED_WEBHOOK_SECRET=""
# Secret for signing JWT tokens for user authentication.
QUICKFEED_AUTH_SECRET=""

# Quickfeed server domain or ip
DOMAIN="example.com"
# Quickfeed server port
PORT="443"

# Comma-separated list of domains to allow certificates for.
# IP addresses and "localhost" are *not* valid.
# The whitelist must also include the domain defined above.
QUICKFEED_WHITELIST="www.example.com,example.com"
```

### Installation

When starting the server for the first time, QuickFeed will automatically detect that the GitHub App has not been configured and guide you through the installation process.

```shell
% make install
% quickfeed
2025/01/23 16:45:22 Loading environment variables from /Users/meling/work/quickfeed/.env
2025/01/23 16:45:22 Important: The GitHub user that installs the QuickFeed App will become the server's admin user.
2025/01/23 16:45:22 Go to https://example.com/manifest to install the QuickFeed GitHub App.
2025/01/23 16:46:00 Successfully installed the QuickFeed GitHub App.
2025/01/23 16:46:00 Starting QuickFeed on example.com:443
```

After starting the server you should see various certificate files saved to `~/.config/quickfeed/certs/`:

```shell
% tree ~/.config/quickfeed
~/.config/quickfeed
└── certs
    ├── acme_account+key
    ├── example.com
    └── quickfeed-app-key.pem
```

In addition, your `.env` file should be populated:

```shell
% cat .env
# GitHub App IDs and secrets for deployment
QUICKFEED_APP_ID=<6 digit ID>
QUICKFEED_CLIENT_ID=Iv1.<16 chars of identifying data>
QUICKFEED_CLIENT_SECRET=<40 chars of secret data>
QUICKFEED_WEBHOOK_SECRET=<40 chars of secret data>
```

## Building QuickFeed Server

After editing files in the `public` folder, run the following command.
This should also work while the application is running.

```sh
% make ui
```

Build and install the `quickfeed` server.

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

| **Flag**        | **Description**                                                   | **Example** |
| --------------- | ----------------------------------------------------------------- | ----------- |
| `database.file` | Path to QuickFeed database                                        | `qf.db`     |
| `http.public`   | Path to content to serve                                          |             |
| `dev`           | Run development server with self-signed certificates              |             |
| `secret`        | Force regeneration of JWT signing secret (will log out all users) |             |

**About the `-secret` flag:**
QuickFeed uses the `QUICKFEED_AUTH_SECRET` environment variable to sign JWT tokens.
On first run, a secret is generated and saved to `.env`.
Use `-secret` to rotate the secret (logs out all users).
Without `-secret`, the existing secret is reused, preserving user sessions across restarts.

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

## Advanced Configuration

### The PORT Environment Variable

The `PORT` environment variable controls which port QuickFeed listens on.
It defaults to `443` (standard HTTPS port).

For development or debugging, you can use an unprivileged port (e.g., `8443`) to avoid permission issues:

```shell
% echo "PORT=8443" >> .env
```

**Running on port 443 (Linux):**
On Linux, binding to privileged ports (below 1024) requires elevated privileges.
The `make install` target automatically grants this capability on Linux.
Alternatively, run manually:

```sh
sudo setcap CAP_NET_BIND_SERVICE=+eip $(which quickfeed)
```

Note: This must be repeated after each recompile.

**VS Code debugging:**
When debugging in VS Code, you can either:

- Use an unprivileged port (recommended) and set `program` to `"${workspaceFolder}/main.go"` in your [launch profile](../.vscode/launch.json), or
- Grant the binary privileged port access and set `program` to `"${env:GOPATH}/bin/quickfeed"`.

### Custom Certificate Folder

The following environment variable can be used to specify a custom certificate folder.

| **Variable**          | **Description**                        | **Default**                  |
| --------------------- | -------------------------------------- | ---------------------------- |
| `QUICKFEED_CERT_PATH` | Directory containing certificate files | `~/.config/quickfeed/certs/` |

Example custom configuration in `.env`:

```shell
# Custom certificate paths (optional)
QUICKFEED_CERT_PATH=/etc/letsencrypt/live/example.com/
```

### Custom Certificate File Paths (Development Mode)

When running QuickFeed in development mode, self-signed certificates are automatically generated and stored in `QUICKFEED_CERT_PATH`.
However, you can override the default certificate file paths by setting the following environment variables:

| **Variable**               | **Description**             | **Default**                          |
| -------------------------- | --------------------------- | ------------------------------------ |
| `QUICKFEED_FULLCHAIN_FILE` | Full certificate chain file | `$QUICKFEED_CERT_PATH/fullchain.pem` |
| `QUICKFEED_PRIVKEY_FILE`   | Private key file            | `$QUICKFEED_CERT_PATH/privkey.pem`   |
| `QUICKFEED_CA_FILE`        | CA certificate file         | `$QUICKFEED_CERT_PATH/cert.pem`      |

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
