DOCKER_IMAGE := mpolden/echoip
OS := $(shell uname)
ifeq ($(OS),Linux)
	TAR_OPTS := --wildcards
endif

all: lint test install

test:
	go test ./...

vet:
	go vet ./...

check-fmt:
	bash -c "diff --line-format='%L' <(echo -n) <(gofmt -d -s .)"

lint: check-fmt vet

install:
	go install ./...

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

heroku-run: geoip-download
ifndef PORT
	$(error PORT must be set)
endif
	echoip -C 1000000 -f data/country.mmdb -c data/city.mmdb -a data/asn.mmdb -p -r -H CF-Connecting-IP -H X-Forwarded-For -l :$(PORT)
