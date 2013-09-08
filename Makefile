TARGET = ifconfig

all: install

clean:
	rm -f -- $(TARGET)

fmt:
	gofmt -tabs=false -tabwidth=4 -w=true *.go

install:
	go build $(TARGET).go

test:
	go test
