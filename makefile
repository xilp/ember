.PHONY: server test all travis

all: server test
travis: all

server:
	go clean ./...
	go install -v ./...

test:
	go test -v ./...
