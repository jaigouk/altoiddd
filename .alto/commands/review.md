---
name: review
description: Structured code review based on Hartwork methodology — bugs, clarity, DDD/SOLID, tests
allowed-tools: Read, Grep, Glob, Bash
---

# /review <target>

Structured code review following Sebastian Pipping's methodology, adapted for DDD + TDD projects.

## Usage

```
/review                          # Review all uncommitted changes (staged + unstaged)
/review --staged                 # Review only staged changes
/review <commit-sha>             # Review a specific commit
/review <branch>                 # Review branch diff against main
/review <file-path>              # Review a specific file
/review <ticket-id>              # Review all changes related to a ticket
```

## Process

### Step 1 — Gather the Diff

Based on `<target>`:
- No target or `--staged`: `git diff --cached` (staged) or `git diff` (all uncommitted)
- Commit SHA: `git show <sha>`
- Branch: `git diff main...<branch>`
- File path: `git diff HEAD -- <path>` + read the full file for context
- Ticket ID: `git log --all --grep="<ticket-id>" --oneline` → review those commits

List all changed files with `git diff --stat` for the relevant range.

### Step 2 — Read Changed Files

For EACH changed file, read:
1. The full diff (what changed)
2. The full file (surrounding context)
3. Related test files (same package, `_test.go` suffix)

### Step 3 — Review Checklist

Evaluate every change against these categories. Report findings per file.

#### A. Bugs & Correctness

- [ ] Off-by-one errors, boundary conditions
- [ ] Null/None handling — can any value be unexpectedly None?
- [ ] Exception handling — are the right exceptions caught? Are any swallowed silently?
- [ ] Race conditions or state mutation issues
- [ ] Enum exhaustiveness — do match/if-else chains handle all variants?
- [ ] Return value correctness — is the right thing returned in all branches?

#### B. Missing Elements

- [ ] Missing test coverage for new/changed code
- [ ] Missing edge case handling (empty input, very large input, special characters)
- [ ] Missing error messages or unclear error messages
- [ ] Missing type annotations
- [ ] Missing validation at system boundaries (user input, file I/O, external data)

#### C. Clarity & Cognitive Complexity

- [ ] Misleading variable/function names
- [ ] Functions doing too many things (> 1 responsibility)
- [ ] Deeply nested logic (> 3 levels) — can it be flattened?
- [ ] Magic numbers or strings — should be constants or enums
- [ ] Dead code, commented-out code, unused imports

#### D. DDD Compliance

- [ ] Domain layer has ZERO external dependencies
- [ ] Business logic lives in domain objects, not services or handlers
- [ ] Value Objects are immutable (unexported fields, constructor validation)
- [ ] Ubiquitous language — do names match `docs/DDD.md` glossary?
- [ ] Aggregate boundaries — one aggregate per transaction
- [ ] Dependencies flow inward: infrastructure → application → domain
- [ ] Ports are interfaces in `application/ports.go`, adapters in `infrastructure/`
- [ ] Compile-time interface checks: `var _ Port = (*Adapter)(nil)`

#### E. SOLID Compliance

- [ ] **SRP**: Does each class/function have one reason to change?
- [ ] **OCP**: Can behavior be extended without modifying existing code?
- [ ] **LSP**: Do subtypes honor the contract of their base type?
- [ ] **ISP**: Are interfaces focused (no "god protocols")?
- [ ] **DIP**: Do modules depend on abstractions, not concretions?

#### F. Test Quality

- [ ] Tests follow RED/GREEN/REFACTOR — test was written first?
- [ ] Tests are independent (no shared mutable state between tests)
- [ ] Tests cover happy path AND edge cases from ticket
- [ ] Test names describe the scenario: `Test<Type>_<Method>_<Condition>`
- [ ] No logic in tests (no if/else, minimal setup)
- [ ] Mocks are minimal — only mock infrastructure, never domain

#### G. Commit Hygiene

- [ ] Each commit does ONE thing (single responsibility)
- [ ] Refactoring is separate from feature/bugfix commits
- [ ] Commit messages explain WHY, not just WHAT
- [ ] No formatting-only changes mixed with logic changes
- [ ] No dead code left behind — Git history preserves deleted code

#### H. Security (lightweight)

- [ ] No secrets, credentials, or PII in code or comments
- [ ] No `exec.Command` with unsanitized input
- [ ] File paths are validated/sanitized before use
- [ ] Serialization/deserialization handles malformed input gracefully
- [ ] Error messages don't leak internal state or stack traces to users

### Step 4 — Produce Report

Output a structured report:

```
## Code Review: <target>

### Summary
- Files changed: N
- Verdict: [APPROVE | REQUEST CHANGES | DISCUSS]
- Risk level: [LOW | MEDIUM | HIGH]

### Findings

#### <file-path>

| # | Severity | Category | Line | Finding |
|---|----------|----------|------|---------|
| 1 | CRITICAL | Bugs | 42 | Description of issue |
| 2 | MAJOR | DDD | 15 | Description of issue |
| 3 | MINOR | Clarity | 88 | Description of issue |
| 4 | NIT | Style | 12 | Description of issue |

### Quality Gate Status

| Gate | Result |
|------|--------|
| `go build ./...` | PASS/FAIL |
| `go vet ./...` | PASS/FAIL |
| `golangci-lint run` | PASS/FAIL |
| `go test -race ./...` | PASS/FAIL |

### Recommendations
1. [Actionable recommendation]
2. [Actionable recommendation]
```

### Severity Definitions

| Severity | Meaning | Action |
|----------|---------|--------|
| **CRITICAL** | Bug, data loss, security hole, broken functionality | Must fix before merge |
| **MAJOR** | DDD/SOLID violation, missing tests, unclear logic | Should fix before merge |
| **MINOR** | Naming, style, minor clarity improvement | Fix if easy, otherwise note |
| **NIT** | Preference, optional improvement | Author's discretion |

### Verdict Rules

- **APPROVE**: Zero CRITICAL, zero MAJOR. Only MINOR/NIT findings.
- **REQUEST CHANGES**: Any CRITICAL or MAJOR finding.
- **DISCUSS**: Architectural question that needs team/user input before deciding.
