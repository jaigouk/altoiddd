---
last_reviewed: 2026-02-22
owner: product
status: complete
spike: vibe-seed-k7m.9
---

# Killer Features Research Report

## Research Question

Which killer features should vibe-seed prioritize to differentiate from MetaGPT, Pythagora, Lovable/Bolt/v0, GPT-Engineer, and PRD generators (WriteMyPrd, Tara AI)?

## Decision

**"4+1" Comprehensive Launch** — Ship 4 features at P0, defer 2 to P1.

## 1. Feature Prioritization Matrix

| # | Feature | Impact | Effort | Moat | Phase | Persona Value |
|---|---------|--------|--------|------|-------|---------------|
| 1 | Architecture Fitness Functions | Very High | Medium | Deep | **P0** | Solo: High, Lead: Very High, Switcher: High |
| 2 | Domain Story to Ticket Pipeline | Very High | Medium | Deep | **P0** | Solo: Very High, Lead: High, Switcher: Medium |
| 3 | Complexity Budget | High | Low | Deep | **P0** | Solo: Medium, Lead: High, Switcher: Low |
| 4 | Tool-Native Context Translation | High | Medium | Medium | **P0** | Solo: Medium, Lead: Medium, Switcher: Very High |
| 5 | Rescue Mode / Structural Migration | Very High | High | Very Deep | **P1** | Solo: Low, Lead: Very High, Switcher: Medium |
| 6 | Living Knowledge Base (Drift Detection) | Medium | Ongoing | Deep | **P1** | Solo: Medium, Lead: Medium, Switcher: High |

### Phase Rationale

**P0 (Launch) — 4 features:**
- Features 1-4 are mutually reinforcing: fitness functions enforce what complexity budget classifies, ticket pipeline generates the work, tool translation distributes the rules.
- All four can be built incrementally during Phase 1 foundation work.
- Complexity Budget (feature 3) is low effort and multiplies the value of features 1 and 2.

**P1 (Growth) — 2 features:**
- Rescue Mode is the deepest moat but requires `vs init` to be solid first. High effort, deferred.
- Living Knowledge Base ships with basic structure in P0 (static `.vibe-seed/knowledge/`), smart drift detection added in P1.

## 2. Persona Validation

### Solo Developer (starting new project)
**Most valued:** Ticket Pipeline (immediate actionable work), Fitness Functions (architecture survives AI coding)
**Least valued:** Rescue Mode (no existing project), Complexity Budget (may not know DDD yet)
**Key insight:** Solo devs need the pipeline to tell them what to build next. That is the core loop.

### Team Lead (adopting on existing codebase)
**Most valued:** Rescue Mode (existing project), Fitness Functions (team-wide enforcement), Complexity Budget (prevent over-engineering)
**Least valued:** Ticket Pipeline (may have existing workflow)
**Key insight:** Team leads adopt after seeing solo dev success. Rescue Mode is the growth driver.

### AI Tool Switcher (using multiple tools)
**Most valued:** Tool Translation (the whole point), Living Knowledge Base (version tracking across tools)
**Least valued:** Complexity Budget (tool-agnostic concern)
**Key insight:** This persona validates that tool translation must be P0, not P1.

## 3. Competitive Landscape (Updated)

### Direct Threats

| Competitor | Status (Feb 2026) | Overlap with vibe-seed | Key Gap |
|---|---|---|---|
| **Amazon Kiro** | Active, proprietary, cloud-only | Spec-driven: requirements, design, tasks, tests | No DDD, no Python, AWS-locked |
| **GitHub Spec Kit** | v0.1.4, MIT, experimental | Constitution + Spec + Plan + Tasks | No DDD, no domain discovery |
| **BMAD Method** | v6, MIT | Multi-agent PRD-to-implementation pipeline | No DDD, Node.js only |
| **MetaGPT** | Stagnating (blocked on Python <3.12) | Multi-agent PRD + architecture + code | No boundary enforcement, no DDD |

### Validated White Space

No tool in the market combines:
1. DDD-guided domain discovery (bounded contexts, aggregates, ubiquitous language)
2. Executable architecture enforcement (fitness functions from domain model)
3. Multi-tool config generation (domain-aware, not generic boilerplate)
4. Python 3.12+ / uv ecosystem (entirely unserved by vibe coding tools)
5. Local-first operation (Kiro and Qlerify are cloud-only)

### Risk Window

The SDD (Spec-Driven Development) pattern is gaining traction in 2026. Kiro (AWS) and Spec-Kit (GitHub) follow requirements-to-tasks workflows but lack DDD. If AWS adds DDD support and Python templates, the gap narrows. The window to establish vibe-seed as the DDD-aware entrant is open now.

## 4. Feature-to-PRD Mapping

### Existing PRD Sections That Cover Features

| Feature | PRD Section | Coverage | Gap |
|---|---|---|---|
| Fitness Functions | Not in PRD | **Missing entirely** | Need new P0 capability |
| Ticket Pipeline | S4 Scenario 1 step 4-5, S5 Beads integration | Partial — mentions "beads epics and spikes" | Need explicit capability for auto-generation with dependency ordering |
| Complexity Budget | Not in PRD | **Missing entirely** | Need new P0 capability under DDD question framework |
| Tool Translation | S5 P1 Multi-tool support | Exists but wrong phase | Move from P1 to P0 |
| Rescue Mode | S4 Scenario 2, S5 Gap analysis | Good coverage | Stays P1 for implementation, spike (k7m.8) in Phase 1 |
| Living KB | S5 Knowledge base (RLM), S5 Doc maintenance | Good coverage | Add drift detection as P1 enhancement |

### New PRD Capabilities Needed

**Add to P0 (Must Have):**
1. **Architecture fitness function generation** — Generate import-linter TOML contracts and/or pytestarch test files from bounded context map. Enforce layer boundaries, dependency direction, and aggregate isolation as executable tests.
2. **Domain story to ticket pipeline** — Auto-generate dependency-ordered beads epics from DDD artifacts. Tickets pre-filled with TDD phases, SOLID mapping, and acceptance criteria derived from domain invariants. Human-in-the-loop: preview and confirm before creation.
3. **Complexity budget classification** — During DDD discovery, classify each subdomain as Core/Supporting/Generic. Budget determines: ticket detail level, fitness function strictness, and recommended implementation approach.

**Move from P1 to P0:**
4. **Multi-tool support** — Currently P1 in PRD section 5. Should be P0 given user preference for comprehensive launch and competitive pressure from Kiro/Spec-Kit.

**Add to P1 (Should Have):**
5. **Knowledge base drift detection** — Detect when tool conventions change between versions, flag stale architecture docs vs actual code structure.

## 5. Feature Dependency Map

### Feature to Spike Dependencies

| Feature | Required Spikes | Status |
|---|---|---|
| Fitness Functions | k7m.10 (fitness function design), k7m.2 (DDD questions for context map input) | k7m.10 created, k7m.2 ready |
| Ticket Pipeline | k7m.11 (pipeline design), k7m.2 (DDD questions for artifact input) | k7m.11 created, k7m.2 ready |
| Complexity Budget | k7m.2 (DDD questions — subdomain classification is part of discovery) | k7m.2 ready |
| Tool Translation | k7m.3 (multi-tool config formats), k7m.1 (knowledge base structure) | Both ready |
| Rescue Mode | k7m.8 (gap analysis design) | Ready, bumped to P1 |
| Living KB | k7m.1 (knowledge base structure) | Ready |

### Feature Interactions

```
Complexity Budget ──────────┐
                            ▼
DDD Questions (k7m.2) ──→ Ticket Pipeline ──→ Beads epics
        │                                        │
        ▼                                        ▼
Bounded Context Map ──→ Fitness Functions ──→ pytest / CI gate
        │
        ▼
Tool Translation ──→ .claude/, .cursor/, etc. (domain-aware)
        │
        ▼
Knowledge Base ──→ version tracking, drift detection (P1)
```

Key interactions:
- Fitness Functions + Complexity Budget = budget-aware boundary tests (stricter for Core, relaxed for Generic)
- DDD Questions + Ticket Pipeline = end-to-end automation from idea to tickets
- Tool Translation + Knowledge Base = version-aware config generation
- Complexity Budget + Ticket Pipeline = right-sized tickets per subdomain type

## 6. Recommended PRD Changes for k7m.6

### Section 5 (Capabilities) Changes

**Add 3 new P0 capabilities (after "Quality gates"):**

```markdown
- [ ] **Architecture fitness function generation** — Generate executable architecture
      tests (import-linter contracts, pytestarch rules) from bounded context map.
      Enforce layer boundaries, dependency direction, aggregate isolation. Tests run
      as part of quality gates.
- [ ] **Domain story to ticket pipeline** — Auto-generate dependency-ordered beads
      epics from DDD artifacts. Tickets include TDD phases, SOLID mapping, acceptance
      criteria from domain invariants. Preview before creation (human-in-the-loop).
- [ ] **Complexity budget** — Classify subdomains as Core/Supporting/Generic during
      DDD discovery. Core gets full DDD treatment, Supporting gets simple services,
      Generic gets CRUD/library recommendations. Budget enforced in tickets and
      fitness functions.
```

**Move "Multi-tool support" from P1 to P0.**

**Add P1 capability:**

```markdown
- [ ] **Knowledge base drift detection** — Detect tool convention changes between
      versions, flag stale architecture docs vs actual code structure, suggest updates.
```

### Section 8 (Success Metrics) Changes

**Add:**

| Metric | Target | Measurement Method |
|--------|--------|-------------------|
| Architecture test generation | 100% of bounded contexts | Every context has at least one fitness function test |
| Ticket pipeline accuracy | Zero manual reordering | Generated dependency order matches actual build order |

### Section 9 (Risks) Changes

**Add:**

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| import-linter API too limited for generation | Medium | Medium | Spike k7m.10; fallback to pytestarch .py generation |
| Kiro adds DDD support | Low | High | Ship first, establish community, local-first advantage |

### Section 10 (Timeline) Changes

**Update phases to reflect 4+1 strategy:**

| Phase | Deliverable |
|-------|-------------|
| Phase 1: Foundation | PRD + DDD + Architecture for vibe-seed itself |
| Phase 2: Core CLI | `vs init` + guided DDD questions + artifact generation |
| Phase 3: Fitness & Tickets | Architecture fitness function generation + ticket pipeline + complexity budget |
| Phase 4: Multi-Tool | Config generation for Claude Code, Cursor, Antigravity, OpenCode |
| Phase 5: MCP Server | MCP tools exposing same core as CLI |
| Phase 6: Rescue Mode | `vs init --existing` with structural migration |
| Phase 7: Doc Maintenance | `vs doc-health`, drift detection, maintenance registry |

### Section 9 (Open Questions) Changes

**Add:**

```markdown
- [ ] **Spike: Fitness function generation** — How to map bounded context map to
      import-linter TOML / pytestarch tests? (k7m.10)
- [ ] **Spike: Ticket pipeline** — How to auto-generate ordered beads tickets from
      DDD artifacts? (k7m.11)
```

## 7. Follow-up Tickets

### Created During This Spike

| ID | Title | Priority | Rationale |
|---|---|---|---|
| k7m.10 | Spike: Architecture fitness function generation design | P1 | P0 feature needs technical design before implementation |
| k7m.11 | Spike: Domain story to ticket pipeline design | P1 | P0 feature needs generation algorithm design |

### Dependency Fixes Applied

| Change | Rationale |
|---|---|
| k7m.7 now depends on k7m.3 | Tool translation is P0; architecture doc needs multi-tool input |
| k7m.5 now depends on k7m.2 | DDD question framework informs how we write our own DDD artifacts |
| k7m.8 bumped to P1 | Rescue Mode spike research should happen in Phase 1 to inform architecture |
| k7m.7 now depends on k7m.10, k7m.11 | Architecture doc needs fitness function and pipeline designs |

### Suggested for Phase 2 Epic (not created yet)

These should be created when the Phase 2 epic is established:
- Task: Implement fitness function generation (import-linter + pytestarch)
- Task: Implement ticket pipeline (DDD artifacts to beads)
- Task: Implement complexity budget classification in DDD question flow
- Task: Implement multi-tool config generation
- Task: Implement knowledge base drift detection (P1)
- Task: Implement rescue mode / structural migration (P1)

## 8. Technical Foundation (from Research)

### Architecture Fitness Functions — Build On

| Tool | Version | License | Role |
|---|---|---|---|
| import-linter | v2.10 (Feb 2026) | BSD-2 | Config-based layer/forbidden/independence contracts |
| pytestarch | v4.0.1 (Aug 2025) | Apache-2.0 | Code-based fluent API, pytest integration |
| grimp | v3.14 (Dec 2025) | BSD-2 | Programmatic import graph API (underlies import-linter) |
| deply | v0.8.0 | BSD-3 | YAML-driven + Mermaid diagram generation (secondary) |

**Key finding:** Java has Context Mapper + ArchUnit + jMolecules for DDD-to-test pipelines. Python has nothing equivalent. vibe-seed would be building novel capability in the Python ecosystem.

**Risk:** import-linter's Python API is read-only. Generating contracts means emitting TOML programmatically. A Pydantic model layer should wrap this to avoid string manipulation bugs.

### Competitive Tools — Key Capabilities

| Tool | Spec-Driven | DDD | Python | Local | Fitness Tests | Multi-Tool |
|---|---|---|---|---|---|---|
| vibe-seed | Yes | Yes | Yes | Yes | Yes | Yes |
| Amazon Kiro | Yes | No | No | No | No | No |
| GitHub Spec Kit | Yes | No | No | Yes | No | Partial |
| BMAD Method | Yes | No | No | Yes | No | No |
| MetaGPT | Partial | No | Broken (< 3.12) | Yes | No | No |
| Qlerify | No | Yes | No | No | No | No |

vibe-seed is the only tool that combines all six capabilities.

## Sources

- MetaGPT GitHub: architecture, QA agent, Python version constraint
- Pythagora/GPT Pilot: archived, redirects to commercial SaaS
- Amazon Kiro: spec-driven IDE, cloud-only, AWS-locked
- GitHub Spec Kit v0.1.4: constitution + spec scaffolding
- BMAD Method v6: multi-agent PRD pipeline, Node.js only
- Qlerify: DDD output from prompts, cloud-only SaaS
- import-linter v2.10 docs: contract types, TOML config, grimp engine
- pytestarch v4.0.1 docs: fluent API, pytest integration
- Ox Security, METR, CodeRabbit studies: AI code quality concerns in 2026

Supporting research reports:
- `docs/research/20260222_metagpt_pythagora_competitive_analysis.md`
- `docs/research/20260222_vibe_coding_landscape.md`
- `docs/research/20260222_python_architecture_testing_fitness_functions.md`
