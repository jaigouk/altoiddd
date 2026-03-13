---
date: 2026-03-12
author: researcher
status: complete
topic: Go Architecture Testing Tools for DDD Layer Enforcement
---

# Go Architecture Testing Tools for DDD Layer Enforcement (2026)

## 1. Research Questions

1. What is the current state of depguard in golangci-lint v2? Can it be auto-generated from a bounded context map?
2. Does go-arch-lint exist and is it mature? What features does it have?
3. Are there other Go architecture testing tools in 2026?
4. How do DDD Go projects enforce layer boundaries (domain cannot import infrastructure)?
5. Best practices for generating depguard config from DDD.md bounded context definitions

## 2. Context

alty generates DDD project scaffolding including bounded context maps, layer structure
(`domain/`, `application/`, `infrastructure/`), and architectural docs. A key capability
(DDD.md line 16, step 16) is auto-generating fitness function tests from bounded context maps.

The project already uses depguard via golangci-lint v2 (`.golangci.yml` lines 44-79).
This research evaluates whether the current approach is sufficient or if additional/alternative
tools would improve enforcement, especially for auto-generation from DDD artifacts.

**Constraints (from `docs/PRD.md`):**
- Go 1.26+, pure Go preferred (no CGO)
- No cloud dependencies -- everything runs locally
- No paid APIs required
- Permissive licenses preferred for project dependencies (MIT, Apache 2.0, BSD)
- Tool dependencies (linters, CI) can be GPL since they don't link into the binary

---

## 3. Tool Landscape

### 3.1 depguard (via golangci-lint)

| Property | Value |
|---|---|
| Current version | v2.2.1 (standalone); integrated in golangci-lint v2.11.3 |
| License | GPL-3.0 (tool-only -- does not affect project license) |
| Go support | Go 1.26+ compatible (via golangci-lint) |
| CGO required | No |
| Maintenance | Active; integrated as default linter in golangci-lint v2 `standard` preset |
| Source | [github.com/OpenPeeDeeP/depguard](https://github.com/OpenPeeDeeP/depguard) |

**How it works:** Configured within `.golangci.yml` under `linters.settings.depguard`.
Defines named rules with file glob patterns and deny/allow lists. Each rule targets
specific file paths and blocks imports of specific packages.

**Current alty config** (`.golangci.yml` lines 44-79):
```yaml
depguard:
  rules:
    domain:
      files:
        - "**/internal/**/domain/**/*.go"
      deny:
        - pkg: "github.com/alty-cli/alty/internal/**/application"
          desc: "Domain layer must not import application layer"
        - pkg: "github.com/alty-cli/alty/internal/**/infrastructure"
          desc: "Domain layer must not import infrastructure layer"
        - pkg: "github.com/spf13/cobra"
          desc: "Domain layer must not import CLI framework"
        # ... more framework denials

    application:
      files:
        - "**/internal/**/application/**/*.go"
      deny:
        - pkg: "github.com/alty-cli/alty/internal/**/infrastructure"
          desc: "Application layer must not import infrastructure layer"
        - pkg: "github.com/spf13/cobra"
          desc: "Application layer must not import CLI framework"
```

**Strengths:**
- Already integrated -- zero additional tooling needed
- Violations are lint errors (CI-blocking)
- Wildcard `**` patterns cover all bounded contexts with a single rule
- Part of the golangci-lint `standard` preset in v2

**Limitations:**
- Configuration is static YAML -- no auto-generation capability
- Cannot express "component A may only depend on component B" (only deny lists)
- No dependency graph visualization
- Glob pattern matching for files can be brittle with nested packages
- Cannot enforce cross-context boundaries (e.g., "bootstrap cannot import discovery")

**Auto-generation potential:** HIGH. The deny-list structure maps directly to DDD layer rules:
- Parse bounded contexts from DDD.md
- For each context, generate domain/application/infrastructure rules
- Template is identical across contexts (only package paths change)

### 3.2 go-arch-lint

| Property | Value |
|---|---|
| Current version | v1.14.0 (released 2025-11-13) |
| License | MIT |
| Go support | Go 1.25+ |
| CGO required | No |
| Maintenance | Active; 169 commits, regular releases through 2025 |
| Source | [github.com/fe3dback/go-arch-lint](https://github.com/fe3dback/go-arch-lint) |

**How it works:** Standalone CLI tool. Reads a `.go-arch-lint.yml` config that maps
Go packages to named "components" and defines allowed dependency relationships between
components. Checks actual import graph against declared rules.

**Config example for DDD hexagonal:**
```yaml
version: 3
workdir: internal

components:
  # Per-context layers
  bootstrap-domain:       { in: bootstrap/domain/** }
  bootstrap-application:  { in: bootstrap/application/** }
  bootstrap-infra:        { in: bootstrap/infrastructure/** }

  discovery-domain:       { in: discovery/domain/** }
  discovery-application:  { in: discovery/application/** }
  discovery-infra:        { in: discovery/infrastructure/** }

  # Shared kernel
  shared-domain:          { in: shared/domain/** }
  shared-application:     { in: shared/application/** }
  shared-infra:           { in: shared/infrastructure/** }

  # Composition root
  composition:            { in: composition/** }

commonComponents:
  - shared-domain

deps:
  bootstrap-domain:
    mayDependOn: []  # domain has ZERO deps (except commonComponents)

  bootstrap-application:
    mayDependOn:
      - bootstrap-domain

  bootstrap-infra:
    mayDependOn:
      - bootstrap-domain
      - bootstrap-application

  composition:
    mayDependOn:
      - bootstrap-domain
      - bootstrap-application
      - bootstrap-infra
      # ... all contexts
```

**Strengths:**
- Positive dependency model ("may depend on") vs negative (deny lists)
- Semantic component names map directly to DDD terminology
- Dependency graph generation (`go-arch-lint graph`)
- Catches ANY undeclared dependency, not just explicitly denied ones
- `commonComponents` maps to shared kernel concept
- Config is closer to how architects think about boundaries

**Limitations:**
- Separate tool -- not integrated into golangci-lint (separate CI step)
- Verbose config when many bounded contexts exist (each needs 3 component entries)
- No golangci-lint integration means different error format
- Less ecosystem adoption than depguard

**Auto-generation potential:** VERY HIGH. The positive dependency model maps perfectly:
- Each bounded context generates 3 components (domain, application, infrastructure)
- Dependency rules are uniform per layer pattern
- `commonComponents` = shared kernel
- Composition root gets "depends on everything" rule

### 3.3 arch-go

| Property | Value |
|---|---|
| Current version | v2.1.2 (released 2026-02-03) |
| License | MIT |
| Go support | Modern Go (modules required) |
| CGO required | No |
| Maintenance | Active; v2 released Feb 2026 |
| Source | [github.com/arch-go/arch-go](https://github.com/arch-go/arch-go) |

**How it works:** Reads an `arch-go.yml` config defining dependency rules, content rules,
function rules, and naming rules. Can run as CLI or programmatically within Go tests.
Generates HTML/JSON compliance reports with threshold enforcement.

**Config example for DDD:**
```yaml
version: 1
threshold:
  compliance: 100
  coverage: 100

dependenciesRules:
  - package: "**.domain.**"
    shouldNotDependsOn:
      internal:
        - "**.application.**"
        - "**.infrastructure.**"
    shouldOnlyDependsOn:
      internal:
        - "**.domain.**"
      external:
        - "$gostd"

  - package: "**.application.**"
    shouldNotDependsOn:
      internal:
        - "**.infrastructure.**"
```

**Strengths:**
- Both positive AND negative rules (`shouldOnlyDependsOn` + `shouldNotDependsOn`)
- Compliance thresholds (can introduce rules gradually: 80% -> 100%)
- Beyond dependencies: content rules (require interfaces in certain packages),
  function rules (max params, max lines), naming rules
- HTML reports for architecture documentation
- Can be embedded in Go test files (programmatic API)

**Limitations:**
- Less community adoption than depguard
- No golangci-lint integration
- HTML report generation adds complexity
- Config syntax is more verbose than go-arch-lint

**Auto-generation potential:** HIGH. Pattern-based rules work well:
- `**.domain.**` covers all contexts at once (like depguard wildcards)
- Threshold feature useful for brownfield projects alty handles (`init --existing`)
- Content rules could enforce "domain packages must have value objects"

### 3.4 go-cleanarch

| Property | Value |
|---|---|
| Current version | v1.2.1 (released 2021-02-18) |
| License | MIT |
| Go support | Older Go versions |
| CGO required | No |
| Maintenance | **ABANDONED** -- no commits since 2021 |
| Source | [github.com/roblaszczak/go-cleanarch](https://github.com/roblaszczak/go-cleanarch) |
| Author | Robert Laszczak (Three Dots Labs / Watermill author) |

**How it works:** Convention-based. Recognizes layers by package name
(domain, application, interfaces, infrastructure). Enforces dependency rule automatically.

**Strengths:**
- Zero configuration -- works if you follow naming conventions
- Written by the Watermill author (same ecosystem as alty)

**Limitations:**
- **Not maintained since 2021** -- 4+ years stale
- Fixed layer names (cannot customize)
- No bounded context awareness
- No cross-context boundary enforcement
- No golangci-lint integration

**Auto-generation potential:** N/A -- tool is convention-based, no config to generate.

**Verdict:** Do not use. Abandoned. Historical interest only.

---

## 4. Options Comparison Table

| Criterion | depguard | go-arch-lint | arch-go | go-cleanarch |
|-----------|----------|-------------|---------|-------------|
| **Latest release** | 2023 (standalone) / 2026 (via golangci-lint v2.11.3) | 2025-11-13 (v1.14.0) | 2026-02-03 (v2.1.2) | 2021-02-18 |
| **License** | GPL-3.0 (tool) | MIT | MIT | MIT |
| **golangci-lint integration** | Native | None | None | None |
| **Dependency model** | Deny list | Allow list (mayDependOn) | Both (shouldOnly + shouldNot) | Convention |
| **Cross-context boundaries** | Manual deny rules | Component-level deps | Package pattern rules | No |
| **Dependency graph viz** | No | Yes (`graph` command) | No (HTML report) | No |
| **Compliance thresholds** | No (binary pass/fail) | No | Yes (0-100%) | No |
| **Content/naming rules** | No | No | Yes | No |
| **Config complexity** | Low (YAML in golangci.yml) | Medium (separate YAML) | Medium (separate YAML) | Zero |
| **Auto-gen difficulty** | Easy (template deny lists) | Medium (component mapping) | Easy (pattern rules) | N/A |
| **CI integration** | Built-in (golangci-lint) | Separate step | Separate step | Separate |
| **Maturity** | High (standard in golangci-lint) | Medium-High | Medium | Dead |

---

## 5. Analysis: How DDD Go Projects Enforce Layer Boundaries in 2026

Based on research across Go DDD community patterns:

### Pattern 1: depguard-only (most common)
Projects using golangci-lint add depguard deny rules per layer. This is the dominant
approach because it requires no additional tooling. Wildcard patterns (`**/domain/**`)
cover all bounded contexts uniformly.

**Used by:** Most DDD Go projects found in 2025-2026 blog posts and tutorials.

### Pattern 2: depguard + go-arch-lint (defense in depth)
depguard catches the obvious violations at lint time. go-arch-lint adds positive
dependency modeling and graph visualization. Used by teams that want architectural
documentation alongside enforcement.

### Pattern 3: Go test-based (arch-go or custom)
Some projects write architecture tests in Go that run as part of `go test`. arch-go
supports this via its programmatic API. This approach is familiar to Java teams coming
from ArchUnit.

### Pattern 4: Go compiler as enforcer (package visibility)
The simplest approach: use Go's `internal/` package convention. Packages under
`internal/bootstrap/` are invisible outside `bootstrap/`. However, this only enforces
external visibility, not inward dependency flow (domain importing infrastructure).

**Recommendation for alty:** Pattern 2 (depguard + go-arch-lint) provides the best
combination for a project that generates DDD scaffolding for others. depguard gives
CI-integrated lint enforcement; go-arch-lint provides the semantic model that maps
to DDD concepts and generates dependency graphs for documentation.

---

## 6. Auto-Generating depguard Config from DDD Bounded Context Map

### 6.1 Current State

No tool exists that auto-generates depguard config from a bounded context map.
This is a greenfield capability that alty would pioneer.

### 6.2 Generation Strategy

**Input:** `docs/DDD.md` bounded context definitions (section 3: Bounded Context Map)

**Output:** `.golangci.yml` depguard rules section

**Algorithm:**
1. Parse DDD.md for bounded context names (e.g., `bootstrap`, `discovery`, `challenge`)
2. Parse subdomain classification (Core, Supporting, Generic) for strictness levels
3. For each context, generate 3 rule sets (domain, application, infrastructure)
4. Generate cross-context isolation rules based on context map relationships
5. Generate shared kernel allowances

### 6.3 Template for Per-Context Rules

```yaml
# Auto-generated from DDD.md bounded contexts
# DO NOT EDIT MANUALLY -- regenerate with: alty fitness generate

depguard:
  rules:
    # === Universal Layer Rules ===
    # These apply to ALL bounded contexts via wildcard patterns

    domain-layer:
      files:
        - "**/internal/**/domain/**/*.go"
      deny:
        - pkg: "github.com/MODULE/internal/**/application"
          desc: "Domain layer must not import application layer (DDD inward dependency rule)"
        - pkg: "github.com/MODULE/internal/**/infrastructure"
          desc: "Domain layer must not import infrastructure layer (DDD inward dependency rule)"
        # Framework isolation
        - pkg: "github.com/spf13/cobra"
          desc: "Domain layer must not import CLI framework"
        - pkg: "github.com/ThreeDotsLabs/watermill"
          desc: "Domain layer must not import event bus framework"
        - pkg: "database/sql"
          desc: "Domain layer must not import database packages"
        - pkg: "net/http"
          desc: "Domain layer must not import HTTP packages"
        - pkg: "os/exec"
          desc: "Domain layer must not import process execution"

    application-layer:
      files:
        - "**/internal/**/application/**/*.go"
      deny:
        - pkg: "github.com/MODULE/internal/**/infrastructure"
          desc: "Application layer must not import infrastructure layer (DDD inward dependency rule)"
        - pkg: "github.com/spf13/cobra"
          desc: "Application layer must not import CLI framework"
        - pkg: "database/sql"
          desc: "Application layer must not import database packages directly"

    # === Cross-Context Isolation Rules ===
    # Generated from DDD.md context map relationships

    # bootstrap context: Supporting subdomain
    bootstrap-isolation:
      files:
        - "**/internal/bootstrap/**/*.go"
      deny:
        - pkg: "github.com/MODULE/internal/discovery"
          desc: "bootstrap must not import discovery (no declared relationship in context map)"
        - pkg: "github.com/MODULE/internal/challenge"
          desc: "bootstrap must not import challenge (no declared relationship in context map)"
        # ... one deny per non-related context

    # === Shared Kernel Allowances ===
    # shared/domain is allowed everywhere (commonComponents equivalent)
    # No deny rules needed -- it's the default allow

    # === Deprecated Packages ===
    deprecated:
      files:
        - $all
      deny:
        - pkg: "io/ioutil"
          desc: "Deprecated since Go 1.16; use io and os directly"
        - pkg: "github.com/pkg/errors"
          desc: "Deprecated; use stdlib errors + fmt.Errorf with %w"
```

### 6.4 Template for go-arch-lint Config (Alternative/Complementary)

```yaml
# Auto-generated from DDD.md bounded contexts
# Regenerate with: alty fitness generate

version: 3
workdir: internal

components:
  # --- bootstrap (Supporting) ---
  bootstrap-domain:       { in: bootstrap/domain/** }
  bootstrap-app:          { in: bootstrap/application/** }
  bootstrap-infra:        { in: bootstrap/infrastructure/** }

  # --- discovery (Core) ---
  discovery-domain:       { in: discovery/domain/** }
  discovery-app:          { in: discovery/application/** }
  discovery-infra:        { in: discovery/infrastructure/** }

  # --- Shared Kernel ---
  shared-domain:          { in: shared/domain/** }
  shared-app:             { in: shared/application/** }
  shared-infra:           { in: shared/infrastructure/** }

  # --- Composition Root ---
  composition:            { in: composition/** }

commonComponents:
  - shared-domain

deps:
  # Layer rules: domain -> nothing, app -> domain, infra -> domain + app
  bootstrap-domain:    { mayDependOn: [] }
  bootstrap-app:       { mayDependOn: [bootstrap-domain] }
  bootstrap-infra:     { mayDependOn: [bootstrap-domain, bootstrap-app, shared-app] }

  discovery-domain:    { mayDependOn: [] }
  discovery-app:       { mayDependOn: [discovery-domain] }
  discovery-infra:     { mayDependOn: [discovery-domain, discovery-app, shared-app] }

  shared-domain:       { mayDependOn: [] }
  shared-app:          { mayDependOn: [shared-domain] }
  shared-infra:        { mayDependOn: [shared-domain, shared-app] }

  composition:
    mayDependOn:
      - bootstrap-domain
      - bootstrap-app
      - bootstrap-infra
      - discovery-domain
      - discovery-app
      - discovery-infra
      - shared-domain
      - shared-app
      - shared-infra
```

### 6.5 Cross-Context Rules from Context Map

The DDD.md context map defines relationships between bounded contexts:
- **Upstream/Downstream** -- downstream may depend on upstream's published language
- **Shared Kernel** -- both contexts share types from `shared/domain`
- **Anti-Corruption Layer** -- downstream wraps upstream's types

For depguard, these translate to:
- **No relationship declared** = deny import between contexts
- **Upstream/Downstream** = downstream's infrastructure may import upstream's domain
- **Shared Kernel** = both may import `shared/domain`

This is the key insight: **the context map IS the dependency specification.**

---

## 7. Recommendation

### Primary: Use arch-go (MIT license)

**REVISED based on license analysis:**

depguard is GPL-3.0. While it doesn't link into the compiled binary, some enterprises have
blanket GPL restrictions even for tooling. Since alty generates configurations that users
adopt, recommending GPL tooling could create adoption barriers.

**arch-go (MIT, v2.1.2, Feb 2026) is the recommended tool because:**
1. **MIT license** — enterprise-friendly, no GPL contamination concerns
2. **Both rule types** — `shouldOnlyDependsOn` (positive) + `shouldNotDependsOn` (negative)
3. **Compliance thresholds** — 0-100% compliance allows gradual adoption for brownfield
4. **Single tool** — replaces depguard + go-arch-lint with one MIT-licensed tool

**What to implement:**
1. **Auto-generation** — alty generates `arch-go.yml` from `docs/DDD.md` bounded contexts
2. **Threshold modes** — 100% for greenfield, 80% for `alty init --existing`
3. **Cross-context isolation** — derive deny rules from context map relationships
4. **Remove depguard** — remove manual depguard rules from `.golangci.yml`

### Do NOT adopt:
- **depguard** — GPL-3.0 license risk
- **go-cleanarch** — abandoned since 2021
- **go-arch-lint alone** — MIT but only has positive rules (no deny), less flexible

### Migration path:
1. Keep golangci-lint for other linters (errcheck, staticcheck, etc.)
2. Add `arch-go` as a separate CI step
3. Remove depguard config from `.golangci.yml`
4. Document license rationale in generated configs

---

## 8. Answers to Research Questions

### Q1: Current state of depguard in golangci-lint v2?

depguard v2.2.1 is integrated into golangci-lint v2.11.3 as part of the `standard` linter
preset. Configuration uses named rules with file glob patterns and deny/allow lists.
Three list modes: `original` (default), `strict`, `lax`. Supports `$all`, `$test`, `$gostd`
variables. **Cannot be auto-generated** from any existing tool -- this would be a new alty
capability.

Source: [golangci-lint settings](https://golangci-lint.run/docs/linters/configuration/),
[depguard GitHub](https://github.com/OpenPeeDeeP/depguard)

### Q2: Does go-arch-lint exist and is it mature?

Yes. v1.14.0 released 2025-11-13. MIT licensed. 169 commits on master. Actively maintained.
Supports component-based architecture definition with `mayDependOn` positive dependency rules.
Has dependency graph generation. Config version 3 syntax. Requires Go 1.25+.

Source: [go-arch-lint GitHub](https://github.com/fe3dback/go-arch-lint)

### Q3: Other Go architecture testing tools in 2026?

- **arch-go v2.1.2** (2026-02-03) -- most recently updated. MIT. Both positive and negative
  rules. Compliance thresholds. Content and function rules beyond just dependencies.
  Source: [arch-go GitHub](https://github.com/arch-go/arch-go)
- **go-cleanarch v1.2.1** (2021) -- abandoned. Do not use.
  Source: [go-cleanarch GitHub](https://github.com/roblaszczak/go-cleanarch)

### Q4: How do DDD Go projects enforce layer boundaries?

Most use depguard deny rules in golangci-lint. Some add go-arch-lint for positive modeling.
Go's `internal/` package convention provides visibility enforcement but not dependency
direction enforcement. No standard approach exists for cross-context isolation.

### Q5: Best practices for generating depguard config from DDD.md?

No existing tooling does this. The generation template is documented in section 6.3 above.
Key insight: DDD.md context map relationships map directly to depguard allow/deny rules.
Universal layer rules use `**` wildcards; cross-context rules are context-specific deny lists
derived from the context map.

---

## 9. Follow-Up Work

1. **Implement arch-go config generation** in the fitness bounded context — generate
   `arch-go.yml` from parsed DDD.md bounded contexts and context map relationships.
2. **Remove depguard from .golangci.yml** — GPL-3.0 license is a concern; arch-go
   provides equivalent functionality under MIT.
3. **Add brownfield threshold support** — `alty init --existing` generates arch-go
   config with 80% compliance threshold for gradual adoption.
4. **Cross-context isolation rules** — derive `shouldNotDependsOn` rules from the
   context map (contexts without declared relationships are isolated).
5. **Update DDD.md** — reference arch-go instead of depguard in Story 4.
