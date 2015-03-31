export GOPATH := $(shell pwd)

.PHONY: server test

all: server test

server:
	go install -v ./...

test:
	go test -v ./...
