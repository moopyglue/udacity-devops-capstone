#!/usr/bin/env bash

# Step 1 
# Build the go
export GOPATH=$PWD/tmpbuild
go build .
[[ $? != 0 ]] && exit $?

./store
