---
date: 2026-02-22
author: researcher
status: complete
topic: Python Architecture Testing and Fitness Function Tools
---

# Spike: Python Architecture Testing and Fitness Function Tools (2026)

## 1. Research Questions

1. Are there Python ports of ArchUnit (Java)? How mature are they?
2. What tools exist for enforcing architectural boundaries in Python via fitness functions?
3. Are there tools that auto-generate architecture tests from domain models or bounded context maps?
4. What is the state of test generation from architectural specifications?
5. If alty wanted to auto-generate architecture fitness function tests from a bounded context map, what libraries would it build on?

## 2. Context

alty generates DDD project scaffolding including bounded context maps, layer structure
(`domain/`, `application/`, `infrastructure/`), and architectural docs. A capability under
consideration is auto-generating architecture fitness function tests from those artifacts so
generated projects arrive with enforced boundaries from day one.

Constraints (from `docs/PRD.md`):
- Python 3.12+, `uv` package manager
- No cloud dependencies — everything runs locally
- No paid APIs required
- Permissive licenses required (Apache 2.0, MIT, BSD)

---

## 3. Tool Landscape

### 3.1 import-linter

| Property | Value |
|---|---|
| Current version | 2.10 (released 2026-02-06) |
| License | BSD 2-Clause |
| Python support | 3.10, 3.11, 3.12, 3.13, 3.14 |
| Underlying engine | grimp (dependency graph library, same author) |
| Maintenance | Actively maintained; frequent 2025–2026 releases |

**How it works:** Reads a `.ini` or `pyproject.toml` config file defining "contracts". Runs as
a CLI tool (`lint-imports`) or can be invoked from pytest using `lint_imports()`. The underlying
`grimp` library (v3.14, released December 2025) builds a queryable directed import graph.

**Contract types:**
- `layers` — enforces that upper layers cannot import from lower layers
- `forbidden` — blocks specific modules from importing specific other modules
- `independence` — ensures a set of modules cannot depend on each other
- `acyclic_siblings` — forbids cyclic dependencies between sibling modules
- `protected` — prevents direct import except from an allow-list
- Custom contract types — extend with Python code

**DDD application example** (`pyproject.toml`):
```toml
[tool.importlinter]
root_package = "myapp"

[[tool.importlinter.contracts]]
name = "DDD layer contract"
type = "layers"
layers = [
  "myapp.infrastructure",
  "myapp.application",
  "myapp.domain",
]
```
This enforces: infrastructure can import application and domain; application can import domain;
domain cannot import anything above it.

**Programmatic API:** Limited. The Python API exposes only `read_configuration()` for
reading contracts from a config file. There is no public API for creating or running
contracts in pure Python code without a config file. To use it programmatically, you would
generate a config file and call the CLI or use pytest integration.

**Source:** https://import-linter.readthedocs.io/en/stable/
**PyPI:** https://pypi.org/project/import-linter/

---

### 3.2 pytestarch

| Property | Value |
|---|---|
| Current version | 4.0.1 (released 2025-08-08) |
| License | Apache-2.0 |
| Python support | 3.9–3.13 |
| GitHub stars | ~145 |
| Contributors | 8 |
| Total releases | 27 |
| Maintenance | Actively maintained |

**How it works:** Scans Python files under a source directory, builds an internal dependency
representation, then allows defining rules via a fluent Python API. Tests integrate directly
with pytest without any config file.

**Fluent API example:**
```python
from pytestarch import get_evaluable_architecture, Rule

evaluable = get_evaluable_architecture(".", ".")

rule = (
    Rule()
    .modules_that()
    .are_named("myapp.infrastructure")
    .should_not()
    .be_imported_by_modules_that()
    .are_named("myapp.domain")
)
rule.assert_applies(evaluable)
```

**DDD layer fixture pattern:**
```python
# conftest.py
import pytest
from pytestarch import get_evaluable_architecture, LayeredArchitecture

@pytest.fixture(scope="session")
def evaluable():
    return get_evaluable_architecture(".", "src")

@pytest.fixture(scope="session")
def layered_architecture():
    return (
        LayeredArchitecture()
        .layer("infrastructure").containing_modules(["src.infrastructure"])
        .layer("application").containing_modules(["src.application"])
        .layer("domain").containing_modules(["src.domain"])
    )
```

**Key differentiator from import-linter:** Rules are Python code, not config files. This
makes programmatic test generation feasible — you can write a function that takes a bounded
context map data structure and emits pytest test functions.

**Visualization:** Optional `pytestarch[visualization]` installs matplotlib and allows
generating dependency graphs.

**Source:** https://zyskarch.github.io/pytestarch/latest/
**PyPI:** https://pypi.org/project/PyTestArch/

---

### 3.3 pytest-archon

| Property | Value |
|---|---|
| Current version | 0.0.7 (released 2025-09-19) |
| License | Apache-2.0 |
| Python support | 3.8–3.11 |
| GitHub stars | ~78 |
| Maintenance | Last commit January 2025 |

**How it works:** Pytest plugin. Rules are written as Python test code using `archrule()`
and chained predicates. Supports fnmatch patterns and regular expressions.

```python
from pytest_archon import archrule

def test_domain_has_no_infrastructure_deps():
    (
        archrule("domain isolation")
        .match("myapp.domain*")
        .should_not()
        .import_modules("myapp.infrastructure*")
        .check("myapp")
    )
```

**Notable features:**
- `only_top_level_imports=True` — checks only top-level (not nested) imports
- `exclude_type_checking=True` — ignores `TYPE_CHECKING` blocks
- Predicate logic for custom constraints

**Concern:** Python 3.8–3.11 listed as supported; no evidence of 3.12/3.13 support.
Project activity appears to have slowed since January 2025.

**Source:** https://github.com/jwbargsten/pytest-archon
**PyPI:** https://pypi.org/project/pytest-archon/

---

### 3.4 pytest-arch

| Property | Value |
|---|---|
| GitHub stars | 14 |
| Releases | None published |
| License | Apache-2.0 |
| Maintenance | Minimal / experimental |

A very small project explicitly described as "a pythonic derivative of ArchUnit, in the form
of a pytest plugin." No releases published. Not recommended for production use.

**Source:** https://github.com/nwilbert/pytest-arch

---

### 3.5 deply

| Property | Value |
|---|---|
| Current version | 0.8.0 |
| License | BSD-3-Clause |
| Python support | 3.8–3.12 |
| GitHub stars | ~169 |
| Config format | YAML |
| Maintenance | Active (120 commits) |

**How it works:** YAML-driven static analysis tool (not pytest-based). Defines layers via
collectors (file patterns, class inheritance, regex), then ruleset defines which layers
can/cannot depend on others.

**Differentiator:** Goes beyond imports — rules can enforce naming conventions, decorator
usage, and inheritance patterns. Generates Mermaid diagrams of the architecture.

**Example config (`deply.yaml`):**
```yaml
paths:
  - src

layers:
  - name: domain
    collectors:
      - type: file_regex
        regex: ".*/domain/.*\\.py"
  - name: infrastructure
    collectors:
      - type: file_regex
        regex: ".*/infrastructure/.*\\.py"

ruleset:
  - name: domain_cannot_use_infra
    layers:
      - domain
    disallow:
      - infrastructure
```

**Concern:** Not a pytest plugin — runs as a standalone CLI tool. Less composable with
pytest-based test suites.

**Source:** https://github.com/vashkatsi/deply

---

### 3.6 grimp (underlying engine)

| Property | Value |
|---|---|
| Current version | 3.14 (released 2025-12-10) |
| License | BSD |
| Author | Same as import-linter (David Seddon) |

grimp is the import graph engine underlying import-linter. It provides a programmatic
`ImportGraph` API with methods like `find_illegal_dependencies_for_layers()`,
`find_upstream_modules()`, and `find_descendants()`. If you need to write custom architecture
analysis, grimp is the lowest-level building block.

**Source:** https://grimp.readthedocs.io/en/stable/

---

## 4. Java Ecosystem Reference: ArchUnit + jMolecules + Context Mapper

For comparison, the Java ecosystem in early 2026 has a much more mature stack:

| Tool | Role |
|---|---|
| **ArchUnit** | Architecture rule testing in JUnit tests, fluent API |
| **jMolecules 2.0** (Nov 2025) | DDD type annotations (`@Entity`, `@BoundedContext`, `@ValueObject`) baked into code |
| **Context Mapper** | DSL for bounded context maps (`.cml` files) + ArchUnit extension for validating code against the model |
| **Spring Modulith** | Detects jMolecules DDD building blocks for module documentation |

The key insight from this stack: **jMolecules makes the DDD model machine-readable in code via
annotations**, and ArchUnit+Context Mapper can then validate the implementation matches the model.
There is no Python equivalent of this full stack.

**Sources:**
- https://contextmapper.org/docs/architecture-validation-with-archunit/
- https://odrotbohm.de/2025/11/jmolecules-2.0-stereotypical/

---

## 5. Auto-Generation of Fitness Functions from Specifications

### 5.1 Current State of Art

There is **no existing Python tool** that auto-generates architecture fitness function tests
from a bounded context map or architectural specification. This is a gap in the ecosystem.

The approaches that exist are:

**1. Manual config files (import-linter):** You write TOML contracts by hand based on your
architecture. No generation from a model — the config IS the model.

**2. Manual test code (pytestarch, pytest-archon):** You write Python test functions by hand
using the fluent API. Programmable but not auto-generated.

**3. Python AST module + grimp (DIY):** The Python stdlib `ast` module and `grimp` together
provide the building blocks to write a custom generator. One approach described in the wild:

```python
import ast
def analyze_imports_with_ast(path: str, forbidden_layer: str):
    with open(path, "r") as f:
        source = ast.parse(f.read())
        # traverse imports, check against forbidden_layer
```

**4. Hypothesis (property-based testing):** The `hypothesis` library generates random inputs
satisfying a specification. In 2025, an AI agent using Hypothesis autonomously found bugs in
NumPy. Hypothesis is not architecture-specific, but it demonstrates auto-generation of tests
from specifications is technically feasible for behavioral testing.

### 5.2 The Generation Gap: What Would Need to Be Built

To auto-generate architecture fitness function tests from a bounded context map (e.g., a
YAML/JSON representation of bounded contexts, their layers, and their allowed dependencies),
you would need to:

1. Parse the bounded context map into a data structure
2. For each bounded context, emit `import-linter` TOML contracts or `pytestarch` test functions
3. For each cross-context boundary, emit `forbidden` or `independence` contracts

**Which library to build on:**

| Need | Best choice | Rationale |
|---|---|---|
| Runtime test generation (pytest tests as code) | `pytestarch` | Rules are pure Python; can emit test functions programmatically |
| Config-file generation (simpler CI integration) | `import-linter` | Generate TOML from bounded context map YAML; run `lint-imports` |
| Low-level custom analysis | `grimp` | Direct API access to the import graph for custom rules |
| Beyond imports (decorators, naming, inheritance) | `deply` | YAML config, generatable from a template |

**Recommended approach for alty:**

Generate `pyproject.toml` import-linter contracts from the bounded context map at project
init time. This is the simplest and most composable option:
1. alty reads bounded context map (YAML)
2. Generates `[[tool.importlinter.contracts]]` sections for each bounded context
3. Developer runs `lint-imports` or `uv run lint-imports` in CI

For richer rule-as-code (fitness function style), generate pytestarch test files:
1. alty generates `tests/architecture/test_layers.py`
2. File contains pytestarch `Rule()` assertions derived from the context map
3. Runs with `uv run pytest tests/architecture/`

---

## 6. Adoption and Maturity Summary

| Tool | Stars | Latest Release | Python 3.12+ | License | Approach | Recommended |
|---|---|---|---|---|---|---|
| import-linter | N/A (high) | 2.10 (2026-02) | Yes | BSD-2 | Config file | Yes (CI/CD focus) |
| pytestarch | ~145 | 4.0.1 (2025-08) | Yes | Apache-2.0 | pytest code | Yes (programmatic) |
| deply | ~169 | 0.8.0 | No (3.8-3.12) | BSD-3 | YAML CLI | Possibly (richer rules) |
| pytest-archon | ~78 | 0.0.7 (2025-09) | Unclear | Apache-2.0 | pytest code | Caution (activity slow) |
| pytest-arch | 14 | None | Unknown | Apache-2.0 | pytest code | No (experimental) |
| grimp | N/A | 3.14 (2025-12) | Yes | BSD | Python API | Yes (low-level DIY) |

**Note:** import-linter's star count is not visible on PyPI; it is the most-referenced tool
in blog posts and the Clean Architecture with Python book (O'Reilly, 2025), indicating high
adoption relative to alternatives.

---

## 7. Fitness Function Definition

"Building Evolutionary Architectures" (Ford, Parsons, Kua, O'Brien) defines:
> An architectural fitness function provides an objective integrity assessment of some
> architectural characteristic(s).

In Python, import boundary enforcement tests qualify as fitness functions when run in CI.
The `handsonarchitects.com` 2026 article explicitly frames pytestarch rules as fitness
functions, citing Mark Richards.

The key principle: **fitness functions must run automatically on every build**, making them
part of the CI pipeline rather than one-off checks.

---

## 8. Recommendations for alty

### Primary Recommendation

Use **import-linter** for generated architecture fitness functions in alty scaffolding.

**Rationale:**
- Actively maintained (released Feb 6, 2026 — days before this research)
- Permissive license (BSD-2)
- Config-driven: alty can generate TOML contracts from a bounded context map without
  any runtime Python code generation
- Runs standalone CLI (`lint-imports`) and integrates with `uv run`
- Supports exactly the DDD layer pattern alty uses (domain/application/infrastructure)
- Highest real-world adoption among Python architecture testing tools

### Secondary Recommendation

Use **pytestarch** for more expressive, code-based fitness functions where rules need to be
context-specific or conditional.

**Rationale:**
- Rules are Python code → composable with alty's Python code generation
- Apache-2.0 license
- Actively maintained (27 releases, 8 contributors)
- Supports visualization of dependency graphs (useful for generated projects)

### Not Recommended

- `pytest-arch` (14 stars, no releases)
- `pytest-archon` (unclear Python 3.12+ support, slowing activity)

### What Does Not Exist Yet

No tool auto-generates fitness function tests from a bounded context map. alty would be
building novel capability by generating `import-linter` TOML blocks or `pytestarch` test
files from its bounded context representation. The building blocks exist; the connector is
the gap.

---

## 9. Sources

- import-linter PyPI: https://pypi.org/project/import-linter/
- import-linter docs: https://import-linter.readthedocs.io/en/stable/
- pytestarch PyPI: https://pypi.org/project/PyTestArch/
- pytestarch docs: https://zyskarch.github.io/pytestarch/latest/
- pytestarch GitHub: https://github.com/zyskarch/pytestarch
- pytest-archon PyPI: https://pypi.org/project/pytest-archon/
- pytest-archon GitHub: https://github.com/jwbargsten/pytest-archon
- pytest-arch GitHub: https://github.com/nwilbert/pytest-arch
- deply GitHub: https://github.com/vashkatsi/deply
- grimp docs: https://grimp.readthedocs.io/en/stable/
- Context Mapper + ArchUnit: https://contextmapper.org/docs/architecture-validation-with-archunit/
- jMolecules 2.0: https://odrotbohm.de/2025/11/jmolecules-2.0-stereotypical/
- Protecting Architecture with Automated Tests in Python (2026): https://handsonarchitects.com/blog/2026/protecting-architecture-with-automated-tests-in-python/
- Fitness Functions in Python — Makimo: https://makimo.com/blog/govern-software-architecture-with-fitness-functions-in-python/
- Building Evolutionary Architectures (O'Reilly): https://www.oreilly.com/library/view/building-evolutionary-architectures/9781492097532/ch04.html
- Hypothesis (property-based testing): https://hypothesis.works/
- Clean Architecture with Python book (O'Reilly, 2025): https://www.oreilly.com/library/view/clean-architecture-with/9781836642893/
