# alty-cli

**Your AI builds apps fast. alty makes sure they don't fall apart.**

---

## The Problem Everyone Has

AI coding tools (Cursor, Claude Code, OpenCode, Roo Code) are amazing at writing code. You type "build me an invoice app" and get a working prototype in minutes.

But here's what nobody tells you: **that prototype becomes unmaintainable within weeks.**

Why? Because the AI jumped straight to writing code without understanding:

- What your business actually does
- Which parts are complex and which are simple
- Where the boundaries should be so changes don't break everything

And even when you _do_ plan, **your plans go stale.** You finish researching one piece and the findings change everything else — but nobody updates the other tasks. Your team (human or AI) starts work based on outdated context, and the problems compound.

The result: you ship fast, then spend months fixing things — or throw it away and start over.

## What alty Does

alty is the **planning step that happens before coding starts.** Think of it as hiring a senior architect who:

1. **Listens to your idea** — You describe what you want in 4-5 plain sentences
2. **Asks the right questions** — Not "what framework?" but "what does your business do? what are the rules? what changes often?"
3. **Draws the blueprint** — Which parts of your app should be separate, how they connect, what the rules are
4. **Creates a build plan** — Ordered tasks that tell your AI tool exactly what to build, in what order, with tests already defined
5. **Sets up guardrails** — Automated checks that catch mistakes before they become problems

Then you hand it to Cursor, Claude Code, or any AI tool — and it builds **within the guardrails**, not from scratch on a blank canvas.

## Why This Matters

| Without alty                        | With alty                                             |
| ----------------------------------- | ----------------------------------------------------- |
| AI guesses at structure             | Structure is planned from your actual business        |
| Change one thing, break five others | Changes stay contained in their area                  |
| No tests until something breaks     | Tests are defined before code is written              |
| Rewrite every few months            | Built to last from day one                            |
| Finish one task, others go stale    | Completing work auto-flags what needs review          |
| Works with one AI tool              | Works with Cursor, Claude Code, Roo Code, OpenCode |

## Three Commands. That's It.

```bash
# Starting a new project
alty init

# Already have a project that's gotten messy? Apply structure to it
alty init --existing

# Check if your documentation is still accurate
alty doc-health
```

**`alty init` guides you through everything.** It shows you what it will do, asks you to confirm, and never touches files without your permission.

## Installation

### From Source (requires Go 1.25+)

```bash
git clone <your-repo-url>
cd alty-cli
make release
./bin/alty version
```

This produces two binaries in `bin/`:
- `alty` — CLI tool
- `alty-mcp` — MCP server for AI tool integration

### Cross-Platform Binaries

```bash
make release-all
```

Builds for 5 platforms: Linux (amd64/arm64), macOS (amd64/arm64), Windows (amd64).

## How It Works (The Simple Version)

```
Your idea (a few sentences)
     |
     v
alty asks questions about your business
     |
     v
Creates a blueprint: what belongs together, what stays separate
     |
     v
Generates a build plan: tasks in the right order, with tests
     |
     v
Configures your AI tool with your specific rules and language
     |
     v
Your AI tool builds it — correctly, within guardrails
     |
     v
Task completed? alty flags affected tasks for review
     |
     v
Your plan stays fresh — no stale context, no outdated assumptions
```

## Six Things That Make alty Different

### 1. It Asks Before It Builds

Every other tool starts writing code immediately. alty starts by understanding your business. The 20 minutes of questions saves you 20 hours of rewrites.

### 2. The Guardrails Are Automatic

alty doesn't just write rules in a document — it creates **automated tests** that catch mistakes. If anyone (human or AI) writes code that crosses a boundary, the test fails. Your architecture enforces itself.

### 3. It Tells You What to Build Next

After planning, you get an ordered list of tasks. Each task tells you:

- What to build
- What test to write first
- What "done" looks like
- What must be built before it

No guessing. No jumping ahead. Just follow the list.

### 4. It Works With Your AI Tool, Not Instead of It

alty is not another AI coding tool. It's the **prep work** for the tool you already use. It generates configuration files in your tool's native format — so Claude Code, Cursor, or any other tool understands your project's rules from the start.

### 5. Your Tasks Never Go Stale

This is the one nobody else does. **When you complete a task, alty automatically flags every related task that might be affected.**

Here's the problem: you finish a research spike and discover the architecture needs to change. But five other tasks were written assuming the old architecture. Every project management tool (Jira, Linear, GitHub) only detects staleness by _time_ — "this ticket hasn't been touched in 30 days." None of them detect staleness by _event_ — "the thing this ticket depends on just changed."

alty does. When a task closes, it:

- Traverses the dependency graph to find affected open tasks
- Records _what changed_ (the context diff) so reviewers know what's different
- Flags those tasks as needing review
- Shows you exactly what might need updating and lets you decide

No more starting work based on outdated assumptions. No more discovering mid-sprint that the plan changed three tickets ago.

### 6. It Can Fix Messy Projects Too

Already have a codebase that's become hard to change? `alty init --existing` analyzes what you have, identifies the problems, and creates a step-by-step migration plan — all on a separate branch. **Your existing code is never touched until you approve every change.**

## Safety Promises

- **You see everything before it happens** — Nothing runs without your OK
- **Your files are never overwritten** — Conflicts get renamed, never replaced
- **Existing projects stay safe** — Always works on a new branch, never on your main code
- **Your tests must still pass** — If anything breaks, everything rolls back automatically

## Development

### Prerequisites

- Go 1.25+
- golangci-lint (for linting)
- gofumpt (for formatting)

### Common Commands

```bash
make build      # Quick build (no optimization)
make test       # Run tests with race detector
make lint       # Run golangci-lint
make check      # All quality gates: build → vet → test → lint → deadcode
make ci         # Alias for check (CI-friendly)
make fmt        # Format code with gofumpt
make clean      # Remove build artifacts
```

## Status

Go implementation complete. Core CLI commands (`init`, `doc-health`, `detect`) are functional.

## License

Apache-2.0
