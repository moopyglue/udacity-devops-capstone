#!/usr/bin/env bash

## Complete the following steps to get Docker running locally

# Step 1:
# Build image and add a descriptive tag
docker build --tag app-server:latest .
[[ $? != 0 ]] && exit $?

# Step 2:
# List docker images
docker images app-server:latest
[[ $? != 0 ]] && exit $?

# Step 3:
# Run flask app
docker run --rm --name nginx -p 8181:80 app-server:latest
