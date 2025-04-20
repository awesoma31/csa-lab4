NAME_TRANSLATOR := translator
NAME_MACHINE := machine
BIN_DIR := bin
VERSION := 0.1.0
GOFLAGS := -ldflags="-s -w -X main.version=$(VERSION)"

PLATFORMS := linux windows darwin
ARCHS := amd64 arm64

.PHONY: all
all: build

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
release: release-translator release-machine

.PHONY: release-translator
release-translator:
	@echo "Building $(NAME_TRANSLATOR) for all platforms..."
	@$(foreach GOOS, $(PLATFORMS),\
		$(foreach GOARCH, $(ARCHS),\
			$(shell \
				export GOOS=$(GOOS);\
				export GOARCH=$(GOARCH);\
				BINARY=$(BIN_DIR)/$(GOOS)-$(GOARCH)/$(NAME_TRANSLATOR);\
				if [ "$(GOOS)" = "windows" ]; then BINARY=$$BINARY.exe; fi;\
				mkdir -p $$(dirname $$BINARY);\
				echo "Building $$BINARY";\
				go build $(GOFLAGS) -o $$BINARY ./cmd/$(NAME_TRANSLATOR);\
			)\
		)\
	)

.PHONY: release-machine
release-machine:
	@echo "Building $(NAME_MACHINE) for all platforms..."
	@$(foreach GOOS, $(PLATFORMS),\
		$(foreach GOARCH, $(ARCHS),\
			$(shell \
				export GOOS=$(GOOS);\
				export GOARCH=$(GOARCH);\
				BINARY=$(BIN_DIR)/$(GOOS)-$(GOARCH)/$(NAME_MACHINE);\
				if [ "$(GOOS)" = "windows" ]; then BINARY=$$BINARY.exe; fi;\
				mkdir -p $$(dirname $$BINARY);\
				echo "Building $$BINARY";\
				go build $(GOFLAGS) -o $$BINARY ./cmd/$(NAME_MACHINE);\
			)\
		)\
	)

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
	@echo "  release-translator  - Cross-compile only translator for all platforms"
	@echo "  release-machine     - Cross-compile only machine for all platforms"
	@echo "  clean               - Remove build artifacts"
	@echo "  help                - Show this help message"
