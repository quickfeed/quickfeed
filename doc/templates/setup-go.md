# Installing Go

The most recent version of Go is 1.14.7 (at the time of writing).
Go 1.15 will be released soon.
If your system already comes with a version of Go, you may check your version like this:

```sh
go version
```

We recommend installing the latest version of Go; you will need at least Go 1.13 for the labs in this course.

## Installing on Linux Systems

These instructions apply if you have a Linux distribution installed natively, in a virtual machine (VM) or in the Windows Subsystem for Linux (WSL).
It is advisable to use your distribution's package manager (e.g. `apt` for Ubuntu) to make it easier to maintain the installation.

### Ubuntu-Based Systems

We recommend adding the `golang-backports` repository to get the latest version of Go, as indicated [here](https://github.com/golang/go/wiki/Ubuntu).

```sh
sudo add-apt-repository ppa:longsleep/golang-backports
sudo apt update
```

Then you can install the latest version of Go with the following command:

```sh
sudo apt install golang-go
```

Alternatively you can install Go with `sudo apt install golang`, though you might get an older version if you are not running the latest Ubuntu release.

### Other Distributions

For other Linux distributions you can install Go with the distribution's package manager.

Alternatively you can [download Go from the website](https://golang.org/dl/) and follow the [instructions on the Go website](https://golang.org/doc/install) to install it manually.

## Installing on macOS

If you use the [Homebrew package manager](https://brew.sh/) you can simply install Go with the following command:

```sh
brew install go
```

Alternatively you can [download Go from the website](https://golang.org/dl/) and follow the [instructions on the Go website](https://golang.org/doc/install) to install it manually.

## Installing on Windows

We highly recommend either running Linux in a VM or WSL if you have a Windows system.
Alternatively you can [download Go from the website](https://golang.org/dl/) and follow the [instructions on the Go website](https://golang.org/doc/install) to install it manually.

## Adding Go to PATH (Linux/Mac)

You should add `$GOPATH/bin` to `PATH` so that binaries installed with `go get` can be used from the command line.
To achieve this you should add the following line to the end of the file `$HOME/.profile`.

```sh
PATH=$PATH:$(go env GOPATH)/bin
```

This can be done with a text editor, e.g. vim or vscode, or simply by running this command:

```sh
echo 'PATH=$PATH:$(go env GOPATH)/bin' >> $HOME/.profile
```

The changes will take effect the next time you log in.
To make the changes take effect at once you should run the following command:

```sh
source $HOME/.profile
```
