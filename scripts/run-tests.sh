#!/bin/bash

alias dir="$GOPATH/src/mainstay"

# run tests sequentially
cd $GOPATH/src/mainstay
go test -v=0 -p=1 ./...
