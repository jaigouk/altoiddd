.PHONY: build test lint vet fmt deadcode check ci release release-all clean

# Version injection
VERSION_PKG := github.com/alty-cli/alty/internal/composition
VERSION := $(shell git describe --tags --always --dirty)
LDFLAGS := -s -w -X $(VERSION_PKG).Version=$(VERSION)

# Cross-compilation targets
PLATFORMS := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64

# Build binaries
build:
	go build ./cmd/alty
	go build ./cmd/alty-mcp

# Run all tests with race detector
test:
	go test ./... -v -race -count=1

# Run golangci-lint v2
lint:
	golangci-lint run

# Run go vet
vet:
	go vet ./...

# Format code with gofumpt
fmt:
	gofumpt -w .

# Detect dead code (production only, from main entry points)
# Pinned version requires Go 1.25+ (deadcode@latest built with Go 1.24)
DEADCODE_VERSION := v0.42.1-0.20260306220548-ff454944261a
deadcode:
	go run golang.org/x/tools/cmd/deadcode@$(DEADCODE_VERSION) ./cmd/...

# Run all quality gates (build + vet + test + lint + deadcode)
check: build vet test lint deadcode

# Build release binaries (current platform)
release:
	CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o bin/alty ./cmd/alty
	CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/alty-mcp ./cmd/alty-mcp

# Build release binaries for all platforms (5 platforms × 2 binaries = 10 total)
release-all:
	@for platform in $(PLATFORMS); do \
		GOOS=$$(echo $$platform | cut -d/ -f1); \
		GOARCH=$$(echo $$platform | cut -d/ -f2); \
		ext=""; \
		if [ "$$GOOS" = "windows" ]; then ext=".exe"; fi; \
		echo "Building alty-$$GOOS-$$GOARCH$$ext"; \
		CGO_ENABLED=0 GOOS=$$GOOS GOARCH=$$GOARCH go build -ldflags="$(LDFLAGS)" -o bin/alty-$$GOOS-$$GOARCH$$ext ./cmd/alty; \
		echo "Building alty-mcp-$$GOOS-$$GOARCH$$ext"; \
		CGO_ENABLED=0 GOOS=$$GOOS GOARCH=$$GOARCH go build -ldflags="-s -w" -o bin/alty-mcp-$$GOOS-$$GOARCH$$ext ./cmd/alty-mcp; \
	done

# CI target (alias for check)
ci: check

# Remove build artifacts
clean:
	rm -rf bin/
	go clean -cache
