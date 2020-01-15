# Autograder developer manual

## Technology stack

- [Go](https://golang.org/doc/code.html)
- [TypeScript](https://www.typescriptlang.org/)
- [gRPC](https://grpc.io/)
- [Protocol Buffers](https://developers.google.com/protocol-buffers/docs/proto3)
- [gRPC-Web](https://github.com/grpc/grpc-web)
- [Envoy](https://www.envoyproxy.io/)
- [NGINX](https://www.nginx.com/resources/wiki/)

## Download and Install

[Installation guide](./Installation.md)  

[Working with go modules](https://blog.golang.org/using-go-modules)

## GitHub integration

[GitHub OAuth2 application setup instructions](./GithubSetup.MD)

[Setting up course organization](./Teacher.MD)

## Starting up the server

The command to start up the server:
`aguis -service.url <url> -database.file <path to database> -http.addr <HTTP listener address> -http.public <path to static files>`

### Flags

- `service.url` - URL you have set up as callback URL for your Autograder GitHub OAuth2 application. Defaul value: `localhost`
- `database.file` - path to application's database. Default: `/tmp/ag.db` - will be temporary, i.e., removed after reboot
- `http.addr` - port number for HTTP listener. Default value: `:8081`
- `http.public` - path to the static files to serve. Default value: `./public`

## SCM tool

SCM tool can be used for working with github organization from command line. This way you can list or remove all repositories and teams for your test organization withiout going to GitHub page and doing everything them manually.

### Prerequisites

SCM tool must be compiled before it can be used. To compile, run `make scm` command or go to `cmd/csm` folder and run `go install`.
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
You must also be an owner of the GitHub organization to be able to access its repositories and teams with SCM tool.

### Example usage:

```
scm --provider github get repo -all -namespace autograder-test
```
  will print out information about all repositories existing for the **autograder-test** organization.

```
scm delete team -all -namespace autograder-test
```
  will remove all teams existing for the **autograder-test** organization

Other examples and instructions are provided in comments in the `cmd/ci/main.go` file.

## Makefile

Makefile allows to simplify running different tasks, such as compiling, updating and starting up the server.

The list at the top of `Makefile` compises variables, that can be altered according to situation (especially port numbers, organization name, and URIs).

### Compiling tasks

Use `make proto` if you want to introduce any changes to ag.proto file. It will recompile all the frontend and backend code and make all the necessary changes to compiled files.
'make install' will recompile Go code for the Autograder server, and 'make ui' will start build `bundle.js` file for Autograder browser-based client.

### Proxy

Use `make envoy-build` to rebuild the Envoy docker container, and `make envoy-run` to start it manually.
Use `make envoy-purge` to clean up all envoy containers and images before rebuilding.

Use `make nginx` to reload nginx. It will not reload if configuration file has invalid syntax.

### Testing

Use `make test` to run all the tests in `web` and `database` packages.

### Utility

`make local` and `make remote` will switch where and how the gRPC client is being run, then will recompile the frontend. Use `make local` when running server on localhost with port forwarding, otherwise use `make remote`.

**Warning:** never push code with local gRPC client settings to the `aguis` repository, it will cause the server to stop responding to client requests. If this happens, just run `make remote` on the server location.


## Server architecture

### Default setup

By default, the gRPC server will be started at port **:9090**. A docker container with Envoy proxy will listen on port **:8080** and redirect gRPC traffic to the gRPC server. 
Webserver is running on one of internal ports, and NGINX, serving the static content, is set up to redirect HTTP traffic to that port, and all gRPC traffic to the port **:8080** (same port Envoy proxy is listening on).
NGINX and Envoy take care of all the relevant heades for gRPC traffic. 

##  Envoy 
Envoy proxy allows making gRPC calls from a browser application.

### Basic configuration
[Default configuration from grpc-web repository](https://github.com/grpc/grpc-web/blob/master/net/grpc/gateway/examples/echo/envoy.yaml)
The main difference in [our configuration](https://github.com/autograde/aguis/blob/grpc-web-merge/envoy/envoy.yaml) is `http_protocol_options: { accept_http_10: true }` line inside HTTP filters list, and an additional header name.

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

        # takes care of general traffic, redirects everything to port 3333 (http.add port provided in aguis command)
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

Application errors can be classified into several groups and handled in different ways.

1. Database errors
Return generic "not found/failed" error message to user, log the original error.

2. SCM errors
Some of these can only be fixed by the user who is calling the method by interacting with UI elements (usually course teacher). 
Example: if a github organization cannot be found, one of the possible issues causing this behavior is disabled third party access. As a result, the requested organization cannot be seen by the application. If a GitHub repo or team cannot be found, they could have been manually deleted from GitHub. Only the current user can remedy the situation, and it is most useful to inform them about the issue in detail and offer a solution.
Another example: sometimes GitHub interactions take too long and the request context can be canceled by GitHub. This is not an application failure, and neither developer team, nor user can actively do anything to correct it. But it is still useful to inform the user that the action must be repeated at later time to succeed.
Return a custom error with method name and original error for logging, and a custom message string to be displayed to user.

3. Access control errors
Return generic "access denied" error message to user.

4. API errors (invalid requests)
Return generic "malformed request" message to user.

5. GitHub API errors (request struct has missing/malformed fields)
Return a custom error with detailed information for logging and generic "action failed" message to user.



[GRPC status codes](https://github.com/grpc/grpc/blob/master/doc/statuscodes.md) are used to allow the client to check whether the error message should be displayed for user, or just logged for developers.

When the client is supposed to show error message to the user, error from the server will have status 9 (gRPC status "failed precondition")

### Backend

Errors are being logged at `Autograder Service` level. All other methods called from there (including database and scm methods) will just wrap and return all error messages directly. Introduce logging on layers deeper than `Autograder Service` only if necessary. 

Errors returned to user interface must be few and informative, yet should not provide too many information about server routines. User must only be informed about details he or she can do something about.

### Frontend

When receiving response from the server, response status code is checked on the frontend. Any message with code different from 0 (0 is gRPC status code `OK`) will be logged to console. Relevant error messages will be displayed to user on course and group creation, and user and group enrollment updates.

[gRPC status codes](https://github.com/grpc/grpc/blob/master/doc/statuscodes.md)

## GitHub API

For GitHub integration we are using [Go implementation](https://github.com/google/go-github/tree/master/github) of [GitHub API](https://developer.github.com/v3/) 

### Webhooks

- GitHub [Webhooks API](https://developer.github.com/webhooks/) is used for building and testing of code submitted by students.
- webhook is created automatically on course creation. It will react to every push event to any of course organization's repositories. 
- depending on the repository the push event is coming from, assignment information will be updated in the Autograder's database, or a docker container with a student solution code will be built
- `name` field for any GitHub webhook is always "web"
- webhook will be using the same callback URL you have provided to the Autograder OAuth2 application and in the server startup command

### User roles/access levels for organization / team / repository

- GitHub API name for organization owner is `admin`
- Repository access levels for any organization member in GitHub API calls are: `read`/`write`/`admin`/`none`
- Individual repository permission levels in GitHub API are: `pull`/`push`/`admin`

### Slugs

When retrieving team, organization or repository by name, GitHub expects a slugified string instead of a full name as displayed on the organization page. 
For example, organization with a name like `Autograder Test Org` will have slugified name `autograder-test-org`.

[URL slugs explained](http://patterns.dataincubator.org/book/url-slug.html)

### Repositories

- `owner` field for any organization repository is a slugified name for that organization
- access policy:
  - on course creation - default repository access across the whole organization is set to `none`, which means that only the organization owners can see any private repository on that organization
  - when students enroll, they receive read/pull access to `assignments` repository and write/push access to a personal student repository as GitHub invitations to their registered GitHub email

### Teams

Student groups will have GitHub teams with the same name created in the course organization.
Group records in the Autograder's database will have references to the corresponding GitHub team ID's.


## Simple testing of UI with dummy data

Since this is a single page application, then you would only be able to navigate to the default index page, every other navigation is handle by javascript. The way this works at the server, is that every request to /app/ returns the index.html page.
Now to get around the problem, we have to use the built in navigation manager, navMan. To do this, go to the main page of the application and open developer tools in the browser. In the command line type in the following command debugData.navMan.navigateTo("/app/admin/courses/new"). You could replace the url with any other url as you like.

## Docker

Autograder application will build code submitted by students, and run tests provided by teachers inside docker containers. An often encountered problem is Docker being unable to resolve DNS due to disabled public DNS.
If you get a build error like that:

```
Docker execution failed{error 25 0  Error response from daemon: Get https://registry-1.docker.io/v2/: dial tcp: lookup registry-1.docker.io on [::1]:53: read udp [::1]:50111->[::1]:53: read: connection refused}
```

then it must be a DNS problem.

One of the solutions is to uncomment or change `DOCKER_OPTS` line in `/etc/default/docker` file, then restart Docker deamon with `service docker restart`.

[Problem description and possible solutions](https://development.robinwinslow.uk/2016/06/23/fix-docker-networking-dns/)

## Webpack

The project uses `Webpack` to compile and bundle all TypeScript files. Currently used TypeScript loader is `awesome-typescript-loader`. It is known for its irregular updates. In case of it having any vulnerable dependencies, `awesome-typescript-loader` can be easily perlaced with an alternative loader - `ts-loader`. To start using another loader, replace `awesome-typescript-loader` line with `ts-loader` in `/public/webpack.config.js`

## npm

`npm install` (or `npm i`) no longer installs all dependencies with versions stated in `package-lock.json`, but will also attempt to load latest versions for all root packages. If you just want to install the package dependencies without altering your `package-lock.json`, run `npm ci` instead.








