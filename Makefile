NAME=ifconfigd

all: deps test install

deps:
	go get -d -v

fmt:
	go fmt ./...

test:
	go test ./...

install:
	go install
