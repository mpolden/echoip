FROM golang:onbuild

EXPOSE 8080

ADD ./assets /go/bin/assets
ADD ./index.html /go/bin/

ENTRYPOINT ["/go/bin/app"]
