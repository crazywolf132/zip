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
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) -v
	./$(BINARY_NAME)

deps:
	$(GOGET) -v -t -d ./...

vet:
	$(GOVET) ./...

fmt:
	$(GOFMT) ./...

lint:
	golangci-lint run

# Cross compilation
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_UNIX) -v

# Installation
install: build
	mv $(BINARY_NAME) $(GOPATH)/bin

# Docker
docker-build:
	docker build -t zip .

docker-run:
	docker run -it --rm zip

# Help
help:
	@echo "Make targets:"
	@echo "  build        - Build the zip binary"
	@echo "  clean        - Clean build artifacts"
	@echo "  test         - Run tests"
	@echo "  coverage     - Run tests with coverage"
	@echo "  run          - Build and run zip"
	@echo "  deps         - Get dependencies"
	@echo "  vet          - Run go vet"
	@echo "  fmt          - Run go fmt"
	@echo "  lint         - Run golangci-lint"
	@echo "  build-linux  - Cross-compile for Linux"
	@echo "  install      - Install zip to GOPATH/bin"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-run   - Run zip in a Docker container"

# Default target
default: build