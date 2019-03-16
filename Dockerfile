# Build
FROM golang:1.12-stretch AS build
WORKDIR /go/src/github.com/mpolden/echoip
COPY . .
ENV GO111MODULE=on
RUN make

# Run
FROM scratch
EXPOSE 8080
COPY --from=build \
     /go/bin/echoip \
     /go/src/github.com/mpolden/echoip/index.html \
     /opt/echoip/
WORKDIR /opt/echoip
ENTRYPOINT ["/opt/echoip/echoip"]
