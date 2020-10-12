build: main.go 
	# ------------------------------------------------------
	# build the go executable
	export GOPATH=$PWD/tmpbuild
	go get github.com/gorilla/websocket
	go build -o eyes-go-server .

lint:
	# ------------------------------------------------------
	# See local hadolint install instructions:   https://github.com/hadolint/hadolint
	# This is linter for Dockerfiles
	../bin/hadolint Dockerfile
	# ------------------------------------------------------
	# This is a linter for Python source code linter: https://www.pylint.org/
	# This should be run from inside a virtualenv
	../bin/golangci-lint run ./*.go

docker:
	# Build image and add a descriptive tag
	docker build --tag eyes-server:latest .
	docker images eyes-server:latest
	
all: lint build docker

run: all foreground

server: all background

foreground:
	# run 'all' then in foreground on port 6161
	( docker stop eyes ; echo )
	docker run --rm --name eyes -p 6161:8080 eyes-server:latest
    
background: 
	# run 'all' then in background on port 6666
	( docker stop eyes-background ; echo )
	docker run --rm --detach --name eyes-background -p 6666:8080 eyes-server:latest

