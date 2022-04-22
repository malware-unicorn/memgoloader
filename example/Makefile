# Makefile
SHELL := /bin/bash
GOCMD=go
GOFMT=gofmt
GOPATH=$(shell $(GOCMD) env GOPATH)
#GOBUILD=$(GOCMD) build -ldflags="-extldflags=-static"
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
BINARY_NAME=memgoloader
GOAPP=./cmd/$(BINARY_NAME)
BINDIR=./bin
GOBIN=$(BINDIR)/$(BINARY_NAME)
GOROOT= $(shell $(GOCMD) env GOROOT)
RED=\033[31m
GREEN=\033[32m
YELLOW=\033[33m
END=\033[0m
comms=default_comms
db=default_db
mkfile_path := $(abspath $(lastword $(MAKEFILE_LIST)))
mkfile_dir := $(dir $(mkfile_path))

export CGO_ENABLED = 0 # This should build everything static, but if we use CGO, we need to use ldflags as above ^

.PHONY: build run clean

all: build run
build:
	@echo "GOPATH=$(GOPATH)"
	@go build -o $(mkfile_dir)dumper -v ./symbol_dumper
	CGO_ENABLED=0 GOOS="linux" GOARCH="amd64" go build -o $(mkfile_dir)ld ./loader
	@$(mkfile_dir)dumper $(mkfile_dir)ld
	GOOS=linux GOARCH=amd64 go build -o $(mkfile_dir)http.o -v ./http;
run:
	$(mkfile_dir)ld -o $(mkfile_dir)http.o -p example -run git.redteam.fb/malwareunicorn/memgoloader/example/http.main
clean:
	rm $(mkfile_dir)ld
	rm $(mkfile_dir)http.o
	rm $(mkfile_dir)dumper
	rm $(mkfile_dir)symbols
