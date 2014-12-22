NAME = ifconfig

all: deps test build

fmt:
	gofmt -w=true *.go

deps:
	go get -d -v

build:
	@mkdir -p bin
	go build -o bin/$(NAME)

test:
	go test

docker-image:
	docker build -t martinp/ifconfig .
