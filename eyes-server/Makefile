go: main.go 
	# ------------------------------------------------------
	# build the go executable
	export GOPATH=$PWD/tmpbuild
	go get github.com/gorilla/websocket
	go get github.com/timtadh/getopt
	go build -o eyes-go-server .
	../bin/golangci-lint run ./*.go

lint:
	# ------------------------------------------------------
	# go linter
	../bin/golangci-lint run ./*.go
	# ------------------------------------------------------
	# See local hadolint install instructions:   https://github.com/hadolint/hadolint
	# This is linter for Dockerfiles
	../bin/hadolint Dockerfile

docker:
	# Build image and add a descriptive tag
	docker build --tag eyes-server:latest .
	docker images eyes-server:latest
	../bin/hadolint Dockerfile
	
all: go docker 

run: all foreground

server: all background

foreground:
	# run 'all' then in foreground on port 6161
	( docker stop eyes-foreground ; echo )
	docker run --rm --name eyes-foreground -p 6161:6001 eyes-server:latest
    
background: 
	# run 'all' then in background on port 6666
	( docker stop eyes ; echo )
	docker run --rm --detach --name eyes -p 6666:6001 eyes-server:latest

test1:
	docker run --rm --name eyes-1 -p 6001:6001 -p 7771:7070 eyes-server:latest
    
test2:
	docker run --rm --name eyes-2 -p 6002:6002 -p 7772:7070 eyes-server:latest
    
test3:
	docker run --rm --name eyes-3 -p 6003:6003 -p 7773:7070 eyes-server:latest
    
