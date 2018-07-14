# Server Installation

This guide details the specifics for go and github, but should be similar setup for gitlab.

1. Install a linux distro server edition (For this we choose ubuntu LTS)
2. Install Golang (Recommended 1.10)
    1. Create folder go in $HOME
    ```
    mkdir $HOME/go
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

    For Github see <a href="GithubSetup.MD"> Github Setup</a>
    
    For Gitlab see <a href="GitlabSetup.MD"> Gitlab Setup</a>


6. Download and set up autograder dev environment
    1. Add environment variables in $HOME/.bashrc
    ```
    export GITHUB_KEY = "key"
    export GITHUB_SECRET = "secret"
    export GOPATH = $HOME/go/
    ```
    ``` 
    go get github.com/autograde/aguis
    go get github.com/autograde/kit (????) 
    cd $GOPATH/src/autograde/aguis/public
    npm install
    ```
7. Build Autograde
    ```
     cd $GOPATH/src/autograde/aguis
     go install
    ```

8. Starting Autograder with specific Database and web path.
    ```
    ./aguis -database.file ./ag.db -http.public $GOPATH/src/github.com/autograde/aguis/public -service.url <url> -script.path <buildscript path>
    ```

9. If this setup is used for production, go into public/index.html and exchange the react-development library with production library
    ```
    <script src="https://unpkg.com/react@16.4.1/umd/react.development.js"></script>
    <script src="https://unpkg.com/react-dom@16.4.1/umd/react-dom.development.js"></script>
    
    to

    <script src="https://unpkg.com/react@16.4.1/umd/react.production.min.js"></script>
    <script src="https://unpkg.com/react-dom@16.4.1/umd/react-dom.production.min.js"></script>
    ```