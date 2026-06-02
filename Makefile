.PHONY: build test smoke lint vet fmt deadcode check ci preflight install-hooks ci-local release release-all clean

# Version injection
VERSION_PKG := github.com/alto-cli/alto/internal/composition
VERSION := $(shell git describe --tags --always --dirty)
LDFLAGS := -s -w -X $(VERSION_PKG).Version=$(VERSION)

# Cross-compilation targets
PLATFORMS := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64

# Build binaries
build:
	go build ./cmd/alto
	go build ./cmd/alto-mcp

# Run all tests with race detector
test:
	go test ./... -v -race -count=1

# Run smoke tests (requires building binary first via TestMain)
smoke:
	go test -tags smoke -v -timeout 60s ./cmd/alto/

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
	CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o bin/alto ./cmd/alto
	CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/alto-mcp ./cmd/alto-mcp

# Build release binaries for all platforms (5 platforms × 2 binaries = 10 total)
release-all:
	@for platform in $(PLATFORMS); do \
		GOOS=$$(echo $$platform | cut -d/ -f1); \
		GOARCH=$$(echo $$platform | cut -d/ -f2); \
		ext=""; \
		if [ "$$GOOS" = "windows" ]; then ext=".exe"; fi; \
		echo "Building alto-$$GOOS-$$GOARCH$$ext"; \
		CGO_ENABLED=0 GOOS=$$GOOS GOARCH=$$GOARCH go build -ldflags="$(LDFLAGS)" -o bin/alto-$$GOOS-$$GOARCH$$ext ./cmd/alto; \
		echo "Building alto-mcp-$$GOOS-$$GOARCH$$ext"; \
		CGO_ENABLED=0 GOOS=$$GOOS GOARCH=$$GOARCH go build -ldflags="-s -w" -o bin/alto-mcp-$$GOOS-$$GOARCH$$ext ./cmd/alto-mcp; \
	done

# CI target (alias for check)
ci: check

# Local CI-parity gate. Same surface as .githooks/pre-push.
# Runs `bd preflight --check` which mirrors .gitea/workflows/ci.yaml
# (tests, lint, gofmt, beads-pollution, nix hash, AGENTS.md sync).
preflight:
	bd preflight --check

# Wire the tracked .githooks/ as this clone's hooks directory.
# Run once per fresh clone. Idempotent.
install-hooks:
	@current="$$(git config core.hooksPath || echo '<unset>')"; \
	if [ "$$current" != ".githooks" ]; then \
		git config core.hooksPath .githooks; \
		echo "✓ core.hooksPath set to .githooks (was: $$current)"; \
	else \
		echo "✓ core.hooksPath already .githooks"; \
	fi
	@chmod +x .githooks/* 2>/dev/null || true
	@echo "✓ hooks installed: $$(ls .githooks/)"

# Full local CI parity (preflight + trivy, matching the CI security job).
# Use before risky pushes; trivy adds ~30-90s on cold cache.
ci-local: preflight
	@if command -v trivy >/dev/null; then \
		echo "→ ci-local: running trivy vuln scan..."; \
		trivy fs . --scanners vuln --severity CRITICAL,HIGH --ignore-unfixed --exit-code 1 --format table; \
		echo "→ ci-local: running trivy secret scan..."; \
		trivy fs . --scanners secret --severity CRITICAL,HIGH --exit-code 1 --format table; \
	else \
		echo "⚠ ci-local: trivy not installed; skipping security scans (CI will run them)"; \
	fi

# Remove build artifacts
clean:
	rm -rf bin/
	go clean -cache
