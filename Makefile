.PHONY: build test lint vet fmt check release clean

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

# Run all quality gates (build + vet + test + lint)
check: build vet test lint

# Build release binaries
release:
	CGO_ENABLED=0 go build -ldflags="-s -w -X main.version=$$(git describe --tags --always --dirty)" -o bin/alty ./cmd/alty
	CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/alty-mcp ./cmd/alty-mcp

# Remove build artifacts
clean:
	rm -rf bin/
	go clean -cache
