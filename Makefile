# Documentation: https://www.gnu.org/software/make/manual/html_node/index.html
VERSION=0.0.0.`git rev-parse --short HEAD`
UTILS=$(shell find util -name '*.go')
GOFLAGS=-i -v -ldflags "-s -w -X main.Version=${VERSION}" -installsuffix cgo

all: client/main origin/main server/main

%/main: %/Dockerfile %/*.go ${UTILS}
	go test ./$*
	CGO_ENABLED=0 GOOS=linux go build ${GOFLAGS} -o $@ ./$*
	docker build --rm -t bign8/cdn:$*-latest ./$*

clean:
	@if [ -f client/main ] ; then rm client/main ; fi
	@if [ -f server/main ] ; then rm server/main ; fi
	@if [ -f origin/main ] ; then rm origin/main ; fi

.PHONY: clean all
