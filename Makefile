# LexScript Compiler — Makefile
# Requires: Go 1.24+

BINARY   := lexs
BUILD_DIR := bin
SRC       := $(shell find . -name '*.go' -not -path './vendor/*')

.PHONY: all build run test clean deps fmt vet lint help \
        compile-rental compile-software parse-rental validate-rental \
        compile-rental-pdf compile-software-pdf \
        fmt-rental fmt-software visualize-rental visualize-software \
        compile-employment compile-employment-pdf \
        compile-employment-de compile-employment-ca compile-employment-uk \
        validate-employment

## ── Default ─────────────────────────────────────────────────────────────────

all: deps build

## ── Build ───────────────────────────────────────────────────────────────────

build: deps                                  ## Build the lexs binary into bin/
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY) .
	@echo "built: $(BUILD_DIR)/$(BINARY)"

## ── Dependencies ────────────────────────────────────────────────────────────

deps:                                        ## Download and tidy Go modules
	go mod tidy

## ── Run examples ────────────────────────────────────────────────────────────

compile-rental: build                        ## Compile examples/rental.lxs → examples/rental.md
	./$(BUILD_DIR)/$(BINARY) compile examples/rental.lxs -o examples/rental.md

compile-software: build                      ## Compile examples/software_dev.lxs
	./$(BUILD_DIR)/$(BINARY) compile examples/software_dev.lxs -o examples/software_dev.md

parse-rental: build                          ## Dump AST JSON for rental.lxs (debug)
	./$(BUILD_DIR)/$(BINARY) parse examples/rental.lxs

validate-rental: build                       ## Run semantic checks on rental.lxs (debug)
	./$(BUILD_DIR)/$(BINARY) validate examples/rental.lxs

## ── Phase 2 commands ─────────────────────────────────────────────────────────

compile-rental-pdf: build                    ## Compile examples/rental.lxs → examples/rental.pdf (Phase 2)
	./$(BUILD_DIR)/$(BINARY) compile examples/rental.lxs -f pdf -o examples/rental.pdf

compile-software-pdf: build                  ## Compile examples/software_dev.lxs → PDF (Phase 2)
	./$(BUILD_DIR)/$(BINARY) compile examples/software_dev.lxs -f pdf -o examples/software_dev.pdf

fmt-rental: build                            ## Format examples/rental.lxs in-place (lexs fmt, Phase 2)
	./$(BUILD_DIR)/$(BINARY) fmt --write examples/rental.lxs

fmt-software: build                          ## Format examples/software_dev.lxs in-place (Phase 2)
	./$(BUILD_DIR)/$(BINARY) fmt --write examples/software_dev.lxs

visualize-rental: build                      ## Export rental.lxs FSM → examples/rental.dot (Phase 2)
	./$(BUILD_DIR)/$(BINARY) visualize examples/rental.lxs -o examples/rental.dot

visualize-software: build                    ## Export software_dev.lxs FSM → .dot (Phase 2)
	./$(BUILD_DIR)/$(BINARY) visualize examples/software_dev.lxs -o examples/software_dev.dot

## ── Phase 3 commands ─────────────────────────────────────────────────────────

compile-employment: build                    ## Compile examples/employment.lxs → MD (Phase 3)
	./$(BUILD_DIR)/$(BINARY) compile examples/employment.lxs -o examples/employment.md

compile-employment-pdf: build                ## Compile examples/employment.lxs → PDF (Phase 3)
	./$(BUILD_DIR)/$(BINARY) compile examples/employment.lxs -f pdf -o examples/employment.pdf

compile-employment-de: build                 ## employment.lxs → MD with Delaware jurisdiction (Phase 3)
	./$(BUILD_DIR)/$(BINARY) compile examples/employment.lxs -j delaware -o examples/employment_delaware.md

compile-employment-ca: build                 ## employment.lxs → MD with California jurisdiction (Phase 3)
	./$(BUILD_DIR)/$(BINARY) compile examples/employment.lxs -j california -o examples/employment_california.md

compile-employment-uk: build                 ## employment.lxs → MD with UK jurisdiction (Phase 3)
	./$(BUILD_DIR)/$(BINARY) compile examples/employment.lxs -j uk -o examples/employment_uk.md

validate-employment: build                   ## Validate examples/employment.lxs (Phase 3)
	./$(BUILD_DIR)/$(BINARY) validate examples/employment.lxs

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
	rm -f examples/*.md examples/*.pdf examples/*.dot

## ── Help ────────────────────────────────────────────────────────────────────

help:                                        ## Show this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-22s\033[0m %s\n", $$1, $$2}'
