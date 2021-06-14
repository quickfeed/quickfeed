# Deployment with Domain Name and Let's Encrypt Certificates

The following instructions assume a fixed IP and domain name for the server to be `cyclone.meling.me`.
Replace the relevant IP address and domain name with your own.

## Configure Fixed IP and Router

In your domain name provider, configure your IP and domain name; for instance:

```text
Type         Host          Value                TTL
A Record     cyclone       92.221.105.172       5 min
```

Set up port forwarding on your router.
External ports 80/443/8080 maps to internal ports 80/443/8080 for TCP.

## Install Tools for Deployment

This assumes you have homebrew installed.
For systems without homebrew, the make target should list well-known packages available on most Unix distributions.

```sh
% make brew
```

## Install Tools for Development

The development tools are only needed for development, and can be skipped for deployment only.
To install:

```sh
% make devtools
```

The `devtools` make target will download and install various Protobuf compiler plugins and the grpcweb Protobuf compiler.

## Generate Certbot Private Key and Certificate

```terminal
% sudo certbot certonly --standalone
Saving debug log to /var/log/letsencrypt/letsencrypt.log
Plugins selected: Authenticator standalone, Installer None
Please enter in your domain name(s) (comma and/or space separated)  (Enter 'c'
to cancel): cyclone.meling.me
Requesting a certificate for cyclone.meling.me
Performing the following challenges:
http-01 challenge for cyclone.meling.me
Waiting for verification...
Cleaning up challenges

IMPORTANT NOTES:
 - Congratulations! Your certificate and chain have been saved at:
   /etc/letsencrypt/live/cyclone.meling.me/fullchain.pem
   Your key file has been saved at:
   /etc/letsencrypt/live/cyclone.meling.me/privkey.pem
   Your certificate will expire on 2021-09-11. To obtain a new or
   tweaked version of this certificate in the future, simply run
   certbot again. To non-interactively renew *all* of your
   certificates, run "certbot renew"
```

## Configuring Docker

To ensure that Docker containers has access to networking, you may need to set up IPv4 port forwarding on your server machine:

```sh
sudo sysctl net.ipv4.ip_forward=1
sudo sysctl -p  (to confirm the change)
sudo service docker restart
```

## Run with Envoy

```sh
% sudo envoy -c envoy-cyclone.yaml
```

## Configure GitHub OAuth Application for QuickFeed

To deploy QuickFeed, you need to configure a GitHub account for communicating with QuickFeed.
See the instructions for configuring a [GitHub OAuth2 application](./github.md).

For this tutorial, we use the following:

```text
Homepage URL: https://cyclone.meling.me
Authorization callback URL: https://cyclone.meling.me/auth/github/callback
```

## Build and Run QuickFeed Server

The `webpack` command must be executed after editing files in the `public` folder.
This works while the application is running.

```bash
cd public
webpack
```

Build and run the `quickfeed` server; here we use all default values:

```bash
% cd $QUICKFEED
% go install
% source quickfeed-env.sh
% quickfeed -service.url cyclone.meling.me
```

## Running the QuickFeed Server

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

### Flags

| **Flag**        | **Description**                        | **Example**     |
|-----------------|----------------------------------------|-----------------|
| `service.url`   | Base DNS name for QuickFeed deployment | `uis.itest.run` |
| `database.file` | Path to QuickFeed database             | `qf.db`         |
| `grpc.addr`     | Listener address for gRPC service      | `:9090`         |
| `http.addr`     | Listener address for HTTP service      | `:3005`         |
| `http.public`   | Path to service content                | `public`        |
| `script.path`   | Path to continuous integration scripts | `ci/scripts`    |

### Custom Docker Image for a Course

QuickFeed will pull publicly available docker images from Docker Hub on demand.
However, you may create custom docker images locally on your QuickFeed server machine, and use these locally only.
That is, you don't need to upload your custom image to Docker Hub or elsewhere.

To prepare a new custom Docker image for a course, prepare the relevant `Dockerfile` and build it.
The `quickfeed-go` make target gives an example:

```sh
make quickfeed-go
```
