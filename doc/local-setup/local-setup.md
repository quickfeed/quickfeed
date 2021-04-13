# Installing Quickfeed on localhost

## Table of Contents
[Ubuntu installation](#ubuntu-installation)
1. [Installing Go](#installing-go)
3. [Installing Envoy](#installing-envoy)
4. [Installing Nginx](#installing-nginx)
5. [Installing Node.js, npm and npm(webpack)](#installing-nodejs-npm-and-npmwebpack)

[Configure Quickfeed for localhost](#configure-quickfeed-for-localhost)
1. [Quickfeed](#quickfeed)
2. [Envoy](#envoy)
3. [Self-signed SSL certificates](#self-signed-ssl-certificates)
4. [Nginx](#nginx)
5. [Setting up a Github login](#setting-up-a-github-login)

[Starting the application](#starting-the-application)
1. [Nginx & Envoy](#nginx--envoy)
2. [Go dependencies](#go-dependencies)
3. [Webpack](#webpack)
4. [Github key](#github-key)
5. [Quickfeed](#quickfeed)

## Ubuntu installation

These instructions should also work on Windows with WSL. First, make sure that you have installed Go, npm, npm(webpack), Envoy and Nginx.

### Installing Go

Start by adding `golang-backports` by using these commands in terminal.

```bash
sudo add-apt-repository ppa:longsleep/golang-backports
sudo apt update
sudo apt install golang-go
```

Make sure to add PATH to `~/.bashrc`. Editors like VIM works great for this. 

```bash
export GOPATH="$HOME/go"
PATH="$GOPATH/bin:$PATH"
```

Restart the terminal or type in this command:

```bash
source ~./bashrc
```

Type `go version` to verify the installation.

### Installing Envoy

Follow the installation guide on [Envoy](https://www.getenvoy.io/install/envoy/ubuntu/)

### Installing Nginx

Nginx can be installed in the terminal with these commands.

```bash
sudo apt-get update
sudo apt-get install nginx
```
### Installing Node.js, npm and npm(webpack)

Node.js, npm can be installed in the terminal with these commands.

```bash
sudo apt update
sudo apt install nodejs npm
```

Verify the installation:

```bash
nodejs --version
```

Install webpack in the `/quickfeed` folder with this command:

```bash
npm install --save-dev webpack
```

## Configure Quickfeed for localhost

### Quickfeed

Make sure to comment these lines in `main.go` (line number might differ, but should be nearby).

```bash
14. "github.com/autograde/quickfeed/envoy"
109. go envoy.StartEnvoy(logger)
```
Also ***don't*** use the `make local` because it will trigger _Same-origin policy_ 

### Envoy

Take a backup of `/envoy/envoy.yaml` and change the file to this.

```yaml
admin:
  access_log_path: /dev/stdout
  address:
    socket_address: { address: 0.0.0.0, port_value: 9901 }

static_resources:
  listeners:
  - name: listener_1
    address:
      socket_address: { address: 0.0.0.0, port_value: 8080 }
    filter_chains:
    - filters:
      - name: envoy.filters.network.http_connection_manager
        typed_config:
          "@type": type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
          codec_type: auto
          stat_prefix: ingress_http
          route_config:
            name: local_route
            virtual_hosts:
            - name: local_service
              domains: ["*"]
              routes:
              - match: { prefix: "/" }
                route:
                  cluster: echo_service
                  max_stream_duration:
                    grpc_timeout_header_max: 0s
              cors:
                allow_origin_string_match:
                - prefix: "*"
                allow_methods: GET, PUT, DELETE, POST, OPTIONS
                allow_headers: keep-alive,user-agent,cache-control,content-type,content-transfer-encoding,custom-header-1,x-accept-content-transfer-encoding,x-accept-response-streaming,x-user-agent,x-grpc-web,grpc-timeout,user
                max_age: "1728000"
                expose_headers: custom-header-1,grpc-status,grpc-message
          http_filters:
          - name: envoy.filters.http.grpc_web
          - name: envoy.filters.http.cors
          - name: envoy.filters.http.router
          http_protocol_options: { accept_http_10: true}
  clusters:
  - name: echo_service
    connect_timeout: 15s
    type: logical_dns
    http2_protocol_options: {}
    lb_policy: round_robin
    load_assignment:
      cluster_name: cluster_0
      endpoints:
        - lb_endpoints:
            - endpoint:
                address:
                  socket_address:
                    address: localhost
                    port_value: 9090
```

## Nginx And Certificates
*Alternatively you can use [this script](https://github.com/plakolaki/bachelor-redesign-quickfeed/blob/main/Quickfeed_setup/create-cert%2Bnginx-conf.sh) instead that configures certificates and nginx config for you. If you did use it you can skip to [GitHub Login](#setting-up-a-github-login)*

### Self-signed SSL certificates

#### WARNING: Never use self-signed SSL certificates for anything other than localhost applications. If the application will be used online, a certificate can be created by `Certbot`or other providers.

These instructions install the certificates inside the folder `/etc/nginx/sites-available/`. If one wants the certificates in a different folder, make sure to change the Nginx `default` file accordingly.

Follow these commands for creating the certificates:

```bash
cd /etc/nginx/sites-available/
sudo openssl req -x509 -nodes -new -sha256 -days 1024 -newkey rsa:2048 -keyout RootCA.key -out RootCA.pem -subj "/C=US/CN=Example-Root-CA"
sudo openssl x509 -outform pem -in RootCA.pem -out RootCA.crt
```
You need certificates because OAuth won't work over regular HTTP

### Nginx

Edit `/etc/nginx/sites-available/default` by adding this to the bottom, make sure that everything else is commented out.

If VIM is installed the file can be accessed like this:
```bash
sudo vim /etc/nginx/sites-available/default
```

```bash
server {
        listen 443 ssl http2;
        listen [::]:443 ssl http2;

        server_name https://127.0.0.1/auth/github/callback;
        location / {
                proxy_pass http://127.0.0.1:8081;
                proxy_redirect off;
                proxy_set_header Host $host;
                proxy_set_header X-Real-IP $remote_addr;
                proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
                proxy_set_header X-Forwarded-Ssl on;
        }

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
}

ssl_certificate /etc/nginx/sites-available/RootCA.crt;
ssl_certificate_key /etc/nginx/sites-available/RootCA.key;
ssl_trusted_certificate /etc/nginx/sites-available/RootCA.pem;
```

To run Nginx:
```bash
sudo nginx
```

If any changes are made, reload the file with this:
```bash
sudo nginx -s reload
```

### Setting up a Github login

Create a new Github app by following the instructions on this [page](https://docs.github.com/en/developers/apps/creating-a-github-app).

Ensure that homepage URL is `127.0.0.1` and callback URL is `127.0.0.1/auth/github/callback` like in the photo.
![Adding URL](./figures/github_app_url.png)

Create a secret key by pressing `Generate new client secret` outline in the picture. Make sure to copy the client key before leaving the page.
![Secret key](./figures/github_app_client_secret.png)

Create a new file in `/quickfeed` named `ag-env.sh` and insert these line, changing `client_key` and `client_secret` with the key and secret generated: 

```sh
export GITHUB_KEY='client_key'
export GITHUB_SECRET='client_secret'
```

Allow oauth access by following the instructions on this [page](https://docs.github.com/en/github/setting-up-and-managing-organizations-and-teams/approving-oauth-apps-for-your-organization).

## Starting the application

### Nginx & Envoy
Make sure Nginx is running, start Envoy with this command inside `/quickfeed/envoy`:
```bash
envoy -c envoy.yaml
```

### Go dependencies
While Nginx and Envoy is running, install Go dependencies with the commands instide `/quickfeed`:

```bash
go mod download
go install
```

### Webpack
Initialise the webpack by running this command inside `/quickfeed/public`:

```bash
webpack
```

This command needs to be run when editing files inside `quickfeed/public`. This works while the application is running.

### Github key
Mount ag-env.sh with this command:

```bash
source ag-env.sh
```

The file must be mounted every time the computer has been restarted.

### Quickfeed
Start quickfeed with this command:

```bash
quickfeed -http.addr ":8081" -service.url "127.0.0.1"
```

You can now visit Quickfeed on the ip-address 127.0.0.1.




### Running Quickfeed using only Envoy

This can be done using the configuration found in [this config file](/local-setup/envoy.yaml).

Replace the content of lines preceeded by a comment in the envoy.yaml file with the configuration your setup is using.

Additionally, you must modify this line in GRPCManager.ts (line 60) to include the port you configure Envoy to listen for GRPC traffic (default :8080).

```ts
    constructor() {
        this.agService = new AutograderServiceClient("https://" + window.location.hostname, null, null);
    }
```

Example:

```ts
    constructor() {
        this.agService = new AutograderServiceClient("https://" + window.location.hostname + ":8080", null, null);
    }
```

