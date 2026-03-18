---
date: 2026-03-03
author: researcher
status: complete
topic: Multi-Language/Multi-Stack Tooling for alto
---

# Research: Multi-Language/Multi-Stack Tooling for alto

**Date:** 2026-03-03
**Status:** Final

## Summary

This spike evaluates OSS tools that could help alto handle multi-language/multi-stack
project scaffolding across five problem areas: stack detection, project layout conventions,
quality gate knowledge, architecture fitness testing, and template-based scaffolding. The
research finds **strong options for scaffolding (Copier) and architecture testing per-language**,
but **no single tool solves stack/framework detection or cross-language layout conventions**.
Those two problems will require alto to build its own knowledge base.

## Research Questions

1. What tools detect the language/framework stack of a project directory?
2. What tools encode idiomatic project layout conventions per stack?
3. What tools map stacks to appropriate linters, type checkers, and test runners?
4. What tools generate architecture boundary tests across languages?
5. What tools generate project boilerplate from parameterized templates?

---

## Problem 1: Stack/Language Detection

### Tools Investigated

#### go-enry (Go library with Python bindings)

| Property | Value |
|---|---|
| GitHub | [go-enry/go-enry](https://github.com/go-enry/go-enry) |
| Stars | 594 |
| License | Apache-2.0 |
| Python bindings | `enry` on PyPI -- v0.1.1, last updated **August 2020** (STALE) |
| Go library | Active -- synced with github/linguist v9.4.0 data |
| What it detects | **Programming languages** (by file extension, shebang, heuristics, Bayesian classifier) |
| What it does NOT detect | **Frameworks** (cannot distinguish Next.js from plain React, FastAPI from Flask) |

#### GitHub Linguist (Ruby)

| Property | Value |
|---|---|
| GitHub | [github-linguist/linguist](https://github.com/github-linguist/linguist) |
| Stars | 12,400+ |
| License | MIT |
| Limitation | Ruby-only, requires a git repo, detects **languages not frameworks** |

### Verdict: No Good Solution Exists

**Build it ourselves.**

No OSS tool detects *frameworks* (Next.js, FastAPI, Django, Chi, Axum) -- they all stop
at *languages* (Python, Go, Rust, TypeScript). Framework detection requires inspecting:

- **Package manifests**: `pyproject.toml` dependencies, `package.json` dependencies,
  `go.mod` requires, `Cargo.toml` dependencies
- **Config files**: `next.config.js`, `vite.config.ts`, `tsconfig.json` paths
- **Directory conventions**: `pages/` or `app/` (Next.js), `src/routes/` (SvelteKit)

**Minimal approach for alto:**

A `StackDetector` domain service that reads package manifests and config files,
returning a `DetectedStack` value object (language, framework, build tool, test runner).
This is a table-driven approach -- a TOML/YAML registry mapping
`{indicator_file: framework}`. Example:

```yaml
indicators:
  - file: "pyproject.toml"
    contains: "fastapi"
    stack: {language: python, framework: fastapi}
  - file: "next.config.js"
    stack: {language: typescript, framework: nextjs}
  - file: "go.mod"
    contains: "github.com/go-chi/chi"
    stack: {language: go, framework: chi}
```

Estimated effort: 1 ticket (small). The indicator registry lives in
`.alto/knowledge/stacks/` and is extensible by users.

---

## Problem 2: Project Layout Conventions

### Tools Investigated

No OSS tool encodes "the idiomatic project structure for Python+FastAPI is X, for
Go+Chi is Y." This is **domain knowledge**, not a tool problem.

### Verdict: Build it ourselves (knowledge base)

**Minimal approach for alto:**

A `LayoutConvention` registry in `.alto/knowledge/stacks/` with per-stack TOML files:

```toml
# .alto/knowledge/stacks/python_fastapi.toml
[layout]
domain = "src/domain/"
application = "src/application/"
infrastructure = "src/infrastructure/"
tests = "tests/"

[layout.ddd_mapping]
entities = "src/domain/models/"
services = "src/domain/services/"
ports = "src/application/ports/"
adapters = "src/infrastructure/"

# .alto/knowledge/stacks/go_chi.toml
[layout]
domain = "internal/domain/"
application = "internal/app/"
infrastructure = "internal/infra/"
tests = "internal/domain/ (co-located _test.go files)"

[layout.ddd_mapping]
entities = "internal/domain/"
services = "internal/domain/"
ports = "internal/app/ports/"
adapters = "internal/infra/"
```

This is the same living knowledge base pattern alto already uses for tool conventions.
Estimated effort: 1-2 tickets per supported stack.

---

## Problem 3: Quality Gate / Linter Knowledge

### Tools Investigated

No OSS tool maps "given stack X, use linters Y." This is also domain knowledge.

### Verdict: Build it ourselves (knowledge base)

**Minimal approach:** Extend the stack convention files:

```toml
# .alto/knowledge/stacks/python_fastapi.toml
[quality_gates]
lint = "uv run ruff check ."
lint_fix = "uv run ruff check --fix ."
typecheck = "uv run mypy ."
test = "uv run pytest"
format = "uv run ruff format ."

[quality_gates.packages]
lint = "ruff"
typecheck = "mypy"
test = "pytest"

# .alto/knowledge/stacks/go_chi.toml
[quality_gates]
lint = "golangci-lint run"
test = "go test ./..."
format = "gofmt -w ."

# .alto/knowledge/stacks/typescript_nextjs.toml
[quality_gates]
lint = "npx eslint ."
typecheck = "npx tsc --noEmit"
test = "npx vitest"
format = "npx prettier --write ."
```

Estimated effort: folded into the layout convention tickets above.

---

## Problem 4: Architecture Fitness Testing Across Languages

This is the most complex problem. Each language has its own ecosystem of tools.

### Python (already researched -- see 20260222_python_architecture_testing_fitness_functions.md)

| Tool | Stars | Version | Last Release | License |
|---|---|---|---|---|
| [import-linter](https://github.com/seddonym/import-linter) | ~800 | 2.10 | 2026-02-06 | BSD-2-Clause |
| [pytestarch](https://github.com/zyskarch/pytestarch) | ~200 | 2.0+ | Active 2025-2026 | Apache-2.0 |

**Status:** Already designed in `20260223_fitness_function_design.md`. alto generates
import-linter TOML contracts + pytestarch test files from the bounded context map.

### TypeScript / JavaScript

#### ArchUnitTS

| Property | Value |
|---|---|
| GitHub | [LukasNiessen/ArchUnitTS](https://github.com/LukasNiessen/ArchUnitTS) |
| Stars | 334 |
| Version | 2.1.63 (npm: `archunit`) |
| Last Release | ~December 2025 (published 3 months ago per npm) |
| License | MIT |
| Test Frameworks | Jest, Vitest, Jasmine |
| Features | Layer validation, circular dependency detection, naming conventions, code metrics, HTML reports |

#### ts-arch

| Property | Value |
|---|---|
| GitHub | [ts-arch/ts-arch](https://github.com/ts-arch/ts-arch) |
| Stars | 613 |
| Version | 5.4.1 |
| Last Release | December 23, 2024 |
| License | MIT |
| Features | File-based and slice-based architecture tests, Nx monorepo support |
| Concern | Last release is 14+ months old. Still usable but borderline on the "active" criterion. |

#### eslint-plugin-boundaries

| Property | Value |
|---|---|
| GitHub | [javierbrea/eslint-plugin-boundaries](https://github.com/javierbrea/eslint-plugin-boundaries) |
| Stars | 795 |
| Version | 5.4.0 |
| Last Release | **February 2, 2026** |
| License | MIT |
| Features | ESLint-based boundary enforcement, element type definitions, dependency rules, real-time IDE feedback |
| Integration | Runs as part of existing ESLint pipeline -- zero extra tooling |

**TypeScript recommendation:** Use **eslint-plugin-boundaries** as primary (most stars, most
recently active, integrates into existing ESLint). ArchUnitTS as secondary for richer
programmatic rules (layer tests, metrics). Both are MIT licensed.

### Go

#### arch-go

| Property | Value |
|---|---|
| GitHub | [arch-go/arch-go](https://github.com/arch-go/arch-go) |
| Stars | 250 |
| Version | 2.1.2 |
| Last Release | **February 3, 2026** |
| License | MIT |
| Features | Dependency checks, package content validation, function property rules, naming conventions, HTML/JSON reports |
| Config | YAML-based (`arch-go.yml`) |
| Usage | CLI tool + programmatic Go test integration |

#### go-arch-lint

| Property | Value |
|---|---|
| GitHub | [fe3dback/go-arch-lint](https://github.com/fe3dback/go-arch-lint) |
| Stars | 451 |
| Version | 1.14.0 |
| Last Release | November 13, 2025 |
| License | MIT |
| Features | Import path analysis, YAML config, CI/CD integration, dependency graphs, Docker support |
| Focus | Hexagonal/onion/DDD/MVC boundary enforcement via import rules |

#### kcmvp/archunit (Go)

| Property | Value |
|---|---|
| GitHub | [kcmvp/archunit](https://github.com/kcmvp/archunit) |
| Stars | 25 |
| Version | 0.2.0-alpha1 |
| Last Release | September 23, 2025 |
| License | Apache-2.0 |
| Note | Too early/small for production use. Interesting "Code as Promotion" AI concept. |

**Go recommendation:** Use **arch-go** as primary (most recently released, richest feature
set, YAML config similar to import-linter's approach). **go-arch-lint** as alternative
(more stars, DDD-focused import rules). Both are MIT licensed.

### Rust

#### cargo-modules

| Property | Value |
|---|---|
| GitHub | [regexident/cargo-modules](https://github.com/regexident/cargo-modules) |
| Stars | 1,200 |
| Version | 0.25.0 |
| Last Release | **October 16, 2025** |
| License | MPL-2.0 |
| Features | Module tree visualization, dependency graph analysis, orphan detection, **`--acyclic` flag for cycle detection** |
| Limitation | Visualization + analysis tool, not a test/assertion framework |

**Rust situation:** There is **no ArchUnit equivalent for Rust**. The Rust module system
and Cargo workspace structure provide built-in boundary enforcement through visibility
(`pub`, `pub(crate)`, `pub(super)`) that other languages lack. Rust's `mod` system means
you cannot accidentally import from a private module.

**Rust recommendation:** For Rust, alto should generate:
1. **Cargo workspace structure** with crate-per-bounded-context (built-in boundary enforcement)
2. **cargo-modules `--acyclic` checks** in CI to detect circular dependencies
3. **`#[cfg(test)]` integration tests** that assert module structure if needed

This is a "build it ourselves with light tooling" approach. Rust needs less tooling because
its module system provides architectural enforcement at the compiler level.

### Cross-Language Summary

| Language | Primary Tool | Config Format | Stars | License | Last Release |
|---|---|---|---|---|---|
| Python | import-linter + pytestarch | TOML + Python | 800+200 | BSD/Apache | Feb 2026 |
| TypeScript | eslint-plugin-boundaries | ESLint config (JS/JSON) | 795 | MIT | Feb 2026 |
| Go | arch-go | YAML | 250 | MIT | Feb 2026 |
| Rust | cargo-modules + workspace layout | Cargo.toml | 1,200 | MPL-2.0 | Oct 2025 |

**Key finding:** Each language has its own config format. alto's fitness function generator
must produce **language-specific outputs** from the same bounded context map input. The
YAML schema designed in `20260223_fitness_function_design.md` is extensible to this -- we
add a `target_language` field and dispatch to language-specific renderers.

---

## Problem 5: Project Scaffolding / Template Engines

### Tools Investigated

#### Copier (RECOMMENDED)

| Property | Value |
|---|---|
| GitHub | [copier-org/copier](https://github.com/copier-org/copier) |
| Stars | 3,200 |
| Version | 9.12.0 |
| Last Release | **February 21, 2026** |
| License | MIT |
| Language | Python (installable via `pip`/`uv`) |
| Template Engine | Jinja2 |
| Python API | `from copier import run_copy` -- full programmatic control |
| Key Feature | **Template updates** -- when template evolves, Copier can update existing projects |

**Why Copier is the best fit for alto:**

1. **Python native** -- alto is Python, Copier is Python. Direct dependency, no subprocess.
2. **Programmatic API** -- `run_copy(template_path, dest, data={...})` does exactly what
   alto needs: generate from a template with answers.
3. **Language agnostic** -- templates are just directories of files with Jinja2 placeholders.
   Can scaffold Python, Go, Rust, TypeScript -- anything.
4. **Template update lifecycle** -- `run_update()` can re-apply template changes to
   existing projects. This directly supports `alto init --existing` rescue mode.
5. **Active maintenance** -- 4 releases in Jan-Feb 2026 alone.
6. **Questionnaire support** -- `copier.yml` defines questions with types, defaults,
   validators. alto can either use this or pass answers directly via `data=`.

```python
from copier import run_copy

# alto generates the answers from guided DDD discovery,
# then passes them to Copier for file generation
worker = run_copy(
    "path/to/alto/templates/python_fastapi",
    "./user-project",
    data={
        "project_name": "my_service",
        "root_package": "my_service",
        "bounded_contexts": ["orders", "payments"],
        "use_docker": True,
    },
    defaults=True,
    overwrite=False,
)
```

#### Cookiecutter (LEGACY)

| Property | Value |
|---|---|
| GitHub | [cookiecutter/cookiecutter](https://github.com/cookiecutter/cookiecutter) |
| Stars | 23,600 |
| Version | 2.6.0 |
| Last Release | **12+ months ago** (no new releases in 2025-2026) |
| License | BSD-3-Clause |
| Status | Sustainable but **effectively in maintenance mode** |

**Why not Cookiecutter:** No new releases in 12+ months. No template update/migration
support. Copier is its modern successor with a superset of features.

#### Yeoman (Node.js)

| Property | Value |
|---|---|
| GitHub | [yeoman/generator](https://github.com/yeoman/generator) |
| Stars | ~3,800 (yo CLI) |
| Version | 7.5.1 |
| Last Release | ~9 months ago |
| License | BSD-2-Clause |
| Concern | Node.js ecosystem. Adding Node.js as a runtime dependency for a Python CLI is undesirable. |

**Why not Yeoman:** Wrong ecosystem. alto is Python; adding a Node.js dependency is
architecturally unsound.

#### Hygen (Node.js)

| Property | Value |
|---|---|
| GitHub | [jondot/hygen](https://github.com/jondot/hygen) |
| Stars | 6,000 |
| Last Commit | **September 2022** (3.5 years ago -- DEAD) |
| License | MIT |

**Why not Hygen:** Unmaintained since 2022. Node.js dependency. Dead project.

#### Scaffold (Go)

| Property | Value |
|---|---|
| GitHub | [hay-kot/scaffold](https://github.com/hay-kot/scaffold) |
| Stars | 126 |
| License | MIT |
| Concern | Too small. Not enough adoption for production dependency. |

**Why not Scaffold:** Under the 1000-star threshold. Go binary -- can't be used as
Python library.

#### cargo-generate (Rust-specific)

| Property | Value |
|---|---|
| GitHub | [cargo-generate/cargo-generate](https://github.com/cargo-generate/cargo-generate) |
| Stars | 2,400 |
| Version | 0.23.7 |
| Last Release | November 20, 2025 |
| License | Apache-2.0 / MIT dual |
| Limitation | **Rust-only** -- generates Cargo projects from templates |

**Why not cargo-generate:** Rust-specific. Not usable for Python/Go/TypeScript projects.

### Scaffolding Recommendation

**Use Copier.** It is the clear winner:
- Python native with programmatic API
- Language-agnostic templates
- Template update lifecycle (critical for rescue mode)
- Most actively maintained of all options
- MIT license

---

## Nx as a Meta-Tool (Considered and Rejected)

| Property | Value |
|---|---|
| GitHub | [nrwl/nx](https://github.com/nrwl/nx) |
| Stars | 28,200 |
| Version | 22.5.0 |
| Last Release | February 9, 2026 |
| License | MIT |
| Language Plugins | Go (nx-go), Python (nx-python), Rust (nx-cargo) |

Nx was considered because it handles multi-language monorepos with task orchestration,
boundary enforcement (`@nx/enforce-module-boundaries`), and project generation.

**Why not Nx for alto:**

1. **Node.js runtime dependency** -- alto is Python. Adding Node.js + npm is a heavy
   dependency for users who are scaffolding Go or Rust projects.
2. **Monorepo-first** -- Nx assumes a monorepo structure. alto scaffolds individual
   projects, not monorepos.
3. **Opinionated project structure** -- Nx imposes its own `apps/` + `libs/` layout that
   conflicts with DDD-native structures.
4. **Overkill** -- alto needs template rendering and config generation, not a build system.

However, alto should be *aware* of Nx -- if a user's project is an Nx monorepo, alto
should respect its structure during `alto init --existing`.

---

## Consolidated Recommendations

### What to adopt as dependencies

| Tool | Problem | How alto uses it | Integration |
|---|---|---|---|
| **Copier** (v9.12.0) | P5: Scaffolding | Template rendering engine for `alto init` | Python dependency (`from copier import run_copy`) |

### What to generate as project output (not alto dependencies)

| Tool | Problem | What alto generates | Target Language |
|---|---|---|---|
| import-linter | P4: Fitness tests | TOML contract configs | Python |
| pytestarch | P4: Fitness tests | Python test files | Python |
| eslint-plugin-boundaries | P4: Fitness tests | ESLint config rules | TypeScript/JS |
| arch-go | P4: Fitness tests | YAML config files | Go |
| cargo-modules (--acyclic) | P4: Fitness tests | CI pipeline commands | Rust |

### What to build ourselves

| Problem | Approach | Estimated Effort |
|---|---|---|
| P1: Stack detection | `StackDetector` service + indicator registry (TOML) | 1 ticket (small) |
| P2: Layout conventions | Per-stack convention files in `.alto/knowledge/stacks/` | 1-2 tickets per stack |
| P3: Quality gate knowledge | Folded into layout convention files | Same as above |
| P4: Cross-language fitness function renderer | Extend existing fitness function generator with language dispatch | 1 ticket (medium) per language |

---

## Risk Assessment

1. **Copier version coupling** -- Copier's API is still evolving (v9.x). Pin to a specific
   minor version and test on upgrade. The `run_copy()` API has been stable since v7.

2. **Per-language fitness tools are shallow** -- arch-go (250 stars), ArchUnitTS (334 stars)
   are small projects. They could become unmaintained. Mitigation: alto generates config
   files, not code that imports these libraries. If a tool dies, users can replace it
   without changing alto.

3. **Stack detection heuristics will have false positives** -- A project with both
   `next.config.js` and `package.json` containing Express could be misdetected. Mitigation:
   always ask the user to confirm detected stack. The heuristic is a suggestion, not a
   decision.

4. **Layout conventions are opinionated** -- "Where does domain code go in Go?" has
   multiple valid answers (`internal/domain/`, `pkg/domain/`, `domain/`). Mitigation:
   conventions are defaults, not enforced. Users can override via `.alto/config.toml`.

5. **Rust needs least tooling, Go needs most** -- Rust's module system provides built-in
   enforcement. Go has no module visibility beyond package-level, so it needs the most
   external tooling. TypeScript and Python fall in between.

---

## References

- [Copier GitHub](https://github.com/copier-org/copier) -- 3,200 stars, MIT, v9.12.0 (Feb 2026)
- [Copier docs](https://copier.readthedocs.io/) -- Python API reference
- [go-enry](https://github.com/go-enry/go-enry) -- 594 stars, Apache-2.0 (Python bindings stale)
- [import-linter](https://github.com/seddonym/import-linter) -- BSD-2-Clause, v2.10 (Feb 2026)
- [eslint-plugin-boundaries](https://github.com/javierbrea/eslint-plugin-boundaries) -- 795 stars, MIT, v5.4.0 (Feb 2026)
- [arch-go](https://github.com/arch-go/arch-go) -- 250 stars, MIT, v2.1.2 (Feb 2026)
- [go-arch-lint](https://github.com/fe3dback/go-arch-lint) -- 451 stars, MIT, v1.14.0 (Nov 2025)
- [cargo-modules](https://github.com/regexident/cargo-modules) -- 1,200 stars, MPL-2.0, v0.25.0 (Oct 2025)
- [ArchUnitTS](https://github.com/LukasNiessen/ArchUnitTS) -- 334 stars, MIT, v2.1.63
- [ts-arch](https://github.com/ts-arch/ts-arch) -- 613 stars, MIT, v5.4.1 (Dec 2024)
- [Cookiecutter](https://github.com/cookiecutter/cookiecutter) -- 23,600 stars, BSD-3, v2.6.0 (stale)
- [Hygen](https://github.com/jondot/hygen) -- 6,000 stars, MIT (dead since Sep 2022)
- [Nx](https://github.com/nrwl/nx) -- 28,200 stars, MIT, v22.5.0 (Feb 2026)
- [cargo-generate](https://github.com/cargo-generate/cargo-generate) -- 2,400 stars, Apache/MIT, v0.23.7 (Nov 2025)
- [Existing fitness function research](docs/research/20260222_python_architecture_testing_fitness_functions.md)
- [Existing fitness function design](docs/research/20260223_fitness_function_design.md)

## Follow-up Tasks

- [ ] Evaluate adding `copier` as a dependency to alto's `pyproject.toml`
- [ ] Design `StackDetector` domain service + indicator registry schema
- [ ] Create `.alto/knowledge/stacks/` convention files for Python, Go, TypeScript, Rust
- [ ] Extend fitness function generator (`20260223_fitness_function_design.md`) with per-language renderers
- [ ] Design the bounded context map schema extension for `target_language` field
