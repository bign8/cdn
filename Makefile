# Documentation: https://www.gnu.org/software/make/manual/html_node/index.html
VERSION=0.0.0.`git rev-parse --short HEAD`
UTILS=$(shell find util -name '*.go')
GOFLAGS=-i -v -ldflags "-s -w -X github.com/bign8/cdn/util/health.Version=${VERSION}" -installsuffix cgo

all: client/main origin/main server/main ui/main

%/main: %/Dockerfile %/*.go %/*/* ${UTILS}
	go test ./$*
	CGO_ENABLED=0 GOOS=linux go build ${GOFLAGS} -o $@ ./$*
	docker build --rm -t bign8/cdn/$*:latest ./$*

test:
	go test -v $$(glide nv) -cover -bench=. -benchmem

install:
	glide install --strip-vendor

clean:
	@if [ -f client/main ] ; then rm client/main ; fi
	@if [ -f server/main ] ; then rm server/main ; fi
	@if [ -f origin/main ] ; then rm origin/main ; fi
	@if [ -f ui/main ] ; then rm ui/main ; fi

.PHONY: clean all test install
