#!/bin/bash

brew install go
go env GOROOT GOPATH

# Add GOPATH to env variables
export GOPATH=`go env GOPATH`

mkdir $GOPATH/src
mkdir $GOPATH/bin

# Add GOBIN to env variables
export GOBIN=$GOPATH/bin

git clone https://github.com/commerceblock/mainstay $GOPATH/src/mainstay
cd $GOPATH/src/mainstay

# Add bin to $PATH
PATH="$GOPATH/bin:$PATH"

# Download and install dependencies
go get

# Compile packages and dependencies
go build
