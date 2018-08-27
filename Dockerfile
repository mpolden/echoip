# Compile
FROM golang:1.11-alpine AS build
ADD . /go/src/github.com/mpolden/echoip
WORKDIR /go/src/github.com/mpolden/echoip
RUN apk --update add git gcc musl-dev
ENV GO111MODULE=on
RUN go get -d -v .
RUN go install ./...

# Run
FROM alpine
RUN mkdir -p /opt/
COPY --from=build /go/bin/echoip /opt/
WORKDIR /opt/
ENTRYPOINT ["/opt/echoip"]
