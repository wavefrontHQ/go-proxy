PREFIX := /usr/local
VERSION := 0.2
TAG := $(shell git describe --exact-match --tags 2>/dev/null)
COMMIT := $(shell git rev-parse --short HEAD)
BRANCH := $(shell git rev-parse --abbrev-ref HEAD)

ifdef GOBIN
PATH := $(GOBIN):$(PATH)
else
PATH := $(subst :,/bin:,$(GOPATH))/bin:$(PATH)
endif

PROXY := wavefront-proxy
LDFLAGS := $(LDFLAGS) -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.branch=$(BRANCH)

ifdef TAG
        LDFLAGS += -X main.tag=$(TAG)
endif

all:
	$(MAKE) deps
	$(MAKE) proxy

deps:
	go get github.com/satori/go.uuid
	go get github.com/rcrowley/go-metrics
	go get github.com/spf13/viper

proxy:
	go build -i -o $(PROXY) -ldflags "$(LDFLAGS)" ./cmd/wavefront-proxy/proxy.go

# Build linux executables/packages
package:
	$(MAKE) deps
	./pkg/build.sh $(VERSION)

# Build executables for all platforms
package-all:
	$(MAKE) deps
	./pkg/build.sh $(VERSION) -all

go-install:
	go install -ldflags "-w -s $(LDFLAGS)" ./cmd/wavefront-proxy

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
	-rm -rf ./build

.PHONY: deps proxy cmd wavefront-proxy install test lint test-all clean
