---
name: design-ticket
description: Design a new ticket with architecture verification — architect designs, tech-lead reviews, then creates beads issue
kind: command
phase: groom
when_to_use: When designing a new ticket and verifying it against the architecture before creating the beads issue
tools_required: Agent, Bash, Read, Grep, Glob
bash_substitution_policy: none
license: Apache-2.0
---

# /design-ticket <title-or-description>

Design a single ticket with built-in architecture and DDD verification. Uses a two-agent chain: architect designs the ticket from codebase analysis, tech-lead reviews it against DDD.md and ARCHITECTURE.md before creation.

## Why This Exists

Creating tickets from rough bullet points produces vague descriptions that fail during implementation. The /groom command catches this AFTER creation — but by then the ticket exists with wrong assumptions baked in. This command prevents bad tickets from being created in the first place.

The key insight: tickets that touch application or infrastructure layers need to trace through existing code (callers, ports, adapters, composition wiring) BEFORE writing the description. A code-architect agent does this systematically.

## Usage

```
/design-ticket StorytellingFlow domain strategy for RAPID/THOROUGH modes
/design-ticket HuhStorytellingPrompter CLI adapter implementation
```

Provide a short description of what the ticket should accomplish. The agent will figure out the scope, dependencies, and design from the codebase.

## Process

### Phase 1 — Architect Designs the Ticket

Launch a `feature-dev:code-architect` agent with this prompt:

> Design a detailed implementation ticket for: {user's description}
>
> Read and analyze:
> - `docs/beads_templates/beads-ticket-template.md` — MANDATORY template structure (read FIRST)
> - `.notes/order.md` — current epic context, phases, settled decisions
> - `docs/DDD.md` — bounded contexts, ubiquitous language glossary
> - `docs/ARCHITECTURE.md` — layer rules, planned file layout, ADRs
> - Existing code in the relevant bounded context (ports, handlers, adapters, domain types)
> - Any spike research reports referenced in order.md
>
> Produce a ticket following `docs/beads_templates/beads-ticket-template.md` EXACTLY. Include ALL sections:
> - Goal / Problem
> - Background / Context (with verified file:line references)
> - DDD Alignment table
> - Design section (data models, method signatures, sequence flow)
> - SOLID Mapping
> - TDD Workflow (RED/GREEN/REFACTOR with specific test names)
> - Steps
> - Acceptance Criteria
> - Edge Cases
> - Quality Gates
> - Pre-Implementation Validation (verified against actual code — cite file:line)
> - Risks / Dependencies
>
> Key requirements:
> - Every file:line claim must be verified by reading the actual file
> - Every method signature must match what's actually in the code
> - Constructor chains must be traced through composition root
> - Name every dependency (no "TBD" or "to be determined")

### Phase 2 — Tech Lead Reviews the Design

Launch a `tech-lead` agent to review the architect's output:

> Review this ticket design for DDD/SOLID/CQRS-lite compliance:
>
> {paste architect output}
>
> Check against:
> 1. **DDD layer rules**: Does the ticket respect infrastructure → application → domain?
> 2. **Bounded context boundaries**: Does it stay within one context? Any cross-context leakage?
> 3. **Ubiquitous language**: Do all type/method names match docs/DDD.md glossary?
> 4. **Port/adapter pattern**: Are ports in application, adapters in infrastructure?
> 5. **SOLID violations**: Any god objects, leaky abstractions, or concrete dependencies?
> 6. **Scope creep**: Does it try to do too much? Should it be split?
> 7. **Missing dependencies**: Are there tickets that should exist but don't?
>
> Output: APPROVED, APPROVED WITH FIXES (list fixes), or NEEDS REDESIGN (explain why)

### Phase 3 — Apply Fixes and Create

If tech-lead says APPROVED WITH FIXES:
- Apply the fixes to the ticket description
- Show the user the changes for approval

If tech-lead says NEEDS REDESIGN:
- Show the user the tech-lead's concerns
- Ask whether to redesign or proceed anyway

Once approved:
```bash
bd create --title="[Phase N] <title>" --type=task --priority=1 --parent=<epic-id> --description="<ticket body>"
bd label add <new-id> storytelling
bd dep add <new-id> <depends-on-id>  # for each dependency
```

### Phase 4 — Report

```
================================================================
TICKET DESIGN REPORT
================================================================

TITLE:           <ticket title>
BEADS ID:        <created id>
PHASE:           <epic phase>

ARCHITECT:       code-architect agent
  Files read:    <N>
  Signatures verified: <N>
  Claims cited:  <N> (all with file:line)

TECH LEAD REVIEW:
  DDD compliance:      [PASS | FIX APPLIED]
  Layer rules:         [PASS | FIX APPLIED]
  Ubiquitous language: [PASS | FIX APPLIED]
  SOLID:               [PASS | FIX APPLIED]
  Scope:               [OK | SPLIT RECOMMENDED]

VERDICT: CREATED / NEEDS USER INPUT
================================================================
```

## Rules

1. **One ticket at a time.** Design quality drops when batching.
2. **Architect reads code, tech-lead reads docs.** Don't mix roles.
3. **Every claim is cited.** No "verified" without file:line.
4. **User approves before creation.** Never auto-create tickets.
5. **Settled decisions are binding.** Check `.notes/order.md` settled decisions section — don't re-debate.

## Product Dogfooding Note

This command implements what alto's ImplementabilityValidator (Ticket Freshness context) should do automatically for users' generated tickets. The architect→tech-lead chain maps to:
- ImplementabilityValidator (section completeness, unspecified dependencies)
- CodebasePortScanner (method signature verification)
- DesignTraceResult (structured findings)

When implementing the Ticket Pipeline context, reference this command as the UX model.
