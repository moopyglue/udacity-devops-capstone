#!/bin/bash -xv
cd /app
# env | sort
exec /app/eyes-go-server "$@"
