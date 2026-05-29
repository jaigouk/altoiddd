# White Hat Hacker — alto Go addenda

## Project language / runtime

- **Go 1.26+** with modules

## Go-specific scanning commands

```bash
# Go vulnerability check
govulncheck ./...

# Dependency audit
go list -m -json all | grep -i "CVE\|vulnerability"

# Static analysis security rules (Go)
golangci-lint run --enable gosec
```

## Hardcoded-secret grep (Go-specific include pattern)

```bash
grep -rn "password\|secret\|api.key\|token" --include="*.go" . | grep -v "_test.go" | grep -v "vendor/"
```
