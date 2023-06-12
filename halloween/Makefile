go: main.go 
	# ------------------------------------------------------
	# build the go executable
	mkdir -p ${PWD}/tmpbuild
	GOPATH=${PWD}/tmpbuild go get github.com/gorilla/websocket
	GOPATH=${PWD}/tmpbuild go get github.com/timtadh/getopt
	GOPATH=${PWD}/tmpbuild go build -o eyes-go-server .

docker: 
	# Build image and add a descriptive tag
	docker build --tag eyes-server:latest .
	docker images eyes-server:latest
	../bin/hadolint Dockerfile
	
build: go docker 

run: build foreground

server: build halloween

halloween: 
	# run 'all' then in halloween on port 7777
	( docker stop halloween ; echo )
	docker run --rm --detach --name halloween -p 7777:6001 eyes-server:latest standalone
