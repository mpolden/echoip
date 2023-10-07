DOCKER ?= docker
DOCKER_IMAGE ?= mpolden/echoip
OS := $(shell uname)
ifeq ($(OS),Linux)
	TAR_OPTS := --wildcards
endif
XGOARCH := amd64
XGOOS := linux
XBIN := $(XGOOS)_$(XGOARCH)/echoip

all: lint test install

test:
	go test ./...

vet:
	go vet ./...

check-fmt:
	bash -c "diff --line-format='%L' <(echo -n) <(gofmt -d -s .)"


lint: check-fmt vet

install: install-config
	go install ./...

install-config:
	sudo install -D etc/echoip/config.toml /etc/echoip/config.toml

databases := GeoLite2-City GeoLite2-Country GeoLite2-ASN

$(databases):
ifndef GEOIP_LICENSE_KEY
	$(error GEOIP_LICENSE_KEY must be set. Please see https://blog.maxmind.com/2019/12/18/significant-changes-to-accessing-and-using-geolite2-databases/)
endif
	mkdir -p data
	@curl -fsSL -m 30 "https://download.maxmind.com/app/geoip_download?edition_id=$@&license_key=$(GEOIP_LICENSE_KEY)&suffix=tar.gz" | tar $(TAR_OPTS) --strip-components=1 -C $(CURDIR)/data -xzf - '*.mmdb'
	test ! -f data/GeoLite2-City.mmdb || mv data/GeoLite2-City.mmdb data/city.mmdb
	test ! -f data/GeoLite2-Country.mmdb || mv data/GeoLite2-Country.mmdb data/country.mmdb
	test ! -f data/GeoLite2-ASN.mmdb || mv data/GeoLite2-ASN.mmdb data/asn.mmdb

geoip-download: $(databases)

# Create an environment to build multiarch containers (https://github.com/docker/buildx/)
docker-multiarch-builder:
	DOCKER_BUILDKIT=1 $(DOCKER) build -o . git://github.com/docker/buildx
	mkdir -p ~/.docker/cli-plugins
	mv buildx ~/.docker/cli-plugins/docker-buildx
	$(DOCKER) buildx create --name multiarch-builder --node multiarch-builder --driver docker-container --use
	$(DOCKER) run --rm --privileged multiarch/qemu-user-static --reset -p yes

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

docker-pushx: docker-multiarch-builder docker-test docker-login
	$(DOCKER) buildx build --platform linux/amd64,linux/arm64,linux/arm/v7 -t $(DOCKER_IMAGE) --push .

xinstall: install-config
	env GOOS=$(XGOOS) GOARCH=$(XGOARCH) go install ./...

publish:
ifndef DEST_PATH
	$(error DEST_PATH must be set when publishing)
endif
	rsync -a $(GOPATH)/bin/$(XBIN) $(DEST_PATH)/$(XBIN)
	@sha256sum $(GOPATH)/bin/$(XBIN)

run:
	go run cmd/echoip/main.go -a data/asn.mmdb -c data/city.mmdb -f data/country.mmdb -H x-forwarded-for -r -s -p
