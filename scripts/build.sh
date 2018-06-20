#!/bin/bash

brew install go
go env GOROOT GOPATH

mkdir $GOPATH/src
mkdir $GOPATH/bin

export GOBIN=$GOPATH/bin

git clone https://github.com/commerceblock/ocean-attestation $GOPATH/src/ocean-attestation
cd $GOPATH/src/ocean-attestation

go build
go install
ocean-attestation
