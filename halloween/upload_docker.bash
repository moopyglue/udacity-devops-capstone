#!/usr/bin/env bash
# This file tags and uploads an image to Docker Hub

# Assumes that an image is built via `run_docker.sh`

source_tag=eyes-server:latest
destination_user=qmkey
destination_reg="https://index.docker.io/v1/"
version=latest
destination_tag=$destination_user/udacity-capstone-eyes:$version

echo "Docker Source ID and Image: $source_tag"
echo "Docker Destination ID and Image: $destination_tag"

# Authentication
credentials="$( docker system info 2>/dev/null | grep -E " (Username|Registry): " | awk '{ printf $2 ":"  }' )" 
[[ $credentials == "$destination_user:$destination_reg:" ]] || docker login --username $destination_user

# Push image to a docker repository
docker tag $source_tag $destination_tag
docker push $destination_tag
[[ $? != 0 ]] && exit $?

# If run by jenkins then tag with build number also
if [[ $JENKINS_NODE_COOKIE != "" && $BUILD_NUMBER != ""  ]] ; then
    version=v${BUILD_NUMBER}
    destination_tag=$destination_user/udacity-capstone-eyes:$version
    docker tag $source_tag $destination_tag
    docker push $destination_tag
fi

exit $?
