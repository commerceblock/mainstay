#!/bin/bash

brew install go
go env GOROOT GOPATH

mkdir $GOPATH/src
mkdir $GOPATH/bin

export GOBIN=$GOPATH/bin

git clone https://github.com/commerceblock/mainstay $GOPATH/src/mainstay
cd $GOPATH/src/mainstay

go build
go install
