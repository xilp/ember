export GOPATH := $(shell pwd)

.PHONY: server test

all: server test

travis: all

server:
	go install -v ./...

test:
	go test -v ./...
