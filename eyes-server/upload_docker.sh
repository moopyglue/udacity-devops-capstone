#!/usr/bin/env bash
# This file tags and uploads an image to Docker Hub

# Assumes that an image is built via `run_docker.sh`

# Step 1:

source_tag=eyes-server:latest
destination_user=qmkey
destination_reg="https://index.docker.io/v1/"
destination_tag=$destination_user/udacity:capstone-eyes-server

echo "Docker Source ID and Image: $source_tag"
echo "Docker Destination ID and Image: $destination_tag"

# Step 2:  
# Tag and Authentication

docker tag $source_tag $destination_tag

credentials="$( docker system info 2>/dev/null | grep -E " (Username|Registry): " | awk '{ printf $2 ":"  }' )" 
[[ $credentials == "$destination_user:$destination_reg:" ]] || docker login --username $destination_user

# Step 3:
# Push image to a docker repository
docker push $destination_tag
exit $?
