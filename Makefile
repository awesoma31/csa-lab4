NAME_TRANSLATOR := translator
NAME_MACHINE := machine
NAME_WEB := web
BIN_DIR := bin
VERSION := 1.1.0
GOFLAGS := -ldflags="-s -w -X main.version=$(VERSION)"
# GOFLAGS :=

PLATFORMS := linux windows darwin
ARCHS := amd64 arm64

# Docker specific variables
DOCKER_IMAGE_WEB := csa-lab4-$(NAME_WEB)
DOCKER_TAG := $(VERSION)
DOCKER_CONTAINER_WEB := csa-lab4-$(NAME_WEB)-container
DOCKER_PORT_WEB := 8080

.PHONY: all
all: test build

.PHONY: build
build: build-translator build-machine build-web


.PHONY: build-web
build-web:
	@echo "Building $(NAME_WEB) for current platform..."
	@mkdir -p $(BIN_DIR)
	go build $(GOFLAGS) -o $(BIN_DIR)/$(NAME_WEB) ./cmd/$(NAME_WEB)

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

# Docker Targets
.PHONY: docker-build-web
docker-build-web: ## Build the Docker image for the web application
	@echo "Building Docker image for $(NAME_WEB)..."
	sudo docker build --build-arg VERSION=$(VERSION) -t $(DOCKER_IMAGE_WEB):$(DOCKER_TAG) .

.PHONY: docker-run-web
docker-run-web: docker-build-web ## Run the web application in a Docker container
	@echo "Running $(NAME_WEB) in Docker container..."
	# Stop and remove any existing container with the same name
	sudo docker stop $(DOCKER_CONTAINER_WEB) 2>/dev/null || true
	sudo docker rm $(DOCKER_CONTAINER_WEB) 2>/dev/null || true
	# REMOVE --env-file app.env if you are copying .env into the image directly
	sudo docker run -d \
		-p $(DOCKER_PORT_WEB):$(DOCKER_PORT_WEB) \
		--name $(DOCKER_CONTAINER_WEB) \
		$(DOCKER_IMAGE_WEB):$(DOCKER_TAG)

.PHONY: docker-stop-web
docker-stop-web: ## Stop the web application Docker container
	@echo "Stopping $(NAME_WEB) Docker container..."
	sudo docker stop $(DOCKER_CONTAINER_WEB)

.PHONY: docker-rm-web
docker-rm-web: ## Remove the web application Docker container
	@echo "Removing $(NAME_WEB) Docker container..."
	sudo docker rm $(DOCKER_CONTAINER_WEB)

.PHONY: docker-clean-web
docker-clean-web: docker-stop-web docker-rm-web ## Stop, remove web container, and prune its image
	@echo "Cleaning up $(NAME_WEB) Docker image and container..."
	sudo docker rmi $(DOCKER_IMAGE_WEB):$(DOCKER_TAG) 2>/dev/null || true # Remove the specific image tag
	sudo docker image prune -f # Remove any dangling images (including old layers)


.PHONY: test
test:        ## go test ./...
	@echo "Running tests…"
	go test -v ./...

.PHONY: ci
ci:
	./check-ci.sh

.PHONY: clean
clean:
	@echo "Cleaning native binaries..."
	@rm -rf $(BIN_DIR)
	@echo "Cleaning Docker containers and images (if any remaining)..."
	sudo docker stop $(DOCKER_CONTAINER_WEB) 2>/dev/null || true
	sudo docker rm $(DOCKER_CONTAINER_WEB) 2>/dev/null || true
	sudo docker image prune -f # Cleans dangling images
	sudo docker builder prune -f # Cleans build cache

.PHONY: help
help:
	@echo "Available targets:"
	@echo "  all                 - Run tests and build all native binaries"
	@echo "  build               - Build all native executables for current platform"
	@echo "  build-translator    - Build only translator for current platform"
	@echo "  build-machine       - Build only machine for current platform"
	@echo "  build-web           - Build only web for current platform"
	@echo "  docker-build-web    - Build the Docker image for the web application"
	@echo "  docker-run-web      - Run the web application in a Docker container"
	@echo "  docker-stop-web     - Stop the web application Docker container"
	@echo "  docker-rm-web       - Remove the web application Docker container"
	@echo "  docker-clean-web    - Stop, remove web container, and prune its image"
	@echo "  test                - Run Go tests"
	@echo "  ci                  - Run CI checks"
	@echo "  clean               - Remove build artifacts and some Docker leftovers"
	@echo "  help                - Show this help message"
