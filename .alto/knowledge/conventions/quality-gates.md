# Quality Gates

Quick reference for AI agents running quality checks before completing tasks.

---

## Gate Order

Run in this order. Fix issues from top to bottom -- lint errors often cause type errors, and type errors can cause test failures.

```
1. Lint    (ruff)   -- syntax and style
2. Types   (mypy)   -- type correctness
3. Tests   (pytest) -- behavior correctness
```

---

## Gate 1: Lint

```bash
uv run ruff check .              # Check for violations
uv run ruff check --fix .        # Auto-fix safe violations
uv run ruff format --check .     # Check formatting
uv run ruff format .             # Auto-format
```

**Requirement:** Zero errors, zero warnings.

**Common issues:**

| Issue | Fix |
|-------|-----|
| `F401` Unused import | Remove the import |
| `F841` Unused variable | Prefix with `_` or remove |
| `E501` Line too long | Break line at 100 chars |
| `I001` Import order | Use ruff's isort: `uv run ruff check --fix --select I .` |
| `UP` Modern Python | Use `list[str]` not `List[str]`, `str \| None` not `Optional[str]` |

**Notes:**
- Run `--fix` freely for safe auto-fixes (imports, formatting)
- Review unsafe fixes (`--unsafe-fixes`) before applying
- Check `pyproject.toml` for project-specific ruff configuration

---

## Gate 2: Types

```bash
uv run mypy .                    # Full type check
uv run mypy src/domain/          # Check specific directory
```

**Requirement:** Zero errors in strict mode.

**Common issues:**

| Issue | Fix |
|-------|-----|
| `Missing return type` | Add `-> ReturnType` to every function |
| `Missing type annotation` | Annotate all function parameters |
| `Incompatible types` | Fix the type or add proper conversion |
| `has no attribute` | Check the type is correct, or narrow with `isinstance` |
| `Module has no attribute` | Check imports, may need `py.typed` marker or stub |
| `Argument of type X not assignable to Y` | Use Protocol for duck typing, or fix the type |

**Type annotation rules:**
- Every function: parameters AND return type annotated
- Use `str \| None` not `Optional[str]`
- Use `list[str]` not `List[str]` (Python 3.12+)
- Use `ClassVar[...]` for mutable class-level attributes
- Use `Protocol` for structural typing (duck typing with type safety)

**Example fixes:**

```python
# Bad: mypy cannot infer
def process(data):
    return data.get("key")

# Good: fully annotated
def process(data: dict[str, Any]) -> str | None:
    return data.get("key")
```

---

## Gate 3: Tests

```bash
uv run pytest                                # Run all tests
uv run pytest tests/domain/                  # Run domain tests only
uv run pytest -v                             # Verbose output
uv run pytest -v --cov=src --cov-report=term-missing  # With coverage
uv run pytest -x                             # Stop on first failure
uv run pytest -k "test_order"                # Run matching tests
```

**Requirements:**
- All tests pass
- No skipped tests (remove `@pytest.mark.skip` before completing)
- Coverage >= 80% on changed code

**Common issues:**

| Issue | Fix |
|-------|-----|
| `ImportError` | Check module path, ensure `__init__.py` exists |
| `ModuleNotFoundError` | Run `uv sync` to install dependencies |
| `AssertionError` | Fix the implementation to match expected behavior |
| `fixture not found` | Check fixture name, scope, and conftest.py location |
| Flaky test (passes sometimes) | Remove external dependencies, fix shared state |
| Test passes but should not | Verify the assertion is actually checking something |

**Test requirements:**
- Tests mirror `src/` structure under `tests/`
- Domain tests: no mocks, no I/O, no external dependencies
- Application tests: mock ports (repositories, external services)
- Infrastructure tests: integration tests with real adapters

---

## Pre-Completion Checklist

Before claiming a task is complete, run all three gates in sequence:

```bash
uv run ruff check .
uv run ruff format --check .
uv run mypy .
uv run pytest -v --cov=src --cov-report=term-missing
```

**If ANY gate fails, you are NOT done.**

Do not ask the user to review until all gates pass. Fix issues yourself.

---

## Troubleshooting

| Problem | Fix |
|---------|-----|
| Lint fix breaks types | Use `TYPE_CHECKING` block for type-only imports |
| Assertion looks wrong but passes | Check for missing `assert` keyword, wrong variable, vacuous pass |
| mypy "no attribute" but code works | Runtime type differs from declared type -- fix annotation or `isinstance` narrow |
| ruff and mypy disagree on imports | Use `if TYPE_CHECKING:` block to satisfy both |
