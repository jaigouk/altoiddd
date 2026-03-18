# Research: TechStack Strategy Design

**Date:** 2026-03-03
**Spike Ticket:** alto-5li.1
**Status:** Final

## Summary

alto's pipeline is hardcoded to Python+uv in ~15 locations across 8 files. The fix: introduce a `StackProfile` protocol so stack-specific knowledge is pluggable, implement `PythonUvProfile` first, and gracefully skip stack-specific stages for unknown stacks.

## Research Question

1. Where does alto assume Python? (answered — see Audit below)
2. Which pipeline stages are stack-neutral vs stack-specific?
3. What's the simplest design to make this extensible?

## Audit: Python Assumptions

### Quality Gates (4 locations)

| File | Line(s) | Hardcoded Value |
|------|---------|-----------------|
| `subprocess_gate_runner.py` | 29-33 | `uv run ruff/mypy/pytest` commands |
| `tool_adapter.py` | 79-88 | `_build_quality_gates()` markdown block |
| `ticket_detail_renderer.py` | 83-86, 211-218 | Quality gates in ticket descriptions |
| `quality_gate.py` | — | `QualityGate` enum names (LINT/TYPES/TESTS/FITNESS) are generic enough |

### Fitness Functions (3 locations)

| File | Line(s) | Hardcoded Value |
|------|---------|-----------------|
| `fitness_test_suite.py` | entire | `import-linter` TOML + `pytestarch` Python test files |
| `fitness_values.py` | entire | `ContractType` = import-linter types, `ArchRule` = pytestarch |
| `fitness_generation_handler.py` | 92-98 | Writes `importlinter.toml` + `tests/architecture/test_fitness.py` |

### Project Structure (3 locations)

| File | Line(s) | Hardcoded Value |
|------|---------|-----------------|
| `rescue_handler.py` | 38-42 | `_REQUIRED_STRUCTURE = ("src/domain/", "src/application/", "src/infrastructure/")` |
| `project_scanner.py` | 29, 33-37 | `pyproject.toml`, same directory structure |
| `generate.py` + `main.py` | 138, 205 | `root_package = Path.cwd().name.replace("-", "_")` |

### Tool Configs (2 locations)

| File | Line(s) | Hardcoded Value |
|------|---------|-----------------|
| `tool_adapter.py` | 142 | `globs: **/*.py` in Cursor rules |
| `tool_adapter.py` | all adapters | `_build_quality_gates()` called in every adapter |

### Discovery Flow

| File | Finding |
|------|---------|
| `question.py` | **No question about tech stack.** Q5 is about workflows, not tech. |
| `domain_model.py` | **No `tech_stack` field.** DomainModel is purely DDD. |

## Pipeline Classification

| Stage | Stack-Neutral? | Needs StackProfile? |
|-------|:-:|:-:|
| Discovery questions (Q1-Q10) | YES | NO |
| DomainModel building | YES | NO |
| PRD generation | YES | NO |
| DDD.md generation | YES | NO |
| ARCHITECTURE.md generation | MOSTLY | Optional (could include stack sections) |
| **Fitness test generation** | **NO** | YES — gate behind `fitness_available()` |
| **Quality gate execution** | **NO** | YES — commands from profile |
| **Quality gate display in configs** | **NO** | YES — markdown from profile |
| **Quality gate display in tickets** | **NO** | YES — markdown from profile |
| **Tool config globs** | **NO** | YES — file glob from profile |
| **Rescue scan (structure)** | **NO** | YES — layout from profile |
| **Rescue scan (configs)** | **NO** | YES — manifest from profile |
| **Root package derivation** | **NO** | YES — naming convention from profile |

**Result:** 6 stages are stack-neutral (no changes). 7 stages are stack-specific (need profile).

## Design

### Option A: StackProfile Protocol (Recommended)

A `StackProfile` protocol in `src/domain/models/` provides all stack-specific knowledge. Each stage reads from the profile instead of hardcoding. For unknown stacks, a `GenericProfile` skips fitness and provides no quality gate commands.

```
StackProfile (Protocol)
├── PythonUvProfile     → full pipeline (ruff, mypy, pytest, import-linter, pytestarch)
└── GenericProfile      → DDD artifacts only, skips fitness/quality gates
```

**Where the profile lives:** On the `DiscoverySession` (set during discovery or pre-flight check). Flows through to handlers via the existing event/handler chain.

**Key protocol surface:**

```python
@runtime_checkable
class StackProfile(Protocol):
    stack_id: str                                    # "python-uv"
    file_glob: str                                   # "**/*.py"
    project_manifest: str                            # "pyproject.toml"
    source_layout: tuple[str, ...]                   # ("src/domain/", ...)
    quality_gate_commands: dict[str, list[str]]       # {"lint": ["uv", "run", "ruff", ...]}
    quality_gate_display: str                        # Markdown block for configs/tickets
    root_package_convention: Callable[[str], str]     # project-name → project_name
    fitness_available: bool                           # True for Python, False for others
```

### Option B: Simple Boolean

Add `is_python: bool` to DomainModel. If True, full pipeline. If False, skip fitness/quality gates.

**Pros:** Minimal change, fast to implement.
**Cons:** Dead end. Adding TypeScript later means a second boolean, then a third, etc.

### Option C: TOML Knowledge Base

Store stack profiles as `.alto/knowledge/stacks/python-uv.toml`. The pipeline reads TOML at runtime.

**Pros:** User-editable, no code changes for new stacks.
**Cons:** Overengineered for now. Same protocol still needed internally.

## Recommendation

**Option A: StackProfile Protocol.**

- Clean DDD: stack knowledge is a domain concept, modeled as a protocol
- Open/Closed: new stacks = new profile class, no existing code changes
- Minimal scope: implement `PythonUvProfile` + `GenericProfile` only
- Defers complexity: Option C (TOML) can layer on top later as a TOML-backed profile

### User-Facing Flow

1. During `alto guide` or `alto init`, ask: "Are you using Python with uv and pyproject.toml? (y/n)"
2. If yes → attach `PythonUvProfile` to session → full pipeline
3. If no → attach `GenericProfile` to session → PRD + DDD.md + ARCHITECTURE.md + tickets (without quality gate commands)

### Migration Path

Each hardcoded location gets refactored to read from the profile:

| Current Hardcode | Reads From |
|-----------------|------------|
| `SubprocessGateRunner._GATE_COMMANDS` | `profile.quality_gate_commands` |
| `_build_quality_gates()` | `profile.quality_gate_display` |
| `_render_quality_gates_section()` | `profile.quality_gate_display` |
| `FitnessTestSuite` | Gated by `profile.fitness_available` |
| `CursorAdapter` globs | `profile.file_glob` |
| `_REQUIRED_STRUCTURE` | `profile.source_layout` |
| `_CONFIG_TARGETS` | `profile.project_manifest` |
| `root_package` derivation | `profile.root_package_convention` |

## Follow-up Tasks

1. **Add `StackProfile` protocol + `PythonUvProfile` + `GenericProfile`** to `src/domain/models/`
2. **Add `TechStack` value object** to `src/domain/models/` (language, package_manager, is_backend_only)
3. **Add tech stack question** to discovery flow (or pre-flight in `alto init`)
4. **Thread profile through session → event → handlers** so each handler can access it
5. **Refactor quality gate commands** — SubprocessGateRunner reads from profile
6. **Refactor quality gate display** — tool_adapter + ticket_detail_renderer read from profile
7. **Refactor fitness generation** — gate behind `profile.fitness_available`
8. **Refactor rescue/scanner** — structure + config targets from profile
9. **Refactor root_package derivation** — use `profile.root_package_convention`
10. **Update generated CLAUDE.md template** — quality gates from profile, not hardcoded

## References

- Related research: `docs/research/20260303_multi_language_multi_stack_tooling.md` — OSS tools survey
- Per-language fitness: import-linter (Python), eslint-plugin-boundaries (TS), arch-go (Go), cargo-modules (Rust)
- Template engine candidate: Copier (v9.12.0, MIT, 3.2k stars) — for future multi-stack scaffolding
