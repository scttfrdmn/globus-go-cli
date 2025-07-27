# SPDX-License-Identifier: Apache-2.0
# SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors

.PHONY: build test lint clean

BINARY_NAME=globus
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILDFLAGS=-ldflags "-X github.com/scttfrdmn/globus-go-cli/cmd.Version=$(VERSION)"

all: build

build:
	go build $(BUILDFLAGS) -o $(BINARY_NAME) .

install: build
	cp $(BINARY_NAME) $(GOPATH)/bin/

lint:
	staticcheck ./...

test:
	go test -v -cover ./...

test-integration:
	go test -v -tags=integration ./...

test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

clean:
	go clean
	rm -f $(BINARY_NAME) coverage.out coverage.html