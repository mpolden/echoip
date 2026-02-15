DOCKER ?= docker
DOCKER_IMAGE ?= mpolden/echoip
OS := $(shell uname)
ifeq ($(OS),Linux)
	TAR_OPTS := --wildcards
endif
XGOARCH := amd64
XGOOS := linux
XBIN := $(XGOOS)_$(XGOARCH)/echoip

all: checkfmt vet test install

test:
	go test ./...

vet:
	go vet ./...

checkfmt:
	@sh -c "test -z $$(gofmt -l .)" || { echo "one or more files need to be formatted: try make fmt to fix this automatically"; exit 1; }

fmt:
	gofmt -w .

install:
	go install ./...

databases := GeoLite2-City GeoLite2-Country GeoLite2-ASN

$(databases):
ifndef GEOIP_LICENSE_KEY
	$(error GEOIP_LICENSE_KEY and MAXMIND_ACCOUNT_ID must be set. See https://dev.maxmind.com/geoip/updating-databases/#directly-downloading-databases
endif
ifndef MAXMIND_ACCOUNT_ID
	$(error GEOIP_LICENSE_KEY and MAXMIND_ACCOUNT_ID must be set. See https://dev.maxmind.com/geoip/updating-databases/#directly-downloading-databases
endif
	mkdir -p data
	@curl -fsSL -m 30 -u $(MAXMIND_ACCOUNT_ID):$(GEOIP_LICENSE_KEY) "https://download.maxmind.com/geoip/databases/$@/download?suffix=tar.gz" | tar $(TAR_OPTS) --strip-components=1 -C $(CURDIR)/data -xzf - '*.mmdb'
	test ! -f data/GeoLite2-City.mmdb || mv data/GeoLite2-City.mmdb data/city.mmdb
	test ! -f data/GeoLite2-Country.mmdb || mv data/GeoLite2-Country.mmdb data/country.mmdb
	test ! -f data/GeoLite2-ASN.mmdb || mv data/GeoLite2-ASN.mmdb data/asn.mmdb

geoip-download: $(databases)

docker-build:
	$(DOCKER) build -t $(DOCKER_IMAGE) .

docker-login:
	@echo "$(DOCKER_PASSWORD)" | $(DOCKER) login -u "$(DOCKER_USERNAME)" --password-stdin

docker-test:
	$(eval CONTAINER=$(shell $(DOCKER) run --rm --detach --publish-all $(DOCKER_IMAGE)))
	$(eval DOCKER_PORT=$(shell $(DOCKER) port $(CONTAINER) | cut -d ":" -f 2))
	curl -fsS -m 5 localhost:$(DOCKER_PORT) > /dev/null; $(DOCKER) stop $(CONTAINER)

docker-push: docker-test docker-login
	$(DOCKER) push $(DOCKER_IMAGE)

xinstall:
	env GOOS=$(XGOOS) GOARCH=$(XGOARCH) go install ./...

publish:
ifndef DEST_PATH
	$(error DEST_PATH must be set when publishing)
endif
	rsync -a $(GOPATH)/bin/$(XBIN) $(DEST_PATH)/$(XBIN)
	@sha256sum $(GOPATH)/bin/$(XBIN)

run:
	go run cmd/echoip/main.go -a data/asn.mmdb -c data/city.mmdb -f data/country.mmdb -H x-forwarded-for -r -s -p
