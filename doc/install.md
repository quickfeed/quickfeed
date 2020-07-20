# Server Installation

This guide details the specifics for go and github, but should be similar setup for gitlab.

(Note to self: When running in the dev environment at UiS, you may wish to update the NGINX configuration on ag2. That is, the following file needs to be updated: `/etc/nginx/sites-available/default` with a new server entry etc.)

(Another note to self: Merge this with the README.md documentation.)
(And: make the process more streamlined; can we avoid the teacher authorization thing, or link to it so that we don't have to remember this...)

To set up IPv4 port forwarding on UiS servers:
```
sudo sysctl net.ipv4.ip_forward=1
sudo sysctl -p  (to confirm the change)
sudo service docker restart
```

1. Install a linux distro server edition (For this we choose ubuntu LTS)
2. Install Golang (Recommended 1.10)
    1. Create folder go in $HOME
    ```
    mkdir $HOME/go
    ```

    (If running on a UiS server; here is how to update Go)
    ```
	wget -O /tmp/go.tar.gz https://dl.google.com/go/go1.12.6.linux-amd64.tar.gz
	[[ -d /usr/local/go ]] && rm -rf /usr/local/go
	tar -C /usr/local -zxf /tmp/go.tar.gz
	rm -f /tmp/go.tar.gz
    ```

3. Install Docker CE via repository (Recommended 18.03)
    
    For ubuntu 
        
        https://docs.docker.com/install/linux/docker-ce/ubuntu/#prerequisites

    Note:
        Other versions should work, but only tested with 18.03 and 17.05

4. Install NodeJS (Only needed for development/compiling, can be omitted for production)
   
    1. Install npm
    ```
    sudo apt install npm
    ```
    2. Install these packages
    ```
    npm install -g webpack
    npm install -g webpack-cli
    ```

5. Setting up the source control managment.

    For GitHub see <a href="GithubSetup.MD"> GitHub Setup</a>
    
    For GitLab see <a href="GitlabSetup.MD"> GitLab Setup</a>


6. Download and set up autograder dev environment
    1. Add environment variables in $HOME/.bashrc
    ```
    export GITHUB_KEY="key"
    export GITHUB_SECRET="secret"
    export GOPATH=$HOME/go/
    ```
    ``` 
    go get github.com/autograde/quickfeed
    cd $GOPATH/src/autograde/quickfeed/public
    npm install
    ```
7. Build Autograde
    ```
     cd $GOPATH/src/autograde/quickfeed
     go install
    ```

8. Starting Autograder with specific Database and web path.
    ```
    ./quickfeed -database.file ./ag.db -http.public $GOPATH/src/github.com/autograde/quickfeed/public -service.url <url> -script.path <buildscript path>
    ```

9. If this setup is used for production, go into public/index.html and exchange the react-development library with production library
    ```
    <script src="https://unpkg.com/react@16.4.1/umd/react.development.js"></script>
    <script src="https://unpkg.com/react-dom@16.4.1/umd/react-dom.development.js"></script>
    
    to

    <script src="https://unpkg.com/react@16.4.1/umd/react.production.min.js"></script>
    <script src="https://unpkg.com/react-dom@16.4.1/umd/react-dom.production.min.js"></script>
    ```
