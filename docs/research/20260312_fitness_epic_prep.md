# Research: Architecture Fitness Functions Epic Prep

**Date:** 2026-03-12
**Spike Ticket:** alto-cli-cyf
**Status:** Final

## Summary

The fitness context exists but generates **Python output** (import-linter TOML + pytestarch tests) while the project is now **Go**. DDD.md references depguard but **depguard is GPL-3.0**, which poses license risks for commercial users of alto. The epic should refactor fitness generation to produce Go-compatible output using **arch-go (MIT)** — a tool that provides both positive (`shouldOnlyDependsOn`) and negative (`shouldNotDependsOn`) dependency rules with compliance thresholds.

## Research Questions Answered

### Q1: Current state of fitness context in the codebase?

**Files audited:**

| File | Purpose | Status |
|------|---------|--------|
| `internal/fitness/domain/fitness_test_suite.go` | Aggregate root, generates contracts | Generates Python output |
| `internal/fitness/domain/fitness_values.go` | Contract, ArchRule, ContractType | Python-specific types |
| `internal/fitness/domain/bounded_context_canvas.go` | BoundedContextCanvas VO (ddd-crew v5) | Reusable |
| `internal/fitness/application/fitness_generation_handler.go` | Handler with preview pattern | Good pattern, wrong output |
| `internal/fitness/application/ports.go` | FitnessGeneration, GateRunner, QualityGateChecker | Needs Go-specific port |
| `internal/fitness/infrastructure/subprocess_gate_runner.go` | Runs quality gates | Defaults to PythonUvProfile |
| `internal/shared/domain/valueobjects/stack_profile.go` | StackProfile interface | Only PythonUvProfile exists |

**Critical gap:** `RenderImportLinterTOML()` and `RenderPytestarchTests()` produce Python output. No `RenderDepguardYAML()` or `RenderGoArchTests()` methods exist.

### Q2: Python→Go mismatch details

**DDD.md Story 4 (lines 106-129) specifies:**
```
4. alto CLI generates depguard rules in .golangci.yml
5. alto CLI generates architecture test files per bounded context
8. Tests run as part of quality gates (go vet + golangci-lint + go test -race + fitness functions)
```

**Current implementation produces:**
- `.importlinter` TOML file (Python import-linter)
- `tests/architecture/test_fitness.py` (Python pytestarch)

**Existing .golangci.yml has MANUAL depguard rules:**
```yaml
depguard:
  rules:
    domain:
      files:
        - "**/internal/**/domain/**/*.go"
      deny:
        - pkg: "github.com/alto-cli/alto/internal/**/application"
          desc: "Domain layer must not import application layer"
        - pkg: "github.com/alto-cli/alto/internal/**/infrastructure"
          desc: "Domain layer must not import infrastructure layer"
```

These rules are correct but not auto-generated from DDD.md bounded context map.

### Q3: PRD/DDD.md/ARCHITECTURE.md freshness

**alto doc-health output:**
- PRD.md: OK
- DDD.md: OK
- ARCHITECTURE.md: OK
- 30+ research docs: missing frontmatter (not blocking)

**DDD.md bounded context list (10 contexts):**

| Context | Classification | Treatment |
|---------|---------------|-----------|
| Guided Discovery | Core | Hexagonal, strict fitness |
| Architecture Testing | Core | Hexagonal, strict fitness |
| Ticket Pipeline | Core | Hexagonal, strict fitness |
| Ticket Freshness | Core | Hexagonal, strict fitness |
| Tool Translation | Supporting | Layered, moderate fitness |
| Knowledge Base | Supporting | Layered, moderate fitness |
| Rescue Mode | Supporting | Layered, moderate fitness |
| File Generation | Generic | ACL wrapper |
| CLI Framework | Generic | ACL wrapper |
| MCP Server Framework | Generic | ACL wrapper |

**Finding:** DDD.md already specifies Go tooling (depguard, go test). No updates needed.

### Q4: What Go architecture testing patterns exist?

**From Context7 research (2026-03-12):**

| Tool | Type | License | Pros | Cons | Maturity |
|------|------|---------|------|------|----------|
| **depguard** | golangci-lint linter | **GPL-3.0** | Integrated, file glob patterns, deny rules | GPL license risk for commercial users, deny-only model | High (part of golangci-lint) |
| **go-arch-lint** | Dedicated tool | MIT | Semantic YAML config, `mayDependOn` rules, graph generation | Allow-only model (no deny), separate tool | High |
| **arch-go** | Dedicated tool | **MIT** | BOTH `shouldOnlyDependsOn` AND `shouldNotDependsOn`, compliance thresholds (0-100%), HTML reports | Separate tool | High (v2.1.2, Feb 2026) |
| **Custom Go tests** | DIY | N/A | Full flexibility | Must implement AST analysis | N/A |

**arch-go configuration example (recommended):**
```yaml
version: 1
threshold:
  compliance: 100
  coverage: 100

dependenciesRules:
  # Positive rules: domain may ONLY depend on domain + stdlib
  - package: "**.domain.**"
    shouldOnlyDependsOn:
      internal:
        - "**.domain.**"
      external:
        - "$gostd"

  # Negative rules: domain must NOT depend on outer layers
  - package: "**.domain.**"
    shouldNotDependsOn:
      internal:
        - "**.application.**"
        - "**.infrastructure.**"

  # Application layer
  - package: "**.application.**"
    shouldNotDependsOn:
      internal:
        - "**.infrastructure.**"
```

**Why arch-go over depguard + go-arch-lint:**
1. **MIT license** — enterprise-friendly, no GPL contamination risk
2. **Both rule types** — `shouldOnlyDependsOn` (positive) + `shouldNotDependsOn` (negative)
3. **Compliance thresholds** — essential for `alto init --existing` (brownfield adoption)
4. **Single tool** — no need for two tools with different configs

### Q5: arch-go auto-generation from DDD.md?

**Auto-generation approach:**

Given DDD.md bounded contexts, generate `arch-go.yml`:

```yaml
# Auto-generated from DDD.md bounded contexts
# Regenerate with: alto fitness generate
version: 1
threshold:
  compliance: 100  # Strict for greenfield
  coverage: 100

dependenciesRules:
  # === Universal Layer Rules ===
  # Domain layer: inward-only dependencies
  - package: "**.domain.**"
    shouldOnlyDependsOn:
      internal:
        - "**.domain.**"
        - "**.shared.domain.**"
      external:
        - "$gostd"
    shouldNotDependsOn:
      internal:
        - "**.application.**"
        - "**.infrastructure.**"

  # Application layer: domain + shared allowed
  - package: "**.application.**"
    shouldNotDependsOn:
      internal:
        - "**.infrastructure.**"

  # === Cross-Context Isolation ===
  # Generated per bounded context from context map
  - package: "github.com/{project}/internal/bootstrap/**"
    shouldNotDependsOn:
      internal:
        - "github.com/{project}/internal/discovery/**"  # No relationship
        - "github.com/{project}/internal/challenge/**"  # No relationship
```

**For brownfield projects (`alto init --existing`):**
```yaml
threshold:
  compliance: 80   # Allow gradual adoption
  coverage: 80
```

## Options Considered

| Option | License | Pros | Cons |
|--------|---------|------|------|
| **A: Depguard only** | GPL-3.0 | Already in golangci-lint | GPL license risk, deny-only model |
| **B: Depguard + go-arch-lint** | GPL + MIT | Best of both models | Two tools, GPL contamination |
| **C: go-arch-lint only** | MIT | Semantic positive model | Allow-only (no deny rules) |
| **D: arch-go only** | **MIT** | Both models, thresholds, single tool | Separate from golangci-lint |
| **E: Custom Go arch tests** | MIT | Full flexibility | Must implement AST analysis |

## Recommendation

**Option D: arch-go only** (MIT)

Rationale:
1. **MIT license** — no GPL contamination risk for commercial users of alto
2. **Both rule types** — `shouldOnlyDependsOn` (like go-arch-lint) + `shouldNotDependsOn` (like depguard)
3. **Compliance thresholds** — essential for brownfield adoption (`alto init --existing`)
4. **Single tool** — simpler than depguard + go-arch-lint combo
5. **Active maintenance** — v2.1.2 released Feb 2026
6. **DDD-friendly patterns** — `**.domain.**` wildcards cover all bounded contexts

**Trade-off:** Not integrated into golangci-lint, requires separate CI step. This is acceptable because:
- golangci-lint still runs for other linters (errcheck, staticcheck, etc.)
- arch-go runs as a separate command: `arch-go` (exit code 0/1 for CI)
- The license safety is more important than lint integration

## Implementation Approach

### Phase 1: GoModProfile (stack_profile.go)
- Add `GoModProfile` implementing StackProfile
- `QualityGateCommands()` returns `go build`, `go vet`, `golangci-lint run`, `go test -race`
- `FitnessAvailable()` returns true
- `FitnessCommands()` returns `golangci-lint run` + `arch-go`

### Phase 2: arch-go Config Generator
- New method: `FitnessTestSuite.RenderArchGoYAML() (string, error)`
- Generates `arch-go.yml` with:
  - `threshold:` compliance/coverage levels (100% for greenfield, 80% for brownfield)
  - `dependenciesRules:` with `shouldOnlyDependsOn` for positive rules
  - `dependenciesRules:` with `shouldNotDependsOn` for negative rules
  - Cross-context isolation rules from context map
- Maps Contract types to arch-go dependency rules

### Phase 3: Remove depguard dependency
- Update DDD.md to reference arch-go instead of depguard
- Remove manual depguard rules from .golangci.yml (keep other linters)
- Document the GPL license rationale

### Phase 4: CLI Commands
- `alto fitness generate [--preview]` — generates `arch-go.yml`
- `alto fitness generate --brownfield` — generates with 80% thresholds
- `alto check --fitness` — runs `golangci-lint run` + `arch-go`

## Follow-up Tickets

| # | Title | Type | Dependencies |
|---|-------|------|--------------|
| 1 | Add GoModProfile to stack_profile.go | Task | None |
| 2 | Implement RenderArchGoYAML in FitnessTestSuite | Task | #1 |
| 3 | Remove depguard from .golangci.yml (GPL→MIT migration) | Task | #2 |
| 4 | Update FitnessGenerationHandler for Go output | Task | #3 |
| 5 | Add `alto fitness generate` CLI command | Task | #4 |
| 6 | Update `alto check --fitness` for Go (arch-go) | Task | #5 |
| 7 | Create bounded_context_map.yaml schema and parser | Task | None |
| 8 | Generate alto's own fitness configs from BC map | Task | #7, #2 |

## References

- Existing fitness design: `docs/research/20260223_fitness_function_design.md`
- DDD.md Story 4: lines 106-129
- arch-go docs: https://github.com/arch-go/arch-go (MIT, v2.1.2, Feb 2026)
- Background research: `docs/research/20260312_go_architecture_testing_ddd_layer_enforcement.md`

## License Decision Rationale

**Why not depguard (GPL-3.0)?**
- alto is a tool that generates configs for user projects
- If users adopt depguard via alto's recommendations, their CI pipelines run GPL-3.0 software
- Some enterprises have blanket GPL restrictions, even for tooling
- arch-go provides equivalent capability under MIT license

**Why not go-arch-lint + depguard?**
- Two tools means two configs, two CI steps
- Still has GPL exposure via depguard
- arch-go alone provides both positive AND negative rules
