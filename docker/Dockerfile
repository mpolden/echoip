FROM golang:alpine as build-env

RUN apk --no-cache add git

RUN go get github.com/mpolden/ipd


FROM alpine as geoipupdate-env

RUN apk add -q --update alpine-sdk curl-dev file wget zlib-dev git autoconf automake libtool\
    && mkdir -p /tmp/geoipupdate && cd /tmp/geoipupdate \
    && git clone --depth=1 https://github.com/maxmind/geoipupdate.git . \
    && ./bootstrap \
    && ./configure \
    && make \
    && make install \
    && mkdir -p /tmp/geoip \
    && geoipupdate -d /tmp/geoip

FROM alpine

MAINTAINER Alessandro Lucaferro <alessandro.lucaferro@gmail.com>

RUN mkdir ipd

WORKDIR /ipd

COPY --from=build-env /go/bin/ipd /ipd/ipd

COPY --from=geoipupdate-env /tmp/geoip/* /ipd/


ENTRYPOINT ["/ipd/ipd"]

EXPOSE 8080
