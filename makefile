export GOPATH := $(shell pwd)

.PHONY: server test all travis
all: server test
travis: all

server:
	go install -v ./...

test:
	go test -v ./...
