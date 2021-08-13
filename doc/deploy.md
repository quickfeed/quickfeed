# Quickfeed Deployments

## Table of Contents
- [Preparing the Environment](#preparing-the-environment)
  - [Configuring Docker](#configuring-docker)
  - [Setup Environment Variables](#setup-environment-variables)
  - [Generate Envoy Configuration File](#generate-envoy-configuration-file)
  - [Configure GitHub OAuth Application for QuickFeed](#configure-github-oauth-application-for-quickfeed)
- [Docker Deployment](#docker-deployment)
- [Bare Metal Deployment](#bare-metal-deployment)
  - [Deployment with Domain Name and Let's Encrypt Certificates](#deployment-with-domain-name-and-lets-encrypt-certificates)
  - [Configure Fixed IP and Router](#configure-fixed-ip-and-router)
  - [Install Tools for Deployment](#install-tools-for-deployment)
  - [Install Tools for Development](#install-tools-for-development)
  - [Generate Certbot Private Key and Certificate](#generate-certbot-private-key-and-certificate)
  - [Run Envoy](#run-envoy)
  - [Build and Run QuickFeed Server](#build-and-run-quickfeed-server)
    - [Troubleshooting](#troubleshooting)
  - [Running the QuickFeed Server Details](#running-the-quickfeed-server-details)
    - [Flags](#flags)
    - [Custom Docker Image for a Course](#custom-docker-image-for-a-course)

## Preparing the Environment

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
# Edit .env according to your domain and exposed ports
```

### Generate Envoy Configuration File

The default envoy configuration for testing can be generated using the existent rules in the Makefile:
```sh
% make envoy-config
```

This configuration does not use TLS. To enable TLS but generate certificates for testing purposes, you can run:

```sh
% go run ./envoy/envoy.go --genconfig --withTLS
```

If you already have certificates that you would like to use you can specify them during the creation of the envoy configuration, running the command below.

```sh
% go run ./envoy/envoy.go --genconfig --withTLS --certFile="fullchain.pem" --keyFile="key.pem"
```

The envoy config will be generate at `$QUICKFEED/envoy/envoy.yaml`.

_The script sets the certificate and key at the following path: `/etc/letsencrypt/live/YOUR_DOMAIN_NAME/(CERTIFICATE | KEY).pem`. Please note that, when running envoy in your host machine, you need to ensure that certificates and necessary keys are stored in the same path specified in the envoy config._


### Configure GitHub OAuth Application for QuickFeed

To deploy QuickFeed, you need to configure a GitHub account for communicating with QuickFeed.
See the instructions for configuring a [GitHub OAuth2 application](./github.md).

For this tutorial, we use the following:

```text
Homepage URL: https://YOUR_DOMAIN_NAME
Authorization callback URL: https://YOUR_DOMAIN_NAME/auth/github/callback
```

## Docker Deployment

Quickfeed and Envoy can be installed and run on containers using the [docker-compose](../docker-compose.yml) configuration.
First, ensure that [docker-compose](https://docs.docker.com/compose/) is installed in your system.
Then, to build and run the containers, run:

```sh
% docker-compose up --build
```

If you would like to run envoy in a container but quickfeed locally in the host machine, please run envoy as described in the section [Run Envoy](#run-envoy) sub-section 3. Then run quickfeed as described in section [Build Quickfeed](#build-and-run-quickfeed-server).


## Bare Metal Deployment

### Deployment with Domain Name and Let's Encrypt Certificates

The following instructions assume a fixed IP and domain name for the server to be `YOUR_DOMAIN_NAME`.
Replace the relevant IP address and domain name with your own.

### Configure Fixed IP and Router

In your domain name provider, configure your IP and domain name; for instance:

```text
Type         Host          Value                TTL
A Record     cyclone       92.221.105.172       5 min
```

Set up port forwarding on your router.
External ports 80/443/8080 maps to internal ports 80/443/8080 for TCP.

### Install Tools for Deployment

This assumes you have homebrew installed.
For systems without homebrew, the make target should list well-known packages available on most Unix distributions.

```sh
% make brew
```

```sh
% brew install tmux overmind
% Procfile
```

### Install Tools for Development

The development tools are only needed for development, and can be skipped for deployment only.
To install:

```sh
% make devtools
```

The `devtools` make target will download and install various Protobuf compiler plugins and the grpcweb Protobuf compiler.

### Generate Certbot Private Key and Certificate

```terminal
% sudo certbot certonly --standalone
Saving debug log to /var/log/letsencrypt/letsencrypt.log
Plugins selected: Authenticator standalone, Installer None
Please enter in your domain name(s) (comma and/or space separated)  (Enter 'c'
to cancel): YOUR_DOMAIN_NAME
Requesting a certificate for YOUR_DOMAIN_NAME
Performing the following challenges:
http-01 challenge for YOUR_DOMAIN_NAME
Waiting for verification...
Cleaning up challenges

IMPORTANT NOTES:
 - Congratulations! Your certificate and chain have been saved at:
   /etc/letsencrypt/live/YOUR_DOMAIN_NAME/fullchain.pem
   Your key file has been saved at:
   /etc/letsencrypt/live/YOUR_DOMAIN_NAME/privkey.pem
   Your certificate will expire on 2021-09-11. To obtain a new or
   tweaked version of this certificate in the future, simply run
   certbot again. To non-interactively renew *all* of your
   certificates, run "certbot renew"
```

```terminal
% sudo certbot certonly --standalone
Saving debug log to /var/log/letsencrypt/letsencrypt.log
Please enter the domain name(s) you would like on your certificate (comma and/or
space separated) (Enter 'c' to cancel): ag2.ux.uis.no
Renewing an existing certificate for ag2.ux.uis.no

Successfully received certificate.
Certificate is saved at: /etc/letsencrypt/live/ag2.ux.uis.no/fullchain.pem
Key is saved at:         /etc/letsencrypt/live/ag2.ux.uis.no/privkey.pem
This certificate expires on 2021-10-28.
These files will be updated when the certificate renews.
Certbot has set up a scheduled task to automatically renew this certificate in the background.

- - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
If you like Certbot, please consider supporting our work by:
 * Donating to ISRG / Let's Encrypt:   https://letsencrypt.org/donate
 * Donating to EFF:                    https://eff.org/donate-le
- - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
```

### Run Envoy

Please choose one of the options below to run envoy in your system:

1. Running with Locally Installed Envoy (macOS homebrew)

```sh
% sudo envoy -c $QUICKFEED/envoy/envoy.yaml &
```

With additional logging:

```sh
% sudo envoy -c $QUICKFEED/envoy/envoy.yaml --log-path envoy.log --enable-fine-grain-logging -l debug &
```

2. Running with func-e (Linux)

Install the `func-e` command in `/usr/local/bin` with:

```sh
% curl https://func-e.io/install.sh | sudo bash -s -- -b /usr/local/bin
```

Alternatively, install via linuxbrew instead:

```sh
% brew install func-e
% sudo ln -s /home/linuxbrew/.linuxbrew/bin/func-e /usr/local/bin
```

Run with:

```sh
% sudo func-e run -c $QUICKFEED/envoy/envoy.yaml &
```

3. Running Envoy using docker-compose

If you want to run envoy using the existent docker-compose configuration you need to copy your certificates to `$QUICKFEED/certs` and run:

```sh
% docker-compose up --build --remove-orphans envoy
```

Or use the makefile:
```sh
% make envoy-build
% make envoy-run
```

### Build and Run QuickFeed Server

After editing files in the `public` folder, run the following command.
This should also work while the application is running.

```bash
% make ui
```

Build and run the `quickfeed` server; here we use all default values:

```bash
% go install
% quickfeed -service.url $DOMAIN  &> quickfeed.log &
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

If you already have NGINX configured it may conflict with the default envoy configuration.
To disable NGINX run:

```sh
# Temporarily stop running NGINX
% sudo systemctl stop nginx
# Permanently disable NGINX
% sudo systemctl disable nginx
```

### Running the QuickFeed Server Details

The following provides additional details for running QuickFeed.
Before running the QuickFeed server, you need to configure [GitHub](./github.md).

The command line arguments for the QuickFeed server looks roughly like this:

```sh
quickfeed -service.url <DNS name of deployed service> -database.file <path to database> -http.addr <HTTP listener address>
```

To view the full usage details:

```sh
quickfeed -help
```

Here is an example with all default values:

```sh
quickfeed -service.url uis.itest.run &> quickfeed.log &
```

*As a bootstrap mechanism, the first user to sign in, automatically becomes administrator for the system.*

#### Flags

| **Flag**        | **Description**                        | **Example**     |
|-----------------|----------------------------------------|-----------------|
| `service.url`   | Base DNS name for QuickFeed deployment | `uis.itest.run` |
| `database.file` | Path to QuickFeed database             | `qf.db`         |
| `grpc.addr`     | Listener address for gRPC service      | `:9090`         |
| `http.addr`     | Listener address for HTTP service      | `:8081`         |
| `http.public`   | Path to service content                | `public`        |

#### Custom Docker Image for a Course

QuickFeed will pull publicly available docker images from Docker Hub on demand.
However, you may create custom docker images locally on your QuickFeed server machine, and use these locally only.
That is, you don't need to upload your custom image to Docker Hub or elsewhere.

To prepare a new custom Docker image for a course, prepare the relevant `Dockerfile` and build it.
The `quickfeed-go` make target gives an example:

```sh
make quickfeed-go
```
