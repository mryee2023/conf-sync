.PHONY: all clean build build-all docker-build docker-push

# Binary names
CLIENT_BINARY=conf-sync-client
SERVER_BINARY=conf-sync-server

# Docker image names
DOCKER_REGISTRY?=docker.io
DOCKER_NAMESPACE?=mryee2023
CLIENT_IMAGE=$(DOCKER_REGISTRY)/$(DOCKER_NAMESPACE)/conf-sync-client
SERVER_IMAGE=$(DOCKER_REGISTRY)/$(DOCKER_NAMESPACE)/conf-sync-server

# Build directory
BUILD_DIR=build

# Version (you can override this using make VERSION=x.x.x)
VERSION?=1.0.0

# Get the current commit hash or use "dev" if not in a git repo
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "dev")

# Build time
BUILD_TIME=$(shell date +%FT%T%z)

# Go build flags
LDFLAGS=-ldflags "-s -w -X main.Version=${VERSION} -X main.CommitHash=${COMMIT} -X main.BuildTime=${BUILD_TIME}"
GCFLAGS=-gcflags="all=-trimpath=${PWD}"
ASMFLAGS=-asmflags="all=-trimpath=${PWD}"

# Build tags for optimization
BUILD_TAGS=-tags 'netgo osusergo static_build'

# Supported platforms
PLATFORMS=darwin/amd64 darwin/arm64 linux/amd64 linux/arm64 windows/amd64 windows/arm64

# Default target
all: clean build-all

# Clean build directory
clean:
	rm -rf ${BUILD_DIR}

# Build both client and server for the current platform
build: build-client build-server

# Build client for the current platform
build-client:
	echo "Building ${CLIENT_BINARY}..."
	CGO_ENABLED=0 go build -trimpath ${BUILD_TAGS} ${LDFLAGS} ${GCFLAGS} ${ASMFLAGS} \
		-o ${BUILD_DIR}/${CLIENT_BINARY} cmd/client/main.go

# Build server for the current platform
build-server:
	echo "Building ${SERVER_BINARY}..."
	CGO_ENABLED=0 go build -trimpath ${BUILD_TAGS} ${LDFLAGS} ${GCFLAGS} ${ASMFLAGS} \
		-o ${BUILD_DIR}/${SERVER_BINARY} cmd/server/main.go

# Build for all platforms
build-all: clean
	echo "Building for all platforms..."
	mkdir -p ${BUILD_DIR}
	$(foreach PLATFORM,${PLATFORMS}, \
		$(eval GOOS=$(word 1,$(subst /, ,${PLATFORM}))) \
		$(eval GOARCH=$(word 2,$(subst /, ,${PLATFORM}))) \
		$(eval SUFFIX=$(if $(findstring windows,${GOOS}),.exe,)) \
		echo "Building client for ${GOOS}/${GOARCH}..." && \
		CGO_ENABLED=0 GOOS=${GOOS} GOARCH=${GOARCH} go build -trimpath ${BUILD_TAGS} ${LDFLAGS} ${GCFLAGS} ${ASMFLAGS} \
			-o ${BUILD_DIR}/${CLIENT_BINARY}-${GOOS}-${GOARCH}${SUFFIX} cmd/client/main.go && \
		echo "Building server for ${GOOS}/${GOARCH}..." && \
		CGO_ENABLED=0 GOOS=${GOOS} GOARCH=${GOARCH} go build -trimpath ${BUILD_TAGS} ${LDFLAGS} ${GCFLAGS} ${ASMFLAGS} \
			-o ${BUILD_DIR}/${SERVER_BINARY}-${GOOS}-${GOARCH}${SUFFIX} cmd/server/main.go; \
	)

# Run tests
test:
	go test -v ./...

# Install locally
install:
	echo "Installing ${CLIENT_BINARY} and ${SERVER_BINARY}..."
	CGO_ENABLED=0 go install -trimpath ${BUILD_TAGS} ${LDFLAGS} ${GCFLAGS} ${ASMFLAGS} ./cmd/client
	CGO_ENABLED=0 go install -trimpath ${BUILD_TAGS} ${LDFLAGS} ${GCFLAGS} ${ASMFLAGS} ./cmd/server

# Build Docker images
docker-build:
	echo "Building Docker images..."
	docker build -t ${CLIENT_IMAGE}:${VERSION} -f docker/client/Dockerfile .
	docker build -t ${SERVER_IMAGE}:${VERSION} -f docker/server/Dockerfile .
	docker tag ${CLIENT_IMAGE}:${VERSION} ${CLIENT_IMAGE}:latest
	docker tag ${SERVER_IMAGE}:${VERSION} ${SERVER_IMAGE}:latest

# Push Docker images
docker-push:
	echo "Pushing Docker images..."
	docker push ${CLIENT_IMAGE}:${VERSION}
	docker push ${SERVER_IMAGE}:${VERSION}
	docker push ${CLIENT_IMAGE}:latest
	docker push ${SERVER_IMAGE}:latest

# Show help
help:
	@echo "Available targets:"
	@echo "  build        - Build both client and server for current platform"
	@echo "  build-all    - Build for all supported platforms"
	@echo "  clean        - Clean build directory"
	@echo "  test         - Run tests"
	@echo "  install      - Install binaries locally"
	@echo "  docker-build - Build Docker images"
	@echo "  docker-push  - Push Docker images"
	@echo ""
	@echo "Environment variables:"
	@echo "  VERSION            - Version tag (default: ${VERSION})"
	@echo "  DOCKER_REGISTRY    - Docker registry (default: ${DOCKER_REGISTRY})"
	@echo "  DOCKER_NAMESPACE   - Docker namespace (default: ${DOCKER_NAMESPACE})"
