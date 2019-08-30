# Autogradee developer manual

## Technology stack

- [Go](https://golang.org/doc/code.html)
- [TypeScript](https://www.typescriptlang.org/)
- [gRPC](https://grpc.io/)
- [Protocol Buffers](https://developers.google.com/protocol-buffers/docs/proto3)
- [gRPC-Web](https://github.com/grpc/grpc-web)
- [Envoy](https://www.envoyproxy.io/)
- [NGINX](https://www.nginx.com/resources/wiki/)

### Package management // Download and Install

- [installation guide](./Installation.md)  (TODO - update with go modules information)
- GOPATH environmental variable (go get)
- go modules (git clone + go install): 
    [go mod](https://blog.golang.org/using-go-modules)

## GitHub integration: OAuth2 application setup

[GitHub setup instructions](./GithubSetup.MD)

[Setting up the course organization](./Teacher.MD)

## Starting up the server

The command to start up the server is of the following form:
`aguis -service.url <url> -database.file <path to database> -http.addr <HTTP listener address> -http.public <path to static files>`

### Flags

- `service.url` - the URL you have set up as callback URL for your Autograder GitHub OAuth2 application. Defaul value: localhost
- `database.file` - the path to application's database. Default: `/tmp/ag.db` - will be temporary, i.e., removed after reboot
- `http.addr` - the port number for HTTP listener. Default value: :8081
- `http.public` - the path to the static files to serve. Default value: `./public`

## SCM tool

SCM tool can be used for working with github organization from command line. This way you can remove all repositories and teams from your test organization withiout going to github page and removing them manually.

### Prerequisites

SCM tool must be compiled before it can be used. To compile, run `make scm` command or go to `cmd/csm` folder and run `go compile`.
After that `scm` command will become active.

To use SCM tool, you have to create a personal github access token. This is done on GitHub's web page:

1. Navigate to Settings (in the personal menu accessible from your avatar picture)
2. Select _Developer settings_ from the menu on the left.
3. Select _Personal access tokens_ and on the next page,
4. Select _Generate new token_. Name the token, e.g. `Autograder Test Token`.
5. Select _Scopes_ as needed; currently I have enabled `admin:org, admin:org_hook, admin:repo_hook, delete_repo, repo, user`, but you may be able to get away with fewer access scopes. It depends on your needs.
6. Copy the generated token string to the `GITHUB_ACCESS_TOKEN` environment variable. You may wish to add this token to your local `ag-setup.sh` script file.

```sh
  export GITHUB_ACCESS_TOKEN=<your token>
```
You must also be owner of the GitHub organization to be able to access its repositories and teams with SCM tool.

### Example usage:

```
scm --provider github get repo -all -namespace autograder-test
```
  will print out information about all repos existing for the **autograder-test** organization.

```
scm delete team -all -namespace autograder-test
```
  will remove all teams existing for the **autograder-test** organization

Other examples and instructions are provided in comments in the `cmd/ci/main.go` file.

## Makefile

Makefile allows to simplify running different tasks, i.e. compiling, updating and starting the server.

The list at the top of `Makefile` compises variables, that can be altered according to situation (especially port numbers, organization name, and URIs).

### Compiling tasks

Use `make proto` if you want to introduce any changes to ag.proto file. It will compile all the frontend and backend code and make all the necesary changes to compiled files. You will have to recompile the server and the client code after that. Use 'make install' and 'make ui' for these tasks.

### Proxy

Use `make envoy-build` to rebuild the envoy container, and `make envoy-run` to start it manually.
Use `make envoy-purge` to clean up all envoy containers and images before rebuilding.

Use `make nginx` to reload nginx. It will not reload if configuration file has invalid syntax.

### Testing

Use `make test` to run all the tests in `web` and `database` packages.

### Utility

`make local` and `make remote` will switch where and how the gRPC client is being run, then will recompile the frontend. Use `make local` when running server locally with port forwarding, otherwise use `make remote`.

**Warning:** never push code with local gRPC client settings to the `aguis` repository, it will cause the server to stop responding to client requests. If this happens, just run `make remote` on the server location.


## Server architecture

### Default setup

By default, the gRPC server will be started at port **:9090**. A docker container with Envoy proxy will listen on port **:8080** and redirect gRPC traffic to the server. Webserver is running on one of internal ports and NGINX erving the static content is set up to redirect HTTP traffic to that port, and all gRPC traffic to the port **:8080** where Envoy proxy is listening.
NGINX and Envoy take care of all the relevant heades for gRPC traffic. 

##  Envoy 
Envoy proxy allows making gRPC call from a browser application.

### Basic configuration
[Default configuration from grpc-web repository](https://github.com/grpc/grpc-web/blob/master/net/grpc/gateway/examples/echo/envoy.yaml)
The main difference in [our configuration](https://github.com/autograde/aguis/blob/grpc-web-merge/envoy/envoy.yaml) is `http_protocol_options: { accept_http_10: true }` line inside HTTP filter options and an additional header name.

## NGINX

[NGINX tutoial](https://www.netguru.com/codestories/nginx-tutorial-basics-concepts)

### Example setup for HTTP and gRPC traffic with Envoy

```
server {
        listen 443 ssl http2;
        listen [::]:443 ssl http2;

        # files you want to include, like ssl config
        include snippets/<example_file>;

        server_name <your callback url>;

        # takes care og general traffic, redirects everythin to port 3333 (http.add port from aguis command)
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

    # ssl certificates can be obtained for free with the help of Certbot (or any other) client for Letsencrypt
    # these lines can be added automatically by Certbot, or manually (replace <certificate folder> with the relevant folder's name)
    ssl_certificate /etc/letsencrypt/live/<cerificate folder>/fullchain.pem; # managed by Certbot
    ssl_certificate_key /etc/letsencrypt/live/<certificate folder>/privkey.pem; # managed by Certbot
}
```

### SSL/TLS certificates with Letsencrypt/Certbot

Obtaining SSL certificates is free and easy with [Letsencrypt](https://letsencrypt.org/). 
First you must install any Letsencrypt client, for example [Certbot](https://certbot.eff.org/about/).
Then run `sudo certbot---nginx -d <URL you wish to protect>`. When working with Autograder, this should be the same URL you provide to GitHub OAuth2 application as callback URL.

## Errors and logging

### Backend

Errors are being logged at `Autograder Service` level. All other methods called from there (including database and scm methods) will just wrap and return all error messages directly. Introduce logging on layerd deeper than `Autograder Service` only if necessary. 

Errors returned to user interface must be few and informative, yet should not provide too many information about server routines. User must only be informed about details he or she can do something about.

### Frontend

When receiving responce from the server, responce status code is checked on the frontend. Any message with code different from 0 (0 = OK status code for gRPC) will be logged to console. Relevant error messages will be displayed to user on course and group creation, and user and group enrollment updates.

[gRPC status codes](https://github.com/grpc/grpc/blob/master/doc/statuscodes.md)

## GitHub API

For GitHub integration we are using [Go implementation](https://github.com/google/go-github/tree/master/github) of [GitHub API](https://developer.github.com/v3/) 

### Webhooks

- GitHub [Webhooks API](https://developer.github.com/webhooks/) is used for building and testing of student solutions.
- Webhook is set automatically on course creation. It will react to every push event to any of course organization's repositories. 
- Depending on the repository, assignment information will be updated in the Autograder's databae, or a docker container with a student solution will be built
- `name` argument for any GitHub webhook is always "web"
- webhook will be using the same callback URL you have provided to the Autograder OAuth2 application and in the server startup command

### User roles/access levels for organization / team / repository

- GitHub API name for organization owner is `admin`
- Repository access levels for any organization member in GitHub API calls are: `read`/`write`/`admin`/`none`
- Individual repository permission levels in GitHub API are: `pull`/`push`/`admin`

### Slugs

When retrieving team, organization or repository by name, GitHub expects a slugified string instead of a full name as displayed on the organization page. 
For example, organization with a name like `Ag Test Org` will have slugified name `ag-test-org`.

[URL slugs explained](http://patterns.dataincubator.org/book/url-slug.html)

### Repositories

- `owner` field for any organization repository is a slugified name for that organization
- access policy:
  - on course creation - default repository access across the whole organization is set to `none`, which means only the organization owners can see any private repository on that organization
  - when students enroll, they receive read/pull access to `assignments` repository and write/push access to a personal student repository

### Teams

Student groups will have GitHub teams with the same name created in the course organization.
Group records in the Autograder's database will have references to the corresponding GitHub team ID's.


## Simple testing of UI with dummy data

Since this is a single page application, then you would only be able to navigate to the default index page, every other navigation is handle by javascript. The way this works at the server, is that every request to /app/ returns the index.html page.
Now to get around the problem, we have to use the built in navigation manager, navMan. To do this, go to the main page of the application and open developer tools in the browser. In the command line type in the following command debugData.navMan.navigateTo("/app/admin/courses/new"). You could replace the url with any other url as you like.











