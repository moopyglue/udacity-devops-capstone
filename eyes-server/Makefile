all: lint build

build: main.go 
	# ------------------------------------------------------
	# build the go executable
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

