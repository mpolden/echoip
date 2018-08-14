OS := $(shell uname)
ifeq ($(OS),Linux)
	TAR_OPTS := --wildcards
endif

all: deps lint test install

fmt:
	go fmt ./...

test:
	go test ./...

vet:
	go vet ./...

megacheck:
ifdef TRAVIS
	megacheck 2> /dev/null; if [ $$? -eq 127 ]; then \
		go get -v honnef.co/go/tools/cmd/megacheck; \
	fi
	megacheck ./...
endif

check-fmt:
	bash -c 'diff --line-format="%L" <(echo -n) <(gofmt -d -s $$(find . -type f -name "*.go" -not -path "./vendor/*"))'

lint: check-fmt vet megacheck

deps:
	go get -d -v ./...

install:
	go install ./...

databases := GeoLite2-City GeoLite2-Country

$(databases):
	mkdir -p data
	curl -fsSL -m 30 http://geolite.maxmind.com/download/geoip/database/$@.tar.gz | tar $(TAR_OPTS) --strip-components=1 -C $(PWD)/data -xzf - '*.mmdb'
	test ! -f data/GeoLite2-City.mmdb || mv data/GeoLite2-City.mmdb data/city.mmdb
	test ! -f data/GeoLite2-Country.mmdb || mv data/GeoLite2-Country.mmdb data/country.mmdb

geoip-download: $(databases)
