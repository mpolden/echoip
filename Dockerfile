# Compile
FROM golang:alpine AS build
ADD . /go/src/github.com/mpolden/ipd
WORKDIR /go/src/github.com/mpolden/ipd
RUN apk --update add git
RUN go get -d -v ./...
RUN go install ./...

# Run
FROM alpine
RUN mkdir -p /opt/
COPY --from=build /go/bin/ipd /opt/
WORKDIR /opt/
ENTRYPOINT ["/opt/ipd"]
