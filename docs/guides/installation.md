---
title: Installation
description: Install alto and its prerequisites on your system
sidebar:
  order: 2
---

## Prerequisites

| Requirement | Version | Purpose |
|-------------|---------|---------|
| Git | any | alto uses git for branch-based scaffolding |
| Beads | latest | Git-native issue tracking (`bd` CLI) |

### Installing Beads

Beads is a git-native issue tracker that stores tickets in `.beads/issues.jsonl` inside your repo. Install it following the beads documentation. Verify with:

```bash
bd --version
```

Beads is optional for basic project bootstrapping but required for ticket generation and the ripple review workflow.

## Install alto

### Download a release binary (recommended)

Download the latest binary for your platform from the [releases page](https://github.com/jaigouk/altoiddd/releases):

```bash
# macOS (Apple Silicon)
curl -L -o /usr/local/bin/alto https://github.com/jaigouk/altoiddd/releases/latest/download/alto-darwin-arm64
chmod +x /usr/local/bin/alto

# macOS (Intel)
curl -L -o /usr/local/bin/alto https://github.com/jaigouk/altoiddd/releases/latest/download/alto-darwin-amd64
chmod +x /usr/local/bin/alto

# Linux (amd64)
curl -L -o /usr/local/bin/alto https://github.com/jaigouk/altoiddd/releases/latest/download/alto-linux-amd64
chmod +x /usr/local/bin/alto
```

On Windows, download `alto-windows-amd64.exe` from the [releases page](https://github.com/jaigouk/altoiddd/releases) and add it to your `PATH`.

### From source

Requires Go 1.26+.

```bash
go install github.com/jaigouk/altoiddd/cmd/alto@latest
```

This places the `alto` binary in your `$GOPATH/bin` (typically `~/go/bin`). Make sure that directory is in your `PATH`:

```bash
export PATH="$PATH:$(go env GOPATH)/bin"
```

### Verify installation

```bash
alto version
```

## Optional tools

alto can detect and integrate with these tools during `alto init`:

| Tool | Purpose |
|------|---------|
| [Claude Code](https://claude.ai/claude-code) | AI coding assistant (CLI) |
| [Cursor](https://cursor.sh) | AI-powered IDE |
| [Roo Code](https://roocode.com) | AI coding assistant |
| [OpenCode](https://opencode.ai) | AI coding assistant |
| [golangci-lint](https://golangci-lint.run) | Go meta-linter for quality gates |
| [Trivy](https://trivy.dev) | Security vulnerability scanner |

alto detects which of these are installed and generates appropriate configuration files. You don't need all of them — just the ones you use.

## Upgrading

Download the latest binary from the [releases page](https://github.com/jaigouk/altoiddd/releases), or if installed from source:

```bash
go install github.com/jaigouk/altoiddd/cmd/alto@latest
```

Check the current version:

```bash
alto version
```
