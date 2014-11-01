TARGET = ifconfig

all: install

clean:
	rm -f -- $(TARGET)

fmt:
	gofmt -w=true *.go

install:
	go build $(TARGET).go

test:
	go test
