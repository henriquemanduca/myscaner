APP_NAME  := scaner-cli
MODULE    := scaner-cli
CMD_PATH  := ./cmd/scaner-cli
BIN_DIR   ?= bin
PREFIX    ?= $(HOME)/.local

GO        ?= go

VERSION   ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT    ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE      ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

LDFLAGS   := -s -w \
	-X '$(MODULE)/internal/buildinfo.Version=$(VERSION)' \
	-X '$(MODULE)/internal/buildinfo.Commit=$(COMMIT)' \
	-X '$(MODULE)/internal/buildinfo.Date=$(DATE)'
BUILD     := CGO_ENABLED=0 $(GO) build -trimpath -ldflags "$(LDFLAGS)"

.DEFAULT_GOAL := help
.PHONY: help build build-all checksums run test test-cover fmt vet lint lint-fix check clean install uninstall

help: ## Show available targets
	@awk 'BEGIN {FS = ":.*## "; printf "\nUsage:\n  make \033[36m<target>\033[0m\n\nTargets:\n"} /^[a-zA-Z_-]+:.*## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the binary into bin/scaner-cli
	@mkdir -p $(BIN_DIR)
	$(BUILD) -o $(BIN_DIR)/$(APP_NAME) $(CMD_PATH)
	@echo "Built $(BIN_DIR)/$(APP_NAME)"

build-all: clean ## Cross-compile release binaries into bin/
	@mkdir -p $(BIN_DIR)
	GOOS=linux   GOARCH=amd64 $(BUILD) -o $(BIN_DIR)/$(APP_NAME)-linux-amd64 $(CMD_PATH)
	GOOS=linux   GOARCH=arm64 $(BUILD) -o $(BIN_DIR)/$(APP_NAME)-linux-arm64 $(CMD_PATH)
	GOOS=linux   GOARCH=arm GOARM=7 $(BUILD) -o $(BIN_DIR)/$(APP_NAME)-linux-armv7 $(CMD_PATH)
	GOOS=darwin  GOARCH=amd64 $(BUILD) -o $(BIN_DIR)/$(APP_NAME)-darwin-amd64 $(CMD_PATH)
	GOOS=darwin  GOARCH=arm64 $(BUILD) -o $(BIN_DIR)/$(APP_NAME)-darwin-arm64 $(CMD_PATH)
	GOOS=windows GOARCH=amd64 $(BUILD) -o $(BIN_DIR)/$(APP_NAME)-windows-amd64.exe $(CMD_PATH)

checksums: ## Write SHA-256 checksums for release binaries
	cd $(BIN_DIR) && sha256sum $(APP_NAME)-* > checksums.txt

run: ## Run from source (e.g. make run ARGS="-p ./documento.pdf")
	$(GO) run $(CMD_PATH) $(ARGS)

test: ## Run tests with the race detector
	$(GO) test -race -count=1 ./...

test-cover: ## Run tests and write coverage.out/coverage.html
	$(GO) test -race -count=1 -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html

fmt: ## Format the code and run go vet
	$(GO) fmt ./...
	$(GO) vet ./...

vet: ## Run go vet
	$(GO) vet ./...

lint: ## Run golangci-lint
	golangci-lint run ./...

lint-fix: ## Apply golangci-lint fixes
	golangci-lint run --fix ./...

check: fmt test ## Run the full local verification gate

clean: ## Remove build and coverage artifacts
	rm -rf $(BIN_DIR) coverage.out coverage.html

install: build ## Install the binary via install.sh
	@./install.sh $(ARGS)

uninstall: ## Uninstall the binary via install.sh
	@./install.sh --uninstall $(ARGS)
