DOCKER_IMAGE := mpolden/echoip
OS := $(shell uname)
ifeq ($(OS),Linux)
	TAR_OPTS := --wildcards
endif

all: deps lint test install

deps:
	go get ./...

test: deps
	go test ./...

vet: deps
	go vet ./...

check-fmt:
	bash -c "diff --line-format='%L' <(echo -n) <(gofmt -d -s .)"

lint: check-fmt vet

install: deps
	go install ./...

databases := GeoLite2-City GeoLite2-Country

$(databases):
	mkdir -p data
	curl -fsSL -m 30 https://geolite.maxmind.com/download/geoip/database/$@.tar.gz | tar $(TAR_OPTS) --strip-components=1 -C $(CURDIR)/data -xzf - '*.mmdb'
	test ! -f data/GeoLite2-City.mmdb || mv data/GeoLite2-City.mmdb data/city.mmdb
	test ! -f data/GeoLite2-Country.mmdb || mv data/GeoLite2-Country.mmdb data/country.mmdb

geoip-download: $(databases)

docker-build:
	docker build -t $(DOCKER_IMAGE) .

docker-login:
	@echo "$(DOCKER_PASSWORD)" | docker login -u "$(DOCKER_USERNAME)" --password-stdin

docker-test:
	$(eval CONTAINER=$(shell docker run --rm --detach --publish-all $(DOCKER_IMAGE)))
	$(eval DOCKER_PORT=$(shell docker port $(CONTAINER) | cut -d ":" -f 2))
	curl -fsS -m 5 localhost:$(DOCKER_PORT) > /dev/null; docker stop $(CONTAINER)

docker-push: docker-test docker-login
	docker push $(DOCKER_IMAGE)
