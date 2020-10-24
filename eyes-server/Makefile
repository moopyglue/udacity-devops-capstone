go: main.go 
	# ------------------------------------------------------
	# build the go executable
	mkdir -p ${PWD}/tmpbuild
	GOPATH=${PWD}/tmpbuild go get github.com/gorilla/websocket
	GOPATH=${PWD}/tmpbuild go get github.com/timtadh/getopt
	GOPATH=${PWD}/tmpbuild go build -o eyes-go-server .

lint:
	# ------------------------------------------------------
	# go linter
	mkdir -p ${PWD}/tmpbuild
	GOPATH=${PWD}/tmpbuild go get github.com/gorilla/websocket
	GOPATH=${PWD}/tmpbuild go get github.com/timtadh/getopt
	GOPATH=${PWD}/tmpbuild ../bin/golangci-lint run ./*.go
	# ------------------------------------------------------
	# See local hadolint install instructions:   https://github.com/hadolint/hadolint
	# This is linter for Dockerfiles
	../bin/hadolint Dockerfile

docker: 
	# Build image and add a descriptive tag
	docker build --tag eyes-server:latest .
	docker images eyes-server:latest
	../bin/hadolint Dockerfile
	
build: go docker 

run: build foreground

upload:
	# 
	./upload_docker.bash

server: build background

foreground:
	# run 'all' then in foreground on port 6161
	( docker stop eyes-foreground ; echo )
	docker run --rm --name eyes-foreground -p 6001:6001 eyes-server:latest standalone
    
background: 
	# run 'all' then in background on port 6666
	( docker stop eyes ; echo )
	docker run --rm --detach --name eyes -p 6666:6001 eyes-server:latest standalone

runhub:
	( docker stop eyes-hub ; echo )
	docker run --rm --name eyes-hub -p 7070:7070 eyes-server:latest hub 
    
runcarrier1:
	docker run --rm --name eyes-carrier1 -p 6001:6001 eyes-server:latest carrier

runcarrier2:
	docker run --rm --name eyes-carrier2 -p 6002:6001 eyes-server:latest carrier

runcarrier3:
	docker run --rm --name eyes-carrier3 -p 6003:6001 eyes-server:latest carrier
 
