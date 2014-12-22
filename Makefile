NAME = ifconfig

all: deps test build

fmt:
	gofmt -w=true *.go

deps:
	go get -d -v

build:
	@mkdir bin
	go build -o bin/$(NAME)

test:
	go test
