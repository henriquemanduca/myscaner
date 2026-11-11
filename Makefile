BINARY_NAME := my-scaner
MODULE_PATH := my-scaner
BUILD_DIR := bin
ENTRYPOINT := cmd/my-scaner/main.go

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "v0.1.0-dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE ?= $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')

LDFLAGS := -w -s \
	-X '$(MODULE_PATH)/internal/buildinfo.Version=$(VERSION)' \
	-X '$(MODULE_PATH)/internal/buildinfo.Commit=$(COMMIT)' \
	-X '$(MODULE_PATH)/internal/buildinfo.Date=$(DATE)'

.PHONY: all build run test lint clean help

all: build

help: ## Exibe a lista de comandos disponíveis
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## Compila o binário aplicando ldflags de versão
	@mkdir -p $(BUILD_DIR)
	go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) $(ENTRYPOINT)
	@echo "Binário compilado em $(BUILD_DIR)/$(BINARY_NAME)"

run: ## Executa a CLI localmente (ex: make run ARGS="scan --host localhost")
	go run -ldflags "$(LDFLAGS)" $(ENTRYPOINT) $(ARGS)

test: ## Executa os testes unitários com detecção de data race
	go test -race -v ./...

clean: ## Remove binários e artefatos de build
	rm -rf $(BUILD_DIR)
