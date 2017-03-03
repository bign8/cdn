# Documentation: https://www.gnu.org/software/make/manual/html_node/index.html
VERSION=0.0.0.`git rev-parse --short HEAD`
UTILS=$(shell find util -name '*.go')
GOFLAGS=-i -v -ldflags "-s -w -X main.Version=${VERSION}" -installsuffix cgo

all: client/main origin/main server/main

client/main: client/Dockerfile client/*.go ${UTILS}
	go test ./client
	CGO_ENABLED=0 GOOS=linux go build ${GOFLAGS} -o $@ ./client
	docker build -q --rm -t bign8/cdn:client-latest ./client

origin/main: origin/Dockerfile origin/*.go ${UTILS}
	go test ./origin
	CGO_ENABLED=0 GOOS=linux go build ${GOFLAGS} -o $@ ./origin
	docker build -q --rm -t bign8/cdn:origin-latest ./origin

server/main: server/Dockerfile server/*.go ${UTILS}
	go test ./server
	CGO_ENABLED=0 GOOS=linux go build ${GOFLAGS} -o $@ ./server
	docker build -q --rm -t bign8/cdn:server-latest ./server

clean:
	@if [ -f client/main ] ; then rm client/main ; fi
	@if [ -f server/main ] ; then rm server/main ; fi
	@if [ -f origin/main ] ; then rm origin/main ; fi

.PHONY: clean all
