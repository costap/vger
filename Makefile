# Variables
APP_NAME    := vger
BIN_DIR     := bin
BINARY      := $(BIN_DIR)/$(APP_NAME)
MAIN_PKG    := ./cmd/vger

# Go toolchain
GOCMD       := go
GOBUILD     := $(GOCMD) build
GOTEST      := $(GOCMD) test
GOVET       := $(GOCMD) vet
GOFMT       := gofmt
GOMOD       := $(GOCMD) mod

# Git metadata for build stamping
VERSION     ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT      ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE        ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS     := -s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)

.PHONY: help all build run test test-cover vet fmt fmt-check tidy verify ci clean deps outdated vuln

## help: Display this help message (default)
help:
	@echo "V'Ger Build Targets:"
	@grep -E '^## [a-zA-Z_-]+:' $(MAKEFILE_LIST) | sed -e 's/## //' | awk -F': ' '{printf "  \033[36m%-14s\033[0m %s\n", $$1, $$2}'

## all: Format, vet, test, and build binary
all: fmt vet test build

## build: Build the vger binary into bin/vger
build:
	@mkdir -p $(BIN_DIR)
	$(GOBUILD) -ldflags "$(LDFLAGS)" -o $(BINARY) $(MAIN_PKG)

## run: Run vger with optional ARGS="..."
run:
	$(GOCMD) run $(MAIN_PKG) $(ARGS)

## test: Run unit tests
test:
	$(GOTEST) -v ./...

## test-cover: Run unit tests and generate coverage report
test-cover:
	@mkdir -p $(BIN_DIR)
	$(GOTEST) -v -coverprofile=$(BIN_DIR)/coverage.out -covermode=atomic ./...
	$(GOCMD) tool cover -func=$(BIN_DIR)/coverage.out

## vet: Run go vet static analysis
vet:
	$(GOVET) ./...

## fmt: Format Go source files using gofmt
fmt:
	$(GOFMT) -w -s .

## fmt-check: Check formatting without modifying files
fmt-check:
	@test -z "$$($(GOFMT) -l .)" || (echo "Unformatted files found. Run 'make fmt':" && $(GOFMT) -l . && exit 1)

## tidy: Run go mod tidy
tidy:
	$(GOMOD) tidy

## verify: Run go mod verify
verify:
	$(GOMOD) verify

## deps: List all direct dependencies and current versions
deps:
	@$(GOCMD) list -u -m -json all | $(GOCMD) tool go-mod-outdated -direct

## outdated: List direct dependencies that have newer versions available
outdated:
	@OUTPUT=$$($(GOCMD) list -u -m -json all | $(GOCMD) tool go-mod-outdated -update -direct); \
	if [ -z "$$OUTPUT" ]; then \
		echo "All direct dependencies are up to date!"; \
	else \
		echo "$$OUTPUT"; \
	fi

## vuln: Run govulncheck vulnerability scanner
vuln:
	@$(GOCMD) tool govulncheck ./...

## ci: Full CI verification suite (tidy, verify, fmt-check, vet, test, build)
ci: tidy verify fmt-check vet test build

## clean: Remove compiled binaries and test coverage files
clean:
	rm -rf $(BIN_DIR) vger

## hooks: Install git hooks from .githooks/ directory
hooks:
	@git config core.hooksPath .githooks
	@chmod +x .githooks/*
	@echo "Git hooks configured from .githooks/"