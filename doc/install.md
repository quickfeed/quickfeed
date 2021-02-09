# QuickFeed Server Installation and Configuration

This guide is specific for GitHub.
Technically, GitLab could be supported.
However, it would require additional development in the `scm` package, and possibly elsewhere in the code base to decouple things from GitHub specifics.

These instructions assume a Linux server with the following packages installed:

- Go, Protobuf, Docker Community Edition, npm, Envoy, NGINX

## Configuring the Server Machine

TODO(meling) Update this section

### Configuring NGINX

This [NGINX tutorial](https://www.netguru.com/codestories/nginx-tutorial-basics-concepts) is probably useful.

### Example NGINX Configuration for HTTP and gRPC traffic with Envoy

```nginx-conf
server {
        listen 443 ssl http2;
        listen [::]:443 ssl http2;

        # files you want to include, like ssl config
        include snippets/<example_file>;

        server_name <your callback url>;

        # takes care of general traffic, redirects everything to port 3333 (http.add port provided in quickfeed command)
        location / {
                proxy_pass http://127.0.0.1:3333;
                proxy_redirect off;
                proxy_set_header Host $host;
                proxy_set_header X-Real-IP $remote_addr;
                proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
                proxy_set_header X-Forwarded-Ssl on;
        }

        # takes care of gRPC traffic, redirects it to port 8080 (the port Envoy proxy listens to)
        location /AutograderService/ {
                grpc_pass 127.0.0.1:8080;
                proxy_redirect off;
                proxy_set_header Host $host;
                proxy_set_header X-Real-IP $remote_addr;
                proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
                proxy_set_header X-Forwarded-Ssl on;

                if ($request_method = 'OPTIONS') {
                  add_header 'Access-Control-Allow-Origin' '*';
                  add_header 'Access-Control-Allow-Methods' 'GET, POST, OPTIONS';
                  add_header 'Access-Control-Allow-Headers' 'DNT,X-CustomHeader,Keep-Alive,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,Content-Transfer-Encoding,Custom-Header-1,X-Accept-Content-Transfer-Encoding,X-Accept-Response-Streaming,X-User-Agent,X-Grpc-Web';
                  add_header 'Access-Control-Max-Age' 1728000;
                  add_header 'Content-Type' 'text/plain charset=UTF-8';
                  add_header 'Content-Length' 0;
                  return 204;
                }
                if ($request_method = 'POST') {
                  add_header 'Access-Control-Allow-Origin' '*';
                  add_header 'Access-Control-Allow-Methods' 'GET, POST, OPTIONS';
                  add_header 'Access-Control-Allow-Headers' 'DNT,X-CustomHeader,Keep-Alive,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,Content-Transfer-Encoding,Custom-Header-1,X-Accept-Content-Transfer-Encoding,X-Accept-Response-Streaming,X-User-Agent,X-Grpc-Web';
                  add_header 'Access-Control-Expose-Headers' 'Content-Transfer-Encoding';
                }
        }

    # ssl certificates can be obtained for free with the help of Certbot (or any other) client for Let's Encrypt
    # these lines can be added automatically by Certbot, or manually (replace <certificate folder> with the relevant folder's name)
    ssl_certificate /etc/letsencrypt/live/<certificate folder>/fullchain.pem; # managed by Certbot
    ssl_certificate_key /etc/letsencrypt/live/<certificate folder>/privkey.pem; # managed by Certbot
}
```

### SSL/TLS Certificates with Let's Encrypt/Certbot

Obtaining SSL certificates is free and easy with [Let's Encrypt](https://letsencrypt.org/).
First you must install any Let's Encrypt client, for example [Certbot](https://certbot.eff.org/about/).
To run Certbot with NGINX and QuickFeed:

```sh
sudo certbot-auto --nginx -d https://uis.itest.run/auth/github/callback
```

Replace the URL for your deployment, according to the callback URL you specified for [GitHub's OAuth2 application](./github.md).

### Adding Another NGINX Endpoint

1. Add a new server entry.
2. Check that the new configuration is correct by running

   ```sh
   sudo nginx -t
   ```

3. Restart nginx:

   ```sh
   sudo service nginx restart
   ```

4. A new endpoint requires a SSL certificate.

   - To list all active certificates:

     ```sh
     sudo certbot certificates
     ```

   - To add the new endpoint to an existing certificate, run the following command:

     ```sh
     sudo certbot-auto --nginx --cert-name <name of the existing certificate> -d <new endpoint URL>
     ```

     This will also add all the necessary SSL-related lines to the NGINX configuration file.

(TODO: Note to self: When running in the dev environment at UiS, you may wish to update the NGINX configuration on ag2.
That is, the following file needs to be updated: `/etc/nginx/sites-available/default` with a new server entry etc.)

### Configuring Docker

To ensure that Docker containers has access to networking, you may need to set up IPv4 port forwarding on your server machine:

```sh
sudo sysctl net.ipv4.ip_forward=1
sudo sysctl -p  (to confirm the change)
sudo service docker restart
```

## Installing Development Tools

The development tools are only needed for development, and can be skipped for deployment only.
To install:

```sh
make devtools
```

The `devtools` make target will download and install various Protobuf compilers, the grpcweb Protobuf compiler, webpack and the TypeScript compiler.

TODO(meling) Update the `npmtools` target to not install with global option (supposedly undesirable); need to `cd public`?

## Installing QuickFeed for Deployment

To build the QuickFeed server:

```sh
make install
```

## Running the QuickFeed Server

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

## Old Stuff

TODO(meling): Is this still a thing? If so, we should add a make target to handle this as part of deployment.

If this setup is used for production, in `public/index.html` and exchange the react-development library with production library

```html
<script src="https://unpkg.com/react@16.4.1/umd/react.development.js"></script>
<script src="https://unpkg.com/react-dom@16.4.1/umd/react-dom.development.js"></script>
```

to

```html
<script src="https://unpkg.com/react@16.4.1/umd/react.production.min.js"></script>
<script src="https://unpkg.com/react-dom@16.4.1/umd/react-dom.production.min.js"></script>
```
