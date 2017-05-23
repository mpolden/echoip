FROM golang:1.8.1-alpine

WORKDIR /app
COPY Makefile .

RUN apk add --update --no-cache git make curl && \
    go get github.com/martinp/ipd && \
    make get-geoip-dbs

COPY index.html .

EXPOSE 8080
ENTRYPOINT ["ipd"]
CMD ["-f", "/app/GeoLite2-Country.mmdb", "-c", "/app/GeoLite2-City.mmdb", "--port-lookup", "--reverse-lookup", "-L", "debug"]

