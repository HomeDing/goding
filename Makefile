# Makefile for goding - build, test, run the server

.DEFAULT_GOAL := build
BINARY := goding.exe
PKG := main.go

.PHONY: all build test run clean help

all: build

build:
	@echo building $(BINARY)...
	go build -v -o $(BINARY) $(PKG)

test:
	@echo testing...
	go test ./...

run:
	go run $(PKG)

clean:
	-rm -f $(BINARY)

help:
	@echo "Usage: make [target]"
	@echo "Targets:"
	@echo "  build  - build the binary (default)"
	@echo "  test   - run go tests"
	@echo "  run    - run with `go run`"
	@echo "  clean  - remove built binaries"
	@echo "  help   - show this help"