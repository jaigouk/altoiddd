---
last_reviewed: 2026-02-23
owner: researcher
status: complete
type: spike
ticket: alty-k7m.8
---

# Gap Analysis Design for Existing Projects

> **Spike:** k7m.8 -- Design how `alty init --existing` analyzes an existing project and produces a gap report
> **Timebox:** 3 hours
> **Date:** 2026-02-23
> **Status:** Final

## Summary

This document defines the gap analysis system for `alty init --existing`. It covers
what a fully-seeded project looks like (the golden state), what to scan, how to
categorize gaps, how to present results to different personas, and how to safely
apply changes with test verification and rollback. P0 (basic scaffold overlay) and
P1 (structural migration) are clearly separated.

## Research Questions Answered

1. What does a "fully-seeded" project look like? (Section 1)
2. What should the scanner look at? (Section 2)
3. How should gaps be categorized? (Section 3)
4. How should the gap report be formatted for different personas? (Section 4)
5. How does the branch workflow and rollback work? (Section 5)
6. How should test verification work? (Section 6)
7. How can we detect existing domain language? (Section 7)
8. What is P0 vs P1? (Section 8)
9. How do migration tickets integrate with two-tier generation? (Section 9)

---

## 1. Golden State: What a Fully-Seeded Project Looks Like

A fully-seeded project has four layers of completeness. The gap analysis compares
an existing project against this reference state.

### 1.1 File Manifest (the canonical checklist)

Every item below is checked during gap analysis. Items are grouped by scan category.

**Layer 1: Project Infrastructure (always required)**

| Path | Purpose | Check Type |
|------|---------|------------|
| `.alty/config.toml` | alty project config | File exists |
| `.alty/knowledge/` | RLM-addressable knowledge base | Directory exists |
| `.alty/maintenance/doc-registry.toml` | Doc review tracking | File exists |
| `.gitignore` | Contains alty + Python entries | Content check |
| `.python-version` | Python 3.12+ declared | File exists + version check |
| `pyproject.toml` | Project metadata + tool config | File exists + content check |

**Layer 2: Documentation (always required)**

| Path | Purpose | Check Type |
|------|---------|------------|
| `README.md` | Project description (4-5 sentences minimum) | File exists + content check (word count) |
| `docs/PRD.md` | Product requirements | File exists + frontmatter check |
| `docs/DDD.md` | DDD artifacts | File exists + frontmatter check |
| `docs/ARCHITECTURE.md` | Architecture decisions | File exists + frontmatter check |
| `docs/templates/PRD_TEMPLATE.md` | PRD template | File exists |
| `docs/templates/DDD_STORY_TEMPLATE.md` | DDD story template | File exists |
| `docs/templates/ARCHITECTURE_TEMPLATE.md` | Architecture template | File exists |
| `docs/beads_templates/beads-epic-template.md` | Epic template | File exists |
| `docs/beads_templates/beads-spike-template.md` | Spike template | File exists |
| `docs/beads_templates/beads-ticket-template.md` | Ticket template | File exists |
| `docs/spikes/ddd_reference.md` | DDD reference | File exists |
| `AGENTS.md` | Agent instructions | File exists |

**Layer 3: Source Structure (DDD layers)**

| Path | Purpose | Check Type |
|------|---------|------------|
| `src/domain/` | Domain layer | Directory exists |
| `src/domain/models/` | Entities, VOs, Aggregates | Directory exists |
| `src/domain/services/` | Domain services | Directory exists |
| `src/domain/events/` | Domain events | Directory exists |
| `src/application/` | Application layer | Directory exists |
| `src/application/commands/` | Write operations | Directory exists |
| `src/application/queries/` | Read operations | Directory exists |
| `src/application/ports/` | Protocols for infrastructure | Directory exists |
| `src/infrastructure/` | Infrastructure layer | Directory exists |
| `src/infrastructure/persistence/` | File/DB storage | Directory exists |
| `src/infrastructure/messaging/` | Events, MCP | Directory exists |
| `src/infrastructure/external/` | Third-party integrations | Directory exists |
| `tests/domain/` | Domain unit tests | Directory exists |
| `tests/application/` | Application unit tests | Directory exists |
| `tests/infrastructure/` | Integration tests | Directory exists |
| `tests/integration/` | End-to-end tests | Directory exists |
| `*/__init__.py` | Package markers in all dirs | File exists per Python dir |

**Layer 4: Tool Configs (per detected tool)**

| Path | Purpose | Check Type |
|------|---------|------------|
| `.claude/CLAUDE.md` | Claude Code config | File exists (if Claude Code detected) |
| `.claude/agents/*.md` | Agent personas | Files exist (6 agents) |
| `.claude/commands/*.md` | Slash commands | Files exist |
| `.cursor/mcp.json` | Cursor MCP config | File exists (if Cursor detected) |
| `.cursor/rules/` | Cursor rules dir | Directory exists |
| `.antigravity/rules.md` | Antigravity rules | File exists (if Antigravity detected) |
| `.beads/` | Beads initialized | Directory exists |

**Layer 5: Quality Gates (content checks in pyproject.toml)**

| Check | What to verify |
|-------|----------------|
| ruff configured | `[tool.ruff]` section in pyproject.toml |
| mypy configured | `[tool.mypy]` section in pyproject.toml |
| pytest configured | `[tool.pytest]` section or pytest dependency |
| import-linter configured | `[tool.importlinter]` section (post-fitness-function generation) |

### 1.2 Non-File Checks

| Check | How |
|-------|-----|
| Git repo exists | `.git/` directory |
| Clean working tree | `git status --porcelain` is empty |
| Beads installed | `command -v bd` |
| uv installed | `command -v uv` |
| Python 3.12+ | `python --version` or `.python-version` content |

---

## 2. Scan Targets

The scanner runs in a fixed order, from cheap/fast to expensive/slow.

### 2.1 Scan Pipeline

```
Phase 1: Environment Check (~0.1s)
  - Git repo? Clean tree? Branch available?
  - uv installed? Python version?
  - AI tools detected? (claude, cursor, antigravity)
  - Beads installed?

Phase 2: File Structure Scan (~0.2s)
  - Walk golden state manifest (Section 1.1)
  - For each item: exists? matches expected content?
  - Detect existing source layout (flat vs DDD vs other)
  - Detect existing test layout

Phase 3: Config Content Scan (~0.3s)
  - pyproject.toml: check tool sections (ruff, mypy, pytest)
  - .gitignore: check for required entries
  - Global settings: check for conflicts (same as alty init)

Phase 4: Documentation Scan (~0.5s)
  - Frontmatter in docs/*.md: last_reviewed, owner, status
  - README.md: word count (4-5 sentences ~ 40+ words)
  - PRD.md, DDD.md, ARCHITECTURE.md: exist? have content?

Phase 5: Domain Language Detection (~1s) [P1 only]
  - Package names -> potential bounded contexts
  - Class names -> potential entities/VOs
  - Docstrings -> potential ubiquitous language terms
  - Module-level comments -> potential domain stories
```

### 2.2 What the Scanner Does NOT Do (Scope Limits)

The scanner is a **static file check**, not a code analysis tool. It does NOT:

- Parse Python ASTs (that is P1 structural migration)
- Analyze import graphs (that is fitness function generation)
- Evaluate code quality (that is ruff/mypy)
- Run existing tests (that is the separate test verification step)
- Modify any files (the scanner is read-only)

The scanner produces a `GapAnalysis` value object; the scaffolder acts on it.

---

## 3. Gap Categories

Every item in the golden state manifest gets exactly one classification.

### 3.1 Category Definitions

| Category | Code | Description | Action |
|----------|------|-------------|--------|
| **Compliant** | `OK` | Item exists and matches golden state | No action |
| **Missing** | `MISSING` | Item does not exist at all | Create during scaffolding |
| **Partial** | `PARTIAL` | Item exists but incomplete (e.g., pyproject.toml without ruff config) | Suggest additions, do not overwrite |
| **Conflicting** | `CONFLICT` | Item exists but differs from golden state (e.g., different CLAUDE.md) | Create `_alty` suffixed copy |
| **Incompatible** | `INCOMPAT` | Item exists and is structurally incompatible (e.g., flat `src/` layout vs DDD layers) | Report only; suggest migration (P1) |

### 3.2 Category Resolution Rules

```python
def classify_gap(item: GoldenStateItem, project: ProjectScan) -> GapCategory:
    if not item.exists_in(project):
        return GapCategory.MISSING

    if item.check_type == "file_exists":
        return GapCategory.COMPLIANT

    if item.check_type == "content_check":
        match = item.content_matches(project)
        if match == ContentMatch.FULL:
            return GapCategory.COMPLIANT
        elif match == ContentMatch.PARTIAL:
            return GapCategory.PARTIAL
        elif match == ContentMatch.DIFFERENT:
            return GapCategory.CONFLICT
        else:  # INCOMPATIBLE
            return GapCategory.INCOMPATIBLE

    if item.check_type == "directory_exists":
        return GapCategory.COMPLIANT
```

### 3.3 Priority Within Categories

Not all MISSING items are equal. Priority is assigned based on the layer:

| Layer | Priority | Rationale |
|-------|----------|-----------|
| Infrastructure (.alty/, .gitignore) | P0 | Needed for alty to function |
| Documentation (PRD, DDD, ARCHITECTURE) | P0 | Foundation for everything else |
| Source structure (DDD layers) | P1 | Existing code may use different layout |
| Tool configs (.claude/, .cursor/) | P0 | Low risk, high value |
| Quality gates (ruff, mypy config) | P0 | Non-breaking additions |
| Beads setup | P0 | Parallel to existing workflow |

---

## 4. Gap Report Format

### 4.1 Terminal Output (Primary)

The gap report is shown before any changes are made. It follows the same
preview-then-confirm pattern as `alty init`.

```
alty init --existing: project-name
========================================

Environment:
  OK       Git repo (clean tree)
  OK       Python 3.12.7 (via .python-version)
  OK       uv 0.6.2
  OK       Claude Code detected (~/.claude/)
  MISSING  Beads (will install)

Documentation:                                       [2/5 present]
  OK       README.md (87 words)
  MISSING  docs/PRD.md
  MISSING  docs/DDD.md
  MISSING  docs/ARCHITECTURE.md
  PARTIAL  pyproject.toml (missing [tool.ruff], [tool.mypy] sections)

Source Structure:                                    [0/4 DDD layers]
  INCOMPAT src/ has flat layout (models.py, views.py, urls.py)
           Detected: Django-style flat structure
           Recommendation: Keep existing layout; DDD dirs added alongside
  MISSING  src/domain/, src/application/, src/infrastructure/
  MISSING  tests/domain/, tests/application/, tests/infrastructure/

Tool Configs:                                        [0/3 configured]
  MISSING  .claude/CLAUDE.md
  MISSING  .claude/agents/ (6 agent personas)
  MISSING  .claude/commands/ (slash commands)
  CONFLICT .cursor/mcp.json (exists, no context7 entry)

alty Config:                                    [0/3 present]
  MISSING  .alty/config.toml
  MISSING  .alty/knowledge/
  MISSING  .alty/maintenance/doc-registry.toml

Quality Gates:
  PARTIAL  pyproject.toml: pytest configured, ruff and mypy missing
  OK       Existing test suite detected (pytest)

Summary:
  Compliant:    4
  Missing:      14
  Partial:      2
  Conflicting:  1
  Incompatible: 1

Proposed changes: 14 files created, 1 merged, 2 sections added
```

### 4.2 Persona-Adapted Output

The report adapts to detected persona (or defaults to developer if no persona
detection has occurred -- `--existing` typically runs before guided discovery).

**Developer (default):**
- Full file paths with action codes (CREATE, SKIP, RENAME, etc.)
- Technical details: config sections, version numbers, directory listings
- Shows every item in the golden state checklist

**Product Owner (`alty init --existing --persona po`):**
- High-level health score: "Your project is 22% ready for structured development"
- Category counts only, no file paths
- Business-language summary: "Missing: product requirements document, domain model, architecture documentation"
- No DDD jargon: "bounded context" becomes "separate business area"

**Team Lead (`alty init --existing --persona lead`):**
- Category counts with file paths for MISSING and CONFLICT only
- Convention compliance summary: "Quality gates: 1/3 configured"
- Emphasis on team impact: "6 agent personas will be added for consistent AI tool behavior"

### 4.3 Persona Output Examples

**Product Owner view:**

```
Project Health: project-name
========================================

Readiness Score: 22% (4 of 18 items present)

What's working:
  - Project has a README with clear description
  - Python environment is set up correctly
  - Tests exist and pass

What's missing:
  - Product Requirements Document (PRD) -- defines what to build and why
  - Domain Model -- captures your business processes and terminology
  - Architecture Document -- technical blueprint
  - AI tool configurations -- helps AI assistants follow your conventions

What needs attention:
  - Code quality tools partially configured (linting works, type checking missing)
  - Source code uses a different organization than recommended
    (this is OK -- we will add new folders alongside existing ones)

Next steps after applying alty:
  1. Work with a developer to create a PRD from your product vision
  2. Answer domain discovery questions to capture your business knowledge
  3. Review and approve generated documentation
```

### 4.4 Markdown Report (Optional)

`alty init --existing --report` writes the gap report to
`docs/gap-analysis-report.md` on the `alty/init` branch. This report
persists as a record of the project's initial state and is useful for:

- Team review before merging the branch
- Tracking progress over time (re-run `alty init --existing --report` to compare)
- P1 migration planning (identifies structural issues)

The Markdown report includes everything from the developer terminal view plus
a machine-readable YAML frontmatter block:

```yaml
---
generated_by: alty
version: 0.1.0
date: 2026-02-23
project: project-name
readiness_score: 0.22
gaps:
  compliant: 4
  missing: 14
  partial: 2
  conflicting: 1
  incompatible: 1
---
```

---

## 5. Branch Workflow Design

### 5.1 Full Lifecycle

```
User runs: alty init --existing
  |
  v
[Pre-flight checks]
  - Is this a git repo? (abort if not)
  - Is working tree clean? (abort if dirty)
  - Does alty/init branch exist? (abort if yes, suggest --force-branch)
  |
  v
[Environment scan] -- read-only, no git operations
  - Detect tools, global settings, existing structure
  |
  v
[Gap analysis] -- read-only, no git operations
  - Compare against golden state
  - Classify every item
  |
  v
[Show gap report + proposed changes] -- preview
  - Terminal output (persona-adapted)
  - "Create branch 'alty/init' and apply these changes? [y/N]"
  |
  v
[User confirms]
  |
  v
[Create branch]
  - git checkout -b alty/init
  - Record original branch name for rollback
  |
  v
[Apply scaffolding] -- write operations
  - Create MISSING files
  - Create _alty copies for CONFLICT files
  - Add config sections for PARTIAL items (append, never overwrite)
  - Skip COMPLIANT and INCOMPATIBLE items
  |
  v
[Post-scaffold setup]
  - Setup .gitignore entries
  - Setup IDE configs (Claude Code MCP, Cursor MCP, Antigravity)
  - Initialize Beads (if not present)
  - uv sync (if pyproject.toml created/modified)
  |
  v
[Test verification] -- hard gate (Section 6)
  - Run existing test suite
  - Compare against baseline
  - On failure: rollback (Section 5.2)
  |
  v
[Commit scaffolding to branch]
  - git add (specific files, not -A)
  - git commit -m "chore: apply alty scaffolding"
  |
  v
[Show summary]
  - Files created/modified count
  - How to review: git diff original...alty/init
  - How to merge: git checkout original && git merge alty/init
  - How to discard: git checkout original && git branch -D alty/init
```

### 5.2 Rollback Mechanism

Rollback happens when:
1. Test verification fails (hard gate)
2. User interrupts (Ctrl-C)
3. Any critical error during scaffolding

```bash
rollback() {
    local original_branch="$1"
    local init_branch="alty/init"

    # Discard all changes on the branch
    git checkout -- .
    git clean -fd

    # Switch back to original branch
    git checkout "$original_branch"

    # Delete the init branch
    git branch -D "$init_branch"

    echo "Rolled back. Branch '$init_branch' deleted."
}
```

Key safety properties:
- Original branch is NEVER modified
- All changes are on `alty/init` only
- Rollback deletes the entire branch (clean slate)
- No partial state: either all scaffolding applies or none does

### 5.3 Force-Branch Option

If `alty/init` already exists (from a previous aborted attempt):

```
alty init --existing --force-branch
```

This deletes the existing `alty/init` branch and creates a fresh one.
Without `--force-branch`, the command aborts with a message explaining the
situation.

---

## 6. Test Verification Design

### 6.1 Test Runner Detection

The scanner detects the existing test runner by checking multiple signals in
priority order:

```python
def detect_test_runner(project_dir: Path) -> TestRunnerConfig | None:
    """Detect existing test runner and build the command to run it.

    Returns None if no test runner detected.
    """
    pyproject = project_dir / "pyproject.toml"
    setup_cfg = project_dir / "setup.cfg"
    tox_ini = project_dir / "tox.ini"
    makefile = project_dir / "Makefile"

    # Priority 1: pyproject.toml [tool.pytest] or pytest dependency
    if pyproject.exists():
        content = pyproject.read_text()
        if "[tool.pytest" in content or "pytest" in content:
            # Check if uv is the package manager
            if (project_dir / "uv.lock").exists():
                return TestRunnerConfig(
                    runner="pytest",
                    command=["uv", "run", "pytest"],
                    source="pyproject.toml + uv.lock",
                )
            return TestRunnerConfig(
                runner="pytest",
                command=["python", "-m", "pytest"],
                source="pyproject.toml",
            )

    # Priority 2: pytest.ini
    if (project_dir / "pytest.ini").exists():
        return TestRunnerConfig(
            runner="pytest",
            command=["python", "-m", "pytest"],
            source="pytest.ini",
        )

    # Priority 3: setup.cfg with [tool:pytest]
    if setup_cfg.exists():
        content = setup_cfg.read_text()
        if "[tool:pytest]" in content:
            return TestRunnerConfig(
                runner="pytest",
                command=["python", "-m", "pytest"],
                source="setup.cfg",
            )

    # Priority 4: tox.ini with pytest
    if tox_ini.exists():
        content = tox_ini.read_text()
        if "pytest" in content:
            return TestRunnerConfig(
                runner="pytest",
                command=["python", "-m", "pytest"],
                source="tox.ini",
            )

    # Priority 5: Makefile with test target
    if makefile.exists():
        content = makefile.read_text()
        if "test:" in content:
            return TestRunnerConfig(
                runner="makefile",
                command=["make", "test"],
                source="Makefile",
            )

    # Priority 6: unittest discover (if tests/ directory exists)
    if (project_dir / "tests").is_dir():
        return TestRunnerConfig(
            runner="unittest",
            command=["python", "-m", "unittest", "discover", "-s", "tests"],
            source="tests/ directory exists",
        )

    # No test runner detected
    return None
```

### 6.2 Baseline Capture

Before any changes, capture the test suite result:

```python
@dataclass
class TestBaseline:
    runner: TestRunnerConfig
    passed: bool
    test_count: int
    failure_count: int
    output: str  # Raw test output for comparison
```

Key edge case: **existing tests may already be failing.** If baseline tests
fail, the verification step checks that scaffolding does not introduce NEW
failures. The comparison is:

| Baseline | After Scaffold | Result |
|----------|---------------|--------|
| All pass | All pass | OK -- proceed |
| All pass | Any fail | FAIL -- rollback |
| N fail | N fail (same tests) | OK -- no regression |
| N fail | N+M fail | FAIL -- rollback (M new failures) |
| N fail | <N fail | OK -- scaffolding fixed something (unlikely but fine) |

### 6.3 Test Comparison Algorithm

```python
def verify_no_regressions(
    baseline: TestBaseline,
    post_scaffold: TestBaseline,
) -> VerificationResult:
    """Compare test results. Returns pass/fail with details."""

    if post_scaffold.passed:
        return VerificationResult(passed=True, message="All tests pass")

    if not baseline.passed:
        # Baseline already had failures -- check for NEW failures
        baseline_failures = parse_failure_names(baseline.output)
        post_failures = parse_failure_names(post_scaffold.output)
        new_failures = post_failures - baseline_failures

        if not new_failures:
            return VerificationResult(
                passed=True,
                message=f"No new failures ({len(post_failures)} pre-existing failures unchanged)",
            )
        return VerificationResult(
            passed=False,
            message=f"{len(new_failures)} NEW test failure(s) introduced by scaffolding",
            new_failures=list(new_failures),
        )

    # Baseline passed, but post-scaffold fails
    return VerificationResult(
        passed=False,
        message=f"{post_scaffold.failure_count} test(s) failed after scaffolding",
        new_failures=parse_failure_names(post_scaffold.output),
    )
```

### 6.4 Timeout and Safety

- Test suite timeout: 5 minutes (configurable via `--test-timeout`)
- If timeout exceeded: treat as failure, rollback
- If test runner not detected: skip verification with warning
  ("No test runner detected -- skipping zero-regression check. Verify manually.")

### 6.5 Pytest-Specific Optimizations

Since pytest is the dominant Python test runner (and alty's target), the
verification step uses pytest-specific flags when pytest is detected:

```bash
# Baseline capture (quiet, no traceback, machine-readable output)
uv run pytest --tb=no -q --co -q 2>/dev/null  # Collect test IDs only (fast)
uv run pytest --tb=no -q                       # Run tests

# Post-scaffold verification
uv run pytest --tb=short -q                    # Run with short traceback for debugging
```

The `--co -q` (collect-only) mode lets us count tests without running them,
useful for detecting if scaffolding accidentally added or removed test
collection paths.

---

## 7. Domain Language Detection (P1)

### 7.1 Approach: Heuristic Extraction, Not AST Analysis

P1 structural migration includes detecting existing domain language. This is
a heuristic process -- it suggests terms for the user to confirm, never
auto-classifies.

### 7.2 Extraction Sources

| Source | What to extract | Confidence |
|--------|----------------|------------|
| Package/module names | Potential bounded contexts or entities | High |
| Class names (from `*.py` filenames + `class X:` grep) | Potential entities/aggregates | Medium |
| README.md content | Business terms, problem domain | High |
| Docstrings (module-level `"""..."""`) | Domain concepts, workflows | Medium |
| Database model names (if Django/SQLAlchemy detected) | Entities with persistence | High |
| API endpoint names (if FastAPI/Flask detected) | Commands/queries | Medium |
| Test names (`test_*.py`, `def test_*`) | Behavioral concepts | Low |

### 7.3 Detection Algorithm

```python
def detect_domain_language(project_dir: Path) -> DomainLanguageHints:
    """Extract potential domain terms from existing codebase.

    Returns hints, NOT classifications. User must confirm.
    """
    hints = DomainLanguageHints()

    # 1. Package names as potential bounded contexts
    for package_dir in find_python_packages(project_dir / "src"):
        name = package_dir.name
        if name not in INFRASTRUCTURE_NAMES:  # Skip: utils, helpers, common, core, base
            hints.add_context_candidate(name, source=f"package: {package_dir}")

    # 2. Class names as potential entities
    for py_file in project_dir.rglob("*.py"):
        for class_name in extract_class_names(py_file):
            if class_name not in FRAMEWORK_CLASSES:  # Skip: TestCase, BaseModel, etc.
                hints.add_entity_candidate(class_name, source=str(py_file))

    # 3. README business terms
    readme = project_dir / "README.md"
    if readme.exists():
        # Extract nouns and noun phrases (simple heuristic, not NLP)
        # Look for capitalized terms, terms after "a/an/the", bullet points
        hints.add_terms_from_text(readme.read_text(), source="README.md")

    return hints
```

### 7.4 Presentation to User

Domain language hints are presented as a confirmation prompt:

```
Detected domain language (please confirm or edit):

Potential business areas (bounded contexts):
  [x] orders       (from: src/orders/ package)
  [x] inventory    (from: src/inventory/ package)
  [ ] utils        (from: src/utils/ -- likely infrastructure, excluded)

Potential entities:
  [x] Order        (from: src/orders/models.py)
  [x] Product      (from: src/inventory/models.py)
  [x] Customer     (from: src/orders/models.py)
  [ ] BaseHandler  (from: src/utils/base.py -- likely infrastructure, excluded)

Terms from README:
  [x] "order fulfillment"
  [x] "inventory management"
  [x] "shipping label"
```

The user confirms or edits. Confirmed terms seed the guided discovery session
(`alty guide`) that follows the gap analysis.

### 7.5 Infrastructure vs Domain Heuristics

To avoid presenting infrastructure names as domain concepts:

```python
# Names that are almost certainly infrastructure, not domain
INFRASTRUCTURE_NAMES = {
    "utils", "helpers", "common", "core", "base", "shared", "lib",
    "config", "settings", "middleware", "migrations", "admin",
    "management", "commands", "fixtures", "static", "templates",
    "api", "views", "serializers", "urls", "routers", "schemas",
}

FRAMEWORK_CLASSES = {
    "TestCase", "BaseModel", "Base", "Meta", "Config",
    "ModelAdmin", "AppConfig", "Migration", "Serializer",
    "APIView", "ViewSet", "Router",
}
```

---

## 8. P0 vs P1 Scope Separation

### 8.1 P0: Basic Scaffold Overlay (Phase 6 deliverable)

**Principle:** Add alty scaffolding ALONGSIDE existing code. Never move,
rename, or restructure existing files.

**What P0 does:**

| Action | Details |
|--------|---------|
| Create `.alty/` directory | Config, knowledge base, maintenance tracking |
| Create missing docs | Templates for PRD, DDD, ARCHITECTURE (empty stubs with frontmatter) |
| Create DDD directory structure | `src/domain/`, `src/application/`, `src/infrastructure/` -- ALONGSIDE existing dirs |
| Create test directory structure | `tests/domain/`, etc. -- ALONGSIDE existing test dirs |
| Copy templates | PRD, DDD Story, Architecture, beads templates |
| Copy agent personas | `.claude/agents/`, `.claude/commands/` |
| Create AGENTS.md | Agent instructions file |
| Setup IDE configs | Claude Code MCP, Cursor MCP, Antigravity rules |
| Setup Beads | `bd init` if not present |
| Update pyproject.toml | Add ruff, mypy, pytest config SECTIONS (append, not overwrite) |
| Update .gitignore | Add missing entries (append) |
| Run test verification | Zero regression gate |

**What P0 does NOT do:**

- Move existing files into DDD layers
- Rename packages or modules
- Modify existing code
- Classify subdomains
- Generate fitness functions
- Generate migration tickets
- Parse ASTs or import graphs
- Auto-detect bounded contexts

**P0 is safe because:** It only ADDS files and APPENDS to configs. The test
verification gate catches any accidental breakage.

### 8.2 P1: Structural Migration (Phase 6+ deliverable)

**Principle:** Analyze existing structure, suggest migration plan, generate
tickets. Never auto-migrate -- human approves every step.

**What P1 adds on top of P0:**

| Capability | Details | PRD Reference |
|-----------|---------|---------------|
| Domain language detection | Extract terms from code/docs (Section 7) | P1 C21 |
| Implicit bounded context detection | Map packages to potential BCs | P1 C21 |
| Anemic model identification | Detect classes with only getters/setters, no business logic | P1 C21 |
| Migration ticket generation | Ordered tickets for moving code into DDD layers | P1 C21 |
| Before/after test verification | Per-migration-step test runs | P1 C21 |
| Fitness function generation | Based on detected/confirmed BCs | P0 C14 (depends on P1 BC detection) |

### 8.3 P1 Migration Ticket Generation

When P1 detects structural issues, it generates migration tickets following the
two-tier system from the ticket pipeline design
(source: `docs/research/20260223_ticket_pipeline_design.md` Section 3).

| Gap Type | Ticket Detail Level | Rationale |
|----------|-------------------|-----------|
| Anemic model in Core subdomain | FULL | Critical migration with high risk |
| Missing domain layer | FULL | Foundational structural change |
| Missing aggregate boundaries | STANDARD | Important but less complex |
| Missing infrastructure ACL | STANDARD | Boundary definition |
| Missing `__init__.py` files | STUB | Trivial fix |
| Missing quality gate config | STUB | Config-only change |

Migration tickets follow the same beads template format as regular tickets and
are created with formal `bd dep add` dependencies:

```
Epic: Structural Migration (project-name)
  1. [FULL] Extract Order domain model from src/orders/models.py
     Deps: none (foundation)
     AC: Order entity in src/domain/models/order.py
         Existing tests still pass
         Import paths updated

  2. [FULL] Extract Inventory domain model from src/inventory/models.py
     Deps: none (foundation)
     AC: Product entity in src/domain/models/product.py
         Existing tests still pass

  3. [STANDARD] Create OrderRepository port + adapter
     Deps: 1
     AC: Protocol in src/application/ports/
         Adapter in src/infrastructure/persistence/

  4. [STUB] Add import-linter contracts for Order context
     Deps: 1, 3
```

### 8.4 P0/P1 Decision Boundary

The implementation plan is:

```
P0 (alty init --existing)
  Input:  Existing project
  Output: Scaffolded project with alty overlay
  Safety: Branch + preview + test verification + rollback
  User:   Reviews branch diff, merges manually

P1 (alty migrate) -- future command, not part of --existing
  Input:  P0-scaffolded project + confirmed domain language
  Output: Migration tickets in Beads
  Safety: Each ticket is a separate PR with its own test verification
  User:   Executes tickets one by one with AI tool assistance
```

Key insight: P1 does NOT apply migrations automatically. It generates tickets
that describe the migrations. The actual migration is done by the developer
(or AI tool) ticket by ticket, each with its own test verification.

---

## 9. Migration Ticket Integration with Two-Tier Generation

### 9.1 Alignment with Ticket Pipeline (k7m.11)

Migration tickets from P1 are structurally identical to tickets generated by
`alty generate tickets`. They use the same FULL/STANDARD/STUB templates
(source: `docs/research/20260223_ticket_pipeline_design.md` Section 3).

The difference is the input source:

| Pipeline | Input | Output |
|----------|-------|--------|
| `alty generate tickets` | `.alty/domain-model.yaml` (from DDD discovery) | New project tickets |
| P1 migration | `GapAnalysis` + confirmed domain language | Migration tickets |

### 9.2 Two-Tier Rules for Migration

Migration tickets follow the same two-tier logic as regular tickets:

**Near-term (actively blocking):** FULL detail
- Extract domain models from existing code
- Create ports/adapters for existing persistence
- Add fitness function tests for newly created boundaries

**Far-term (can wait):** STUB detail
- Move remaining utility code to proper layers
- Add comprehensive edge case tests
- Configure drift detection

The detail level is determined by the gap category:

| Gap Category | Migration Priority | Detail Level |
|-------------|-------------------|-------------|
| INCOMPATIBLE (structural) | Near-term | FULL |
| MISSING (critical docs/layers) | Near-term | FULL |
| PARTIAL (config gaps) | Near-term | STANDARD |
| MISSING (nice-to-have) | Far-term | STUB |

### 9.3 Formal Dependencies in Migration Tickets

Migration tickets get formal `bd dep add` dependencies, just like regular
pipeline tickets. The ordering follows the DDD dependency direction:

```
1. Domain model extraction (no deps -- foundation)
2. Port/Protocol creation (depends on 1)
3. Adapter implementation (depends on 2)
4. Fitness function tests (depends on 1, 2)
5. Config updates (depends on 1)
```

This ensures `bd ready` shows the correct next migration step and `bd blocked`
shows what is waiting.

---

## 10. Domain Model: GapAnalysis Aggregate

### 10.1 Value Objects and Entities

```python
# Domain layer types (conceptual -- not production code)

@dataclass(frozen=True)
class GapCategory:
    """Value Object: classification of a single gap."""
    COMPLIANT = "compliant"
    MISSING = "missing"
    PARTIAL = "partial"
    CONFLICT = "conflict"
    INCOMPATIBLE = "incompatible"

@dataclass(frozen=True)
class GapItem:
    """Value Object: one item in the gap analysis."""
    path: str                    # Relative path (e.g., "docs/PRD.md")
    category: str                # GapCategory value
    layer: str                   # infrastructure | documentation | source | tool | quality
    priority: str                # p0 | p1
    description: str             # Human-readable description
    action: str                  # What the scaffolder will do (create | append | rename | skip | report)
    existing_content: str | None # For PARTIAL/CONFLICT, what exists

@dataclass(frozen=True)
class ProjectScan:
    """Value Object: snapshot of existing project state."""
    has_git: bool
    is_clean: bool
    python_version: str | None
    detected_tools: list[str]          # ["claude-code", "cursor", "antigravity"]
    existing_files: set[str]           # Relative paths of all files
    existing_dirs: set[str]            # Relative paths of all directories
    pyproject_content: str | None      # Raw content of pyproject.toml
    gitignore_content: str | None      # Raw content of .gitignore
    test_runner: TestRunnerConfig | None
    readme_word_count: int

@dataclass(frozen=True)
class TestRunnerConfig:
    """Value Object: detected test runner."""
    runner: str          # "pytest" | "unittest" | "makefile" | "tox"
    command: list[str]   # Full command to run tests
    source: str          # Where the detection came from

@dataclass(frozen=True)
class DomainLanguageHints:
    """Value Object: extracted domain terms (P1 only)."""
    context_candidates: list[tuple[str, str]]   # (name, source)
    entity_candidates: list[tuple[str, str]]     # (name, source)
    terms_from_docs: list[tuple[str, str]]       # (term, source)

@dataclass
class GapAnalysis:
    """Aggregate Root: complete gap analysis for one project."""
    project_name: str
    scan: ProjectScan
    gaps: list[GapItem]
    readiness_score: float          # 0.0 - 1.0
    domain_hints: DomainLanguageHints | None  # P1 only

    # --- Invariants ---
    # 1. Every golden state item has exactly one GapItem
    # 2. Readiness score = compliant_count / total_count
    # 3. No GapItem can have action "create" if category is "compliant"
    # 4. INCOMPATIBLE items always have action "report" (never auto-fix)

    def items_by_category(self, category: str) -> list[GapItem]:
        """Filter gaps by category."""
        return [g for g in self.gaps if g.category == category]

    def items_for_scaffolding(self) -> list[GapItem]:
        """Return items that the scaffolder will act on (MISSING + PARTIAL)."""
        return [g for g in self.gaps if g.action in ("create", "append")]

    def summary(self) -> dict[str, int]:
        """Category counts."""
        from collections import Counter
        return dict(Counter(g.category for g in self.gaps))
```

### 10.2 RescuePort Interface

```python
# src/application/ports/rescue_port.py
from __future__ import annotations
from typing import Protocol

class RescuePort(Protocol):
    def scan_project(self, project_dir: str) -> ProjectScan:
        """Scan existing project structure and environment."""
        ...

    def analyze_gaps(self, scan: ProjectScan) -> GapAnalysis:
        """Compare scan against golden state, produce gap report."""
        ...

    def detect_domain_language(self, project_dir: str) -> DomainLanguageHints:
        """Extract domain language hints from code/docs (P1)."""
        ...

    def scaffold(self, analysis: GapAnalysis, project_dir: str) -> ScaffoldResult:
        """Apply scaffolding changes based on gap analysis."""
        ...

    def verify_tests(
        self,
        runner: TestRunnerConfig,
        baseline: TestBaseline,
    ) -> VerificationResult:
        """Run tests and compare against baseline."""
        ...

    def rollback(self, project_dir: str, original_branch: str) -> None:
        """Roll back all changes and delete alty/init branch."""
        ...
```

---

## 11. Competitive Analysis: How Others Handle Brownfield

### 11.1 Spec-Kit (GitHub, 2025-2026)

Spec-Kit's brownfield bootstrap extension scans project structure to identify
tech stack, frameworks, and patterns. It produces:
- `.specify/memory/constitution.md` (project principles from code analysis)
- `.specify/templates/spec-template.md` (tech-stack-specific spec template)
- `.specify/templates/plan-template.md` (customized planning template)
- `.specify/templates/tasks-template.md` (task template with TDD enforcement)

Key principles: respect existing architecture, lock tech stack, search for
reusable code, match existing style conventions.

Source: [GitHub spec-kit issue #1436](https://github.com/github/spec-kit/issues/1436)

### 11.2 BMAD Method (2025-2026)

BMAD Method for brownfield requires documenting the existing system first using
the Analyst agent's `document-project` task, then creating brownfield PRDs,
architecture adapters, and tight QA/regression coverage. Uses a "flattener"
tool to convert existing projects into AI-consumable XML format.

Source: [BMAD-METHOD DeepWiki](https://deepwiki.com/bmad-code-org/BMAD-METHOD/3.5-brownfield-development)

### 11.3 Cruft (2024-2025)

Cruft tracks which cookiecutter template generated a project and can update the
project when the template changes. Uses `.cruft.json` for template tracking.
Does NOT do gap analysis -- it assumes the project was initially generated from
the template.

Source: [Cruft GitHub](https://github.com/cruft/cruft)

### 11.4 Classical Scaffolding Tools (Patterns Worth Adopting)

**Django `inspectdb`** generates Python models from an existing database schema,
marking them as auto-generated and requiring manual cleanup. alty adopts this
"generate-then-refine" pattern in P1 domain language detection (Section 7):
extract candidates, present for human confirmation, refine.
Source: [Django legacy databases docs](https://docs.djangoproject.com/en/6.0/howto/legacy-databases/)

**Rails `rails new --skip-*`** uses selective scaffolding flags (skip-git,
skip-test, skip-bundle) to avoid generating files the project already has. alty
achieves the same effect via gap categories: COMPLIANT items are automatically
skipped, no flags needed.
Source: [Rails guides](https://guides.rubyonrails.org/getting_started.html)

**Yeoman generators** have built-in per-file conflict resolution -- when a
generated file conflicts with an existing one, it prompts: overwrite, skip,
show diff, or force. alty uses a stricter policy: never overwrite, create
`_alty` suffixed copies for conflicts (PRD section 6 file safety rules).
This is safer for AI-tool-assisted workflows where the human may not be
watching every prompt.
Source: [Yeoman docs](https://yeoman.io/learning/)

### 11.5 alty Differentiators

| Feature | Spec-Kit | BMAD | Cruft | alty |
|---------|----------|------|-------|-----------|
| Gap analysis report | No | No | No | Yes |
| Branch safety | No | No | No | Yes (hard gate) |
| Test verification | No | QA coverage | No | Zero regression gate |
| Domain language detection | No | Document-project task | No | Yes (P1) |
| Persona-adapted output | No | No | No | Yes |
| Migration ticket generation | No | PRD generation | No | Yes (P1, two-tier) |
| Rollback on failure | No | No | Template rollback | Full branch rollback |

alty's key differentiator is the combination of gap analysis + test
verification + branch safety + persona-adapted output. No competitor has all
four.

---

## 12. Risks and Mitigations

| Risk | Severity | Mitigation |
|------|----------|------------|
| Test runner detection fails (custom runner) | Medium | Allow `--test-command "custom cmd"` override |
| Large projects slow to scan | Low | Scan is file-existence checks, not content analysis; typically <1s |
| pyproject.toml append breaks existing config | Medium | Parse TOML properly; only add sections that don't exist; use tomllib for reading, tomli-w for writing |
| Existing `src/` layout is deeply incompatible | Medium | P0 adds DDD dirs alongside; P1 generates migration tickets |
| Domain language detection produces noise | Low | Always human-confirmed; show confidence scores; allow dismissal |
| User merges branch before reviewing | Low | Not our problem -- we create the branch, user owns the merge |
| Scaffolding adds files that confuse existing test discovery | Medium | Test verification catches this; pytest `testpaths` config respected |

---

## 13. Follow-Up Implementation Tickets

These tickets would be created for the Phase 6 implementation:

### P0 Tickets (Basic Scaffold Overlay)

| # | Title | Type | Detail Level | Depends On |
|---|-------|------|-------------|------------|
| 1 | Define GapAnalysis domain model (GapItem, ProjectScan, GapCategory) | Task | FULL | -- |
| 2 | Implement golden state manifest (file checklist) | Task | FULL | 1 |
| 3 | Implement ProjectScanner (environment + file structure + config scan) | Task | FULL | 1, 2 |
| 4 | Implement gap classification algorithm | Task | FULL | 1, 2, 3 |
| 5 | Implement test runner detection (pytest, unittest, makefile, tox) | Task | STANDARD | 1 |
| 6 | Implement test verification with baseline comparison | Task | FULL | 5 |
| 7 | Implement branch workflow (create, rollback, force-branch) | Task | STANDARD | -- |
| 8 | Implement scaffolder (apply changes based on GapAnalysis) | Task | FULL | 1, 4, 7 |
| 9 | Implement gap report formatter (terminal output, persona-adapted) | Task | STANDARD | 1, 4 |
| 10 | Implement Markdown report generation (--report flag) | Task | STUB | 9 |
| 11 | Wire `alty init --existing` CLI command (Typer adapter) | Task | STUB | 3, 4, 6, 7, 8, 9 |
| 12 | Integration tests: full --existing workflow end-to-end | Task | FULL | 11 |

### P1 Tickets (Structural Migration -- deferred)

| # | Title | Type | Detail Level | Depends On |
|---|-------|------|-------------|------------|
| 13 | Implement domain language detection (package/class/README extraction) | Task | STANDARD | P0 complete |
| 14 | Implement bounded context candidate detection | Task | STANDARD | 13 |
| 15 | Implement anemic model detection heuristics | Task | STANDARD | 13 |
| 16 | Implement migration ticket generation (FULL/STANDARD/STUB templates) | Task | FULL | 14, 15 |
| 17 | Wire `alty migrate` CLI command | Task | STUB | 16 |

---

## Recommendation

**For P0 implementation:**

Build the gap analysis in 3 phases:

1. **Domain model + golden state manifest** (tickets 1-2) -- Define what "fully
   seeded" means as a data structure. This is the foundation for everything else.

2. **Scanner + classifier + scaffolder** (tickets 3-4, 8) -- The core gap analysis
   logic. Scanner is read-only; scaffolder is write-only. They share the
   GapAnalysis aggregate as the intermediary.

3. **Safety net** (tickets 5-7, 11-12) -- Test verification, branch workflow,
   CLI wiring. These can be developed in parallel with phase 2.

**For P1:** Defer until P0 is shipped and validated on real projects. P1 requires
AST analysis and import graph traversal, which are significantly more complex.
The domain language detection (Section 7) is a good spike candidate before full
P1 implementation.

**Critical design principle:** The scanner NEVER modifies files. The scaffolder
ONLY creates/appends. The test verification is the safety gate between them.
Separation of concerns makes rollback clean and debugging straightforward.

---

## Sources

- alty PRD: `docs/PRD.md` (Section 4 Scenario 2, Section 5 P0 C6/C7, P1 C21, Section 6 File Safety Rules)
- alty DDD: `docs/DDD.md` (Section 4 Rescue bounded context, Section 3 subdomain classification)
- CLI/MCP design: `docs/research/20260222_cli_mcp_design.md` (Section 2 command tree, RescuePort)
- Ticket pipeline design: `docs/research/20260223_ticket_pipeline_design.md` (Section 3 FULL/STANDARD/STUB templates, Section 4 BeadsWriterPort)
- Ripple review design: `docs/research/20260223_ripple_review_design.md` (Section 1 context diff format)
- Existing `alty init --existing` implementation: `bin/alty` (lines 673-806)
- [Spec-Kit brownfield bootstrap proposal](https://github.com/github/spec-kit/issues/1436)
- [BMAD Method brownfield development](https://deepwiki.com/bmad-code-org/BMAD-METHOD/3.5-brownfield-development)
- [Cruft -- update existing projects from cookiecutter templates](https://github.com/cruft/cruft)
- [Python testing guide -- Real Python](https://realpython.com/python-testing/)
- [Django legacy databases / inspectdb](https://docs.djangoproject.com/en/6.0/howto/legacy-databases/)
- [Yeoman -- web scaffolding tool](https://yeoman.io/learning/)
- [Rails guides -- getting started](https://guides.rubyonrails.org/getting_started.html)
