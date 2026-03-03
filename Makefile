# LexScript Compiler — Makefile
# Requires: Go 1.24+

BINARY   := lexscript
BUILD_DIR := bin
SRC       := $(shell find . -name '*.go' -not -path './vendor/*')

.PHONY: all build run test clean deps fmt vet lint help \
        compile-rental compile-software parse-rental validate-rental

## ── Default ─────────────────────────────────────────────────────────────────

all: deps build

## ── Build ───────────────────────────────────────────────────────────────────

build: deps                                  ## Build the lexscript binary into bin/
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY) .
	@echo "built: $(BUILD_DIR)/$(BINARY)"

## ── Dependencies ────────────────────────────────────────────────────────────

deps:                                        ## Download and tidy Go modules
	go mod tidy

## ── Run examples ────────────────────────────────────────────────────────────

compile-rental: build                        ## Compile examples/rental.lexscript → examples/rental.md
	./$(BUILD_DIR)/$(BINARY) compile examples/rental.lexscript -o examples/rental.md

compile-software: build                      ## Compile examples/software_dev.lexscript
	./$(BUILD_DIR)/$(BINARY) compile examples/software_dev.lexscript -o examples/software_dev.md

parse-rental: build                          ## Dump AST JSON for rental.lexscript (debug)
	./$(BUILD_DIR)/$(BINARY) parse examples/rental.lexscript

validate-rental: build                       ## Run semantic checks on rental.lexscript (debug)
	./$(BUILD_DIR)/$(BINARY) validate examples/rental.lexscript

## ── Testing ─────────────────────────────────────────────────────────────────

test:                                        ## Run all unit tests
	go test ./... -v

test-race:                                   ## Run tests with race detector
	go test -race ./... -v

bench:                                       ## Run benchmark tests
	go test ./... -bench=. -benchmem

## ── Code quality ─────────────────────────────────────────────────────────────

fmt:                                         ## Format all Go source files
	go fmt ./...

vet:                                         ## Run go vet
	go vet ./...

lint: fmt vet                                ## Format, vet (add golangci-lint for full linting)
	@echo "lint complete"

## ── Cleanup ─────────────────────────────────────────────────────────────────

clean:                                       ## Remove built binary and generated outputs
	rm -rf $(BUILD_DIR)
	rm -f examples/*.md

## ── Help ────────────────────────────────────────────────────────────────────

help:                                        ## Show this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-22s\033[0m %s\n", $$1, $$2}'
