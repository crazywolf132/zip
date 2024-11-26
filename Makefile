# Makefile for zip VCS

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOVET=$(GOCMD) vet
GOFMT=$(GOCMD) fmt
BINARY_NAME=zip
BINARY_UNIX=$(BINARY_NAME)_unix

# Build flags
LDFLAGS=-ldflags "-X main.Version=$(shell git describe --tags --always --long --dirty)"

# Main package path
MAIN_PACKAGE=./cmd/zip

.PHONY: all build clean test coverage run deps vet fmt lint help

all: test build

build:
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) -v $(MAIN_PACKAGE)

clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)

test:
	$(GOTEST) -v ./...

coverage:
	$(GOTEST) -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out

run:
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) -v $(MAIN_PACKAGE)
	./$(BINARY_NAME)

deps:
	$(GOGET) -v -t -d ./...

vet:
	$(GOVET) ./...

fmt:
	$(GOFMT) ./...

lint:
	golangci-lint run

help:
	@echo "Available targets:"
	@echo "  build     - Build the binary"
	@echo "  clean     - Clean build files"
	@echo "  test      - Run tests"
	@echo "  coverage  - Generate test coverage"
	@echo "  run       - Build and run"
	@echo "  deps      - Get dependencies"
	@echo "  vet       - Run go vet"
	@echo "  fmt       - Run go fmt"
	@echo "  lint      - Run linter"