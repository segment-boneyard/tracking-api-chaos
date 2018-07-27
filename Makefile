VERSION := $(shell git describe --tags --always --dirty="-dev")
LDFLAGS := -ldflags='-X "main.Version=$(VERSION)"'
DOCKER_TAG := "tracking-api-chaos:$(VERSION)"

.PHONY: vendor test go-bindata

build: vendor
	go build $(LDFLAGS) ./cmd/tracking-api-chaos

vendor:
	go get -v github.com/kardianos/govendor
	govendor sync -v

docker:
	docker build -f Dockerfile -t $(DOCKER_TAG) .

test: vendor
	govendor test -race -cover -v +local
