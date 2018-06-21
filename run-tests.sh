#!/bin/bash

alias dir="$GOPATH/src/ocean-attestation"

# run tests sequentially
cd $GOPATH/src/ocean-attestation
go test -v -p=1 ./...
