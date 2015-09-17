NAME=ifconfigd

all: deps test install

deps:
	go get -d -v

fmt:
	go fmt ./...

test:
	go test ./...

install:
	go install

get-geoip-db:
	curl -s http://geolite.maxmind.com/download/geoip/database/GeoLite2-Country.mmdb.gz | gunzip > GeoLite2-Country.mmdb
