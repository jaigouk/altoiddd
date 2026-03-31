# Competitive Landscape: Spec-Driven Development / AI Project Bootstrapping Tools

**Date:** 2026-03-31
**Purpose:** Meetup talk comparison table update

## Comparison Table

| Tool | Spec-Driven? | DDD? | Fitness Tests? | Multi-Tool? | Local-First? |
|------|:---:|:---:|:---:|:---:|:---:|
| **alto** | Yes | Yes (Domain Storytelling, bounded contexts, ubiquitous language) | Yes (arch-go, depguard layer enforcement) | Yes (Claude Code, Cursor, Roo Code, OpenCode) | Yes |
| **Amazon Kiro** | Yes (Requirements -> Design -> Tasks) | No | No | No (Kiro IDE only, VS Code fork) | No (AWS cloud) |
| **GitHub Spec Kit** | Yes (Constitution -> Specify -> Plan -> Tasks) | No | No (checklists only) | Yes (Copilot, Claude Code, Gemini CLI) | Yes |
| **BMAD Method** | Yes (multi-agent SDLC personas) | No (upfront interview, not continuous discovery) | No | Yes (Claude Code, Cursor, Cline) | Yes |
| **MetaGPT** | Yes (PRD -> Architecture -> Tasks) | No | No | No (own runtime) | Partial (local + API keys) |
| **gstack** | Partial (office-hours forcing questions, eng-review) | No | No | Yes (Claude Code primary, Codex, Cursor, Gemini CLI) | Yes |
| **Augment Intent** | Yes (living specs, bidirectional sync) | No | No (Verifier agent, not arch tests) | Yes (Claude Code, Codex, OpenCode via MCP) | No (Augment cloud) |
| **Tessl Framework** | Yes (spec-as-source, 1:1 spec-to-code mapping) | No | No | Yes (agent-agnostic via tiles) | Partial (closed beta, registry is cloud) |
| **OpenSpec** | Yes (proposal -> apply -> archive state machine) | No | No | Yes (Claude Code, Cursor, Copilot, Cline, Windsurf) | Yes |

## Key Findings Per Tool

### gstack (Garry Tan)
- **What it is:** 23 Claude Code slash commands organized as a virtual engineering team (CEO, Designer, Eng Manager, QA, etc.) following Think -> Plan -> Build -> Review -> Test -> Ship -> Reflect sprint phases.
- **Spec-driven:** Partial. `/office-hours` asks 6 "forcing questions" before coding. `/plan-eng-review` locks architecture, data flow, and edge cases. But no formal spec document is generated or versioned.
- **DDD:** No. No domain discovery, no bounded contexts, no ubiquitous language.
- **Fitness tests:** No. Has `/qa` and `/benchmark` but these are runtime test execution, not architecture fitness functions.
- **Multi-tool:** Yes. Claude Code primary, but `/codex` sends code to OpenAI Codex for second opinion. Also supports Cursor, Gemini CLI, Factory Droid.
- **Local-first:** Yes. Pure `.claude/commands/` directory, no background processes, no cloud.
- **Stars:** 16k+ GitHub stars, 1.8k forks (as of March 2026).
- **Source:** [github.com/garrytan/gstack](https://github.com/garrytan/gstack)

### Augment Intent (NEW -- not in your list)
- **What it is:** Augment Code's post-IDE workspace. Coordinator Agent proposes spec, fans work to Implementor agents in isolated git worktrees, Verifier checks results.
- **Key differentiator:** Living specs that update bidirectionally (requirements -> agents AND agents -> spec).
- **DDD:** No.
- **Source:** [augmentcode.com/product/intent](https://www.augmentcode.com/product/intent)

### Tessl Framework (NEW -- not in your list)
- **What it is:** Spec-as-source platform. Reverse-engineers specs from existing code, maintains 1:1 spec-to-code mapping. Generated code marked `// GENERATED FROM SPEC - DO NOT EDIT`.
- **Status:** Closed beta. Spec Registry in open beta.
- **DDD:** No.
- **Source:** [tessl.io](https://tessl.io/), [Martin Fowler analysis](https://martinfowler.com/articles/exploring-gen-ai/sdd-3-tools.html)

### OpenSpec (NEW -- not in your list)
- **What it is:** Lightweight SDD framework by Fission AI. Strict 3-phase state machine (proposal, apply, archive). Specifically targets brownfield/existing codebase evolution with delta markers (ADDED/MODIFIED/REMOVED).
- **DDD:** No.
- **Source:** [github.com/Fission-AI/OpenSpec](https://github.com/Fission-AI/OpenSpec)

## alto's Unique Position

**No other tool in this landscape does DDD or generates architecture fitness tests.**

The Martin Fowler article on SDD tools explicitly notes that none of the tools he reviewed address DDD or fitness tests. The innoq article "Spec-Driven Development is Domain-Driven Design's Impatient Cousin" argues that SDD tools assume upfront discovery is sufficient, while DDD requires continuous discovery during implementation -- exactly the gap alto fills with Domain Storytelling and iterative bounded context refinement.

Key differentiators for alto:
1. **Only tool with domain discovery** (Domain Storytelling, boundary detection, ubiquitous language extraction)
2. **Only tool with executable architecture tests** (arch-go for layer enforcement, depguard for dependency rules)
3. **Continuous discovery** vs. upfront-only spec generation
4. **Subdomain classification** (Core/Supporting/Generic with different treatment per type)

## Sources

- [gstack README](https://github.com/garrytan/gstack)
- [gstack Medium analysis](https://medium.com/@luongnv89/gstack-is-not-a-dev-tool-its-garry-tan-s-brain-on-ai-b813e09b32c7)
- [Amazon Kiro](https://kiro.dev/)
- [GitHub Spec Kit blog post](https://github.blog/ai-and-ml/generative-ai/spec-driven-development-with-ai-get-started-with-a-new-open-source-toolkit/)
- [BMAD Method](https://github.com/bmad-code-org/BMAD-METHOD)
- [MetaGPT](https://github.com/geekan/MetaGPT)
- [Augment Intent](https://www.augmentcode.com/product/intent)
- [Tessl](https://tessl.io/)
- [OpenSpec](https://github.com/Fission-AI/OpenSpec)
- [Martin Fowler: SDD Tools Comparison](https://martinfowler.com/articles/exploring-gen-ai/sdd-3-tools.html)
- [innoq: SDD is DDD's Impatient Cousin](https://www.innoq.com/en/blog/2026/03/sdd-ddd-why-bmad-wont-save-you/)
- [Augment Code: 6 Best SDD Tools 2026](https://www.augmentcode.com/tools/best-spec-driven-development-tools)
