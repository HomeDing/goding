# Makefile for goding - build, test, run the server

BINARY := main.exe
PKG := main.go

.PHONY: all build test run clean help

all: build

build:
	go build -v -o $(BINARY) $(PKG)

test:
	go test ./...

run:
	go run $(PKG)

clean:
	-rm -f $(BINARY) $(BINARY).exe

help:
	@echo "Usage: make [target]"
	@echo "Targets:"
	@echo "  build  - build the server binary into bin/"
	@echo "  test   - run go tests"
	@echo "  run    - run the server with go run"
	@echo "  clean  - remove built binaries"
	@echo "  help   - show this help"
