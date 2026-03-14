---
title: Installation
description: Install alty and its prerequisites on your system
sidebar:
  order: 2
---

## Prerequisites

| Requirement | Version | Purpose |
|-------------|---------|---------|
| Go | 1.26+ | Runtime and build tool |
| Git | any | alty uses git for branch-based scaffolding |
| Beads | latest | Git-native issue tracking (`bd` CLI) |

### Installing Go

Follow the official instructions at [go.dev/dl](https://go.dev/dl/). Verify with:

```bash
go version
# go version go1.26.x ...
```

### Installing Beads

Beads is a git-native issue tracker that stores tickets in `.beads/issues.jsonl` inside your repo. Install it following the beads documentation. Verify with:

```bash
bd --version
```

Beads is optional for basic project bootstrapping but required for ticket generation and the ripple review workflow.

## Install alty

### From source (recommended)

```bash
go install github.com/alty-cli/alty/cmd/alty@latest
```

This places the `alty` binary in your `$GOPATH/bin` (typically `~/go/bin`). Make sure that directory is in your `PATH`:

```bash
export PATH="$PATH:$(go env GOPATH)/bin"
```

### From a release binary

Download the latest binary for your platform from the releases page and place it in a directory on your `PATH`:

```bash
# Example for Linux amd64
curl -L -o /usr/local/bin/alty <release-url>/alty-linux-amd64
chmod +x /usr/local/bin/alty
```

### Verify installation

```bash
alty version
```

## Optional tools

alty can detect and integrate with these tools during `alty init`:

| Tool | Purpose |
|------|---------|
| [Claude Code](https://claude.ai/claude-code) | AI coding assistant (CLI) |
| [Cursor](https://cursor.sh) | AI-powered IDE |
| [Roo Code](https://roocode.com) | AI coding assistant |
| [OpenCode](https://opencode.ai) | AI coding assistant |
| [golangci-lint](https://golangci-lint.run) | Go meta-linter for quality gates |
| [Trivy](https://trivy.dev) | Security vulnerability scanner |

alty detects which of these are installed and generates appropriate configuration files. You don't need all of them — just the ones you use.

## Upgrading

```bash
go install github.com/alty-cli/alty/cmd/alty@latest
```

Check the current version:

```bash
alty version
```
