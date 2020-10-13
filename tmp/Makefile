build: main.go 
	# ------------------------------------------------------
	# build the go executable
	export GOPATH=$PWD/tmpbuild
	# go get github.com/gorilla/websocket
	go build -o sample .

run: build foreground

foreground:
	# run 'all' then in foreground on port 6161
	./sample
    
