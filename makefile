.PHONY: server test all

all: server test

server:
	go clean ./...
	go install -v ./...

test:
	go test -v ./...
