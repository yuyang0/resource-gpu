.PHONY: deps binary build test cloc unit-test

REPO_PATH := github.com/projecteru2/resource-gpu
REVISION := $(shell git rev-parse HEAD || unknown)
BUILTAT := $(shell date +%Y-%m-%dT%H:%M:%S)
VERSION := $(shell git describe --tags $(shell git rev-list --tags --max-count=1))
GO_LDFLAGS ?= -X $(REPO_PATH)/version.REVISION=$(REVISION) \
			  -X $(REPO_PATH)/version.BUILTAT=$(BUILTAT) \
			  -X $(REPO_PATH)/version.VERSION=$(VERSION)
ifneq ($(KEEP_SYMBOL), 1)
	GO_LDFLAGS += -s
endif

deps:
	go mod vendor

binary:
	go build -ldflags "$(GO_LDFLAGS)" -o resource-gpu

build: deps binary

test: deps unit-test

.ONESHELL:

cloc:
	cloc --exclude-dir=vendor,3rdmocks,mocks,tools,gen --not-match-f=test .

unit-test:
	go vet `go list ./... | grep -v '/vendor/' | grep -v '/tools'` && \
	go test -race -timeout 600s -count=1 -vet=off -cover \
	./gpu/.

lint:
	golangci-lint run
