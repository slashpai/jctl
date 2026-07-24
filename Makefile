BINARY      := jctl
MODULE      := github.com/slashpai/jctl
GOFILES     := $(shell find . -name '*.go' -not -path './vendor/*')
GIT_VERSION := $(shell git describe --tags --exact-match 2>/dev/null)
GIT_COMMIT  := $(shell git rev-parse --short HEAD 2>/dev/null)
LDFLAGS     := -X $(MODULE)/internal/version.version=$(GIT_VERSION) -X $(MODULE)/internal/version.commit=$(GIT_COMMIT)

.PHONY: build install clean test lint fmt vet tidy update run help

## build: compile the binary
build:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) .

## install: install to $GOPATH/bin
install:
	go install -ldflags "$(LDFLAGS)" .

## clean: remove build artifacts
clean:
	rm -f $(BINARY)
	go clean -cache -testcache

## test: run all tests with verbose output
test:
	go test -v ./...

## lint: run go vet and staticcheck (if installed)
lint: vet
	@which staticcheck > /dev/null 2>&1 && staticcheck ./... || echo "staticcheck not installed — run: go install honnef.co/go/tools/cmd/staticcheck@latest"

## vet: run go vet
vet:
	go vet ./...

## fmt: format all Go files
fmt:
	gofmt -s -w $(GOFILES)

## fmt-check: check formatting without writing
fmt-check:
	@test -z "$$(gofmt -l $(GOFILES))" || (echo "Files need formatting:" && gofmt -l $(GOFILES) && exit 1)

## tidy: tidy go.mod and go.sum
tidy:
	go mod tidy

## update: update all dependencies
update:
	go get -u ./...
	go mod tidy

## run: build and run with arguments (usage: make run ARGS="issue list -p PROJ")
run: build
	./$(BINARY) $(ARGS)

## help: show this help
help:
	@echo "Usage: make <target>"
	@echo ""
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## //' | column -t -s ':'
