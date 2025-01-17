.PHONY: all clean build build-all

# Binary name
BINARY=conf-sync

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

# Build for the current platform
build:
	echo "Building ${BINARY}..."
	CGO_ENABLED=0 go build -trimpath ${BUILD_TAGS} ${LDFLAGS} ${GCFLAGS} ${ASMFLAGS} \
		-o ${BUILD_DIR}/${BINARY} cmd/app/main.go

# Build for all platforms
build-all: clean
	echo "Building for all platforms..."
	mkdir -p ${BUILD_DIR}
	$(foreach PLATFORM,${PLATFORMS}, \
		$(eval GOOS=$(word 1,$(subst /, ,${PLATFORM}))) \
		$(eval GOARCH=$(word 2,$(subst /, ,${PLATFORM}))) \
		$(eval SUFFIX=$(if $(findstring windows,${GOOS}),.exe,)) \
		echo "Building for ${GOOS}/${GOARCH}..." && \
		CGO_ENABLED=0 GOOS=${GOOS} GOARCH=${GOARCH} go build -trimpath ${BUILD_TAGS} ${LDFLAGS} ${GCFLAGS} ${ASMFLAGS} \
			-o ${BUILD_DIR}/${BINARY}-${GOOS}-${GOARCH}${SUFFIX} cmd/app/main.go; \
	)

# Run tests
test:
	echo "Running tests..."
	go test -v ./...

# Install locally
install:
	echo "Installing ${BINARY}..."
	CGO_ENABLED=0 go install -trimpath ${BUILD_TAGS} ${LDFLAGS} ${GCFLAGS} ${ASMFLAGS} ./cmd/app

# Show help
help:
	echo "Available targets:"
	echo "  build       - Build for current platform"
	echo "  build-all   - Build for all platforms"
	echo "  clean       - Clean build directory"
	echo "  test        - Run tests"
	echo "  install     - Install locally"
	echo ""
	echo "Supported platforms: ${PLATFORMS}"
	echo ""
	echo "Examples:"
	echo "  make build"
	echo "  make build-all"
	echo "  make VERSION=2.0.0 build-all"
