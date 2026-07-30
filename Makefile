# Makefile for goding - build, test, run the server

$(info )
$(info ===== Make of GoDing =====)
$(info )

.DEFAULT_GOAL := build
BINARY := goDing.exe
PKG := main.go

.PHONY: all build test run clean help

all: build

build:
	$(info building $(BINARY)...)
	go build -v -o $(BINARY) $(PKG)

test:
	$(info testing...)
	go test ./...

run:
	go run $(PKG)

clean:
	-rm -f $(BINARY)

help:
	$(info Usage: make [target])
	$(info )
	$(info Targets:)
	$(info   build  - build the binary (default))
	$(info   test   - run go tests)
	$(info   run    - run with `go run`)
	$(info   clean  - remove built binaries)
	$(info   help   - show this help)
	$(info )
