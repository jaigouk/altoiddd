---
last_verified: "2026-03-04"
confidence: high
next_review_date: "2026-06-04"
---

# Quality Gates

Quick reference for AI agents running quality checks before completing tasks.

## Gate Order

Run in this order. Fix from top to bottom -- lint errors often cause type errors.

```
1. Lint    (ruff)   -- syntax and style
2. Types   (mypy)   -- type correctness
3. Tests   (pytest) -- behavior correctness
```

## Gate 1: Lint

```bash
uv run ruff check .              # Check for violations
uv run ruff check --fix .        # Auto-fix safe violations
uv run ruff format --check .     # Check formatting
```

**Requirement:** Zero errors, zero warnings.

| Issue | Fix |
|-------|-----|
| `F401` Unused import | Remove the import |
| `F841` Unused variable | Prefix with `_` or remove |
| `E501` Line too long | Break line at 100 chars |
| `I001` Import order | `uv run ruff check --fix --select I .` |

## Gate 2: Types

```bash
uv run mypy .
```

**Requirement:** Zero errors in strict mode. Every function needs parameter AND return annotations.

Use `str | None` not `Optional[str]`. Use `list[str]` not `List[str]`.

## Gate 3: Tests

```bash
uv run pytest -v --cov=src --cov-report=term-missing
```

**Requirements:** All tests pass. No skipped tests. Coverage >= 80% on changed code.

## Pre-Completion Checklist

```bash
uv run ruff check .
uv run ruff format --check .
uv run mypy .
uv run pytest -v --cov=src --cov-report=term-missing
```

**If ANY gate fails, you are NOT done.** Fix issues yourself before requesting review.
