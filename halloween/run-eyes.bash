#!/usr/bin/env bash

# Step 1 
# Build the go
export GOPATH=$PWD/tmpbuild
go get github.com/gorilla/websocket
go build -o ./eyes-go-server .
[[ $? != 0 ]] && exit $?

# Step 2:
# Build image and add a descriptive tag
docker build --tag eyes-server:latest .
[[ $? != 0 ]] && exit $?

# Step 3:
# List docker images
docker images eyes-server:latest
[[ $? != 0 ]] && exit $?

# Step 3:
# Run theapp
docker run --rm --name eyes -p 6161:8080 eyes-server:latest
