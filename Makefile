NAME_TRANSLATOR := translator
NAME_MACHINE := machine
BIN_DIR := bin
VERSION := 0.1.0
GOFLAGS := -ldflags="-s -w -X main.version=$(VERSION)"
# GOFLAGS := 

PLATFORMS := linux windows darwin
ARCHS := amd64 arm64

.PHONY: all
all: test build

.PHONY: build
build: build-translator build-machine   

.PHONY: build-translator
build-translator:
	@echo "Building $(NAME_TRANSLATOR) for current platform..."
	@mkdir -p $(BIN_DIR)
	go build $(GOFLAGS) -o $(BIN_DIR)/$(NAME_TRANSLATOR) ./cmd/$(NAME_TRANSLATOR)

.PHONY: build-machine
build-machine:
	@echo "Building $(NAME_MACHINE) for current platform..."
	@mkdir -p $(BIN_DIR)
	go build $(GOFLAGS) -o $(BIN_DIR)/$(NAME_MACHINE) ./cmd/$(NAME_MACHINE)

.PHONY: release
release: test build-translator build-machine

.PHONY: test
test:        ## go test ./...
	@echo "Running tests…"
	go test -v ./...

.PHONY: ci
ci:
	./check-ci.sh

.PHONY: clean
clean:
	@echo "Cleaning..."
	@rm -rf $(BIN_DIR)


.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build               - Build both executables for current platform"
	@echo "  build-translator    - Build only translator for current platform"
	@echo "  build-machine       - Build only machine for current platform"
	@echo "  release             - Cross-compile both executables for all platforms"
	@echo "  clean               - Remove build artifacts"
	@echo "  help                - Show this help message"
