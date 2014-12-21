NAME = ifconfig

all: test build

fmt:
	gofmt -w=true *.go

build:
	@mkdir bin
	go build -o bin/$(NAME)

test:
	go test
