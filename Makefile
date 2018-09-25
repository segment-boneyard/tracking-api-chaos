# Goals:
# - user can build binaries on their system without having to install special tools
# - user can fork the canonical repo and expect to be able to run CircleCI checks
#
# This makefile is meant for humans

VERSION := $(shell git describe --tags --always --dirty="-dev")
LDFLAGS := -ldflags='-X "main.Version=$(VERSION)"'
DOCKER_TAG := "tracking-api-chaos:$(VERSION)"

all: dist/tracking-api-chaos-$(VERSION)-darwin-amd64 dist/tracking-api-chaos-$(VERSION)-linux-amd64

test: | govendor
	govendor sync
	govendor test -cover -v +local

clean:
	rm -rf ./dist

docker:
	docker build -f Dockerfile -t $(DOCKER_TAG) .

dist/:
	mkdir -p dist

dist/tracking-api-chaos-$(VERSION)-darwin-amd64: | govendor dist/
	govendor sync
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o $@

dist/tracking-api-chaos-$(VERSION)-linux-amd64: | govendor dist/
	govendor sync
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o $@

govendor:
	go get -u github.com/kardianos/govendor

.PHONY: clean all govendor test docker
