# Server Installation

This guide details the specifics for go and github, but should not be too hard to follow this setup for gitlab.

1. Install a linux distro server edition
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

4. Install NodeJS (Only needed for development)
   
    1. Install npm
    ```
    sudo apt install npm
    ```
    2. Install these packages
    ```
    npm install -g webpack
    npm install -g webpack-cli
    ```

5. Download and set up autograder dev environment
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
6. Build Autograde
    ```
     cd $GOPATH/src/autograde/aguis
     go install
    ```

7. Starting Autograder with specific Database and web path.
    ```
    ./aguis -database.file ./ag.db -http.public $GOPATH/src/github.com/autograde/aguis/public -service.url <url> -script.path <buildscript path>
    ```

8. Add webhook url to providers organization.

    https://{baseurl}/hook/{provider}/events

