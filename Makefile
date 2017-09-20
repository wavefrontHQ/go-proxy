PREFIX := /usr/local
VERSION := 0.1

ifdef GOBIN
PATH := $(GOBIN):$(PATH)
else
PATH := $(subst :,/bin:,$(GOPATH))/bin:$(PATH)
endif

PROXY := wavefront-proxy

all:
	$(MAKE) deps
	$(MAKE) proxy

deps:
	go get github.com/satori/go.uuid

proxy:
	go build -i -o $(PROXY) ./proxy/proxy.go

go-install:
	go install ./proxy

install: proxy
	mkdir -p $(DESTDIR)$(PREFIX)/bin/
	cp $(PROXY) $(DESTDIR)$(PREFIX)/bin/

test:
	go test -short ./...

lint:
	go vet ./...

test-all: lint
	go test ./...

clean:
	-rm -f wavefront-proxy
	-rm -f wavefront-proxy.exe

.PHONY: deps proxy wavefront-proxy install test lint test-all clean
