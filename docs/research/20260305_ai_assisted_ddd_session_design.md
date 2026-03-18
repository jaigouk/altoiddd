# Research: AI-Assisted Iterative DDD Session Design

**Date:** 2026-03-05
**Spike Ticket:** alto-20c.1
**Status:** Final

## Summary

alto's current guided discovery is a single-pass 10-question flow that produces shallow DDD artifacts in ~15 minutes. This spike designs an iterative, AI-assisted discovery session that produces deep artifacts (Bounded Context Canvas, rich ubiquitous language, well-defined aggregates) in under 2 hours — without requiring users to attend month-long workshop series.

**Key insight:** AI agents can compress DDD workshop time by playing three roles simultaneously — Facilitator (asks questions), Challenger (finds gaps), and Domain Researcher (gathers context) — creating rapid iteration loops that would take multiple human workshop sessions. A pomodoro-style cadence (25 min focus + 5 min review break) creates natural reflection points and prevents user fatigue.

## Research Questions Answered

### 1. Session Structure — Multi-Round Protocol

Current flow: `README → 10 questions → artifacts (single pass, ~15 min)`

Proposed flow: **Three rounds with increasing depth, complexity-budget-aware.**

```
Round 1: Express Discovery (existing 10-question flow)
  Input:  README (1-2 paragraphs)
  Output: DDD.md v1 (shallow: contexts, language, stories, classification)
  Time:   ~15 min
  AI role: Facilitator only

Round 2: AI Challenge & Deepen (NEW) — Pomodoro format
  Input:  DDD.md v1 + AI domain research
  Output: DDD.md v2 (deeper: canvases, invariants, contradictions resolved)
  AI roles: Challenger + Domain Researcher
  Focus:  Core subdomains only (complexity budget gates depth)

  ┌─────────────────────────────────────────────────────────┐
  │ Pomodoro 2A — Challenge: Gaps & Contradictions (25 min) │
  └─────────────────────────────────────────────────────────┘

  0:00-0:03  ORIENT (3 min)
    AI: Presents domain research summary (competitive landscape,
        common patterns in this problem domain, known pitfalls).
    AI: Lists Core subdomains that will be challenged.
    User: Reads, asks clarifying questions if needed.

  0:03-0:08  LANGUAGE CHALLENGE (5 min)
    AI: Scans ubiquitous language for ambiguity and overlap.
        "You use 'account' in both Billing and Identity —
         do they mean the same thing?"
        "Is 'booking' the same as 'reservation'? In hotels
         these mean different things."
    User: Clarifies each term. AI updates glossary in real-time.
    Target: 2-4 disambiguation challenges.

  0:08-0:16  INVARIANT CHALLENGE (8 min)
    AI: For each Core aggregate, probes for missing rules.
        "Your Booking aggregate has no invariants. What prevents
         double-booking the same sitter?"
        "You said orders can be cancelled — but what about
         after shipping?"
    User: States the rule or says "no rule needed."
    AI: Adds confirmed invariants to aggregate section.
    Target: 1-2 invariant challenges per Core aggregate.

  0:16-0:22  FAILURE MODE CHALLENGE (6 min)
    AI: For each Core workflow, asks "what goes wrong?"
        "What if payment fails after booking is confirmed?"
        "What if a sitter cancels 1 hour before the sitting?"
    User: Describes the failure handling or says "not yet decided."
    AI: Adds failure stories to domain stories section.
    Target: 1 failure mode per Core workflow.

  0:22-0:25  BOUNDARY CHALLENGE (3 min)
    AI: Questions context boundaries based on research.
        "Should 'sitter rating' live in Booking or Trust?
         Rating affects booking decisions but is owned by Trust."
        "Your Search context shares 5 entities with Booking —
         are they really separate contexts?"
    User: Confirms or adjusts boundaries.
    Target: 1-2 boundary challenges.

  ┌──────────────────────────────────────────────────────────┐
  │ Break 1 — Review & Reflect (5 min)                       │
  └──────────────────────────────────────────────────────────┘

  AI: Shows diff v1 → v2-draft:
      "+ 4 invariants added"
      "~ 2 terms disambiguated"
      "+ 2 failure stories"
      "~ 1 boundary adjusted"
  AI: Flags open items: "You said 'not yet decided' for payment
      failure handling — want to tackle this in 2B?"
  User: Reviews diff. Marks items to revisit. Stretches.
  AI: If diff < 3 changes total, suggests: "Model looks solid.
      Skip 2B and move to scenario testing?"
  User: Decides continue or skip.

  ┌─────────────────────────────────────────────────────────────┐
  │ Pomodoro 2B — Deepen: Aggregates & Canvas Design (25 min)   │
  └─────────────────────────────────────────────────────────────┘

  0:00-0:03  REVISIT (3 min)
    AI: Pulls up open items from Break 1 ("not yet decided" items).
    User: Resolves or explicitly defers each one.

  0:03-0:12  AGGREGATE DEEP DIVE (9 min)
    AI: For each Core aggregate, walks through design canvas:
        - "What commands does Booking handle?"
        - "What events does it emit after each command?"
        - "What's the consistency boundary? Can BookingRequest
           and Payment be in the same aggregate?"
        - "What's the expected throughput? Affects aggregate size."
    User: Answers. AI fills in the Aggregate Design Canvas section.
    Target: Full aggregate design for each Core aggregate root.

  0:12-0:19  COMMUNICATION PATTERNS (7 min)
    AI: Maps how contexts talk to each other:
        - "When Booking emits BookingConfirmed, who listens?"
        - "Does Payments need to query Booking, or does Booking
           push the data? (sync vs async)"
        - "Is the integration Shared Kernel, Customer-Supplier,
           or Conformist?"
    User: Decides integration style per relationship.
    AI: Fills inbound/outbound sections of each canvas.
    Target: Communication patterns for all Core context pairs.

  0:19-0:25  BOUNDED CONTEXT CANVAS ASSEMBLY (6 min)
    AI: For each Core context, assembles the full canvas:
        Purpose, Strategic Classification, Domain Roles,
        Inbound/Outbound Communication, Ubiquitous Language,
        Business Decisions, Assumptions, Open Questions.
    AI: Presents each canvas for quick validation.
    User: Confirms or corrects. Flags open questions.
    Target: Complete canvas draft for every Core context.

  ┌──────────────────────────────────────────────────────────┐
  │ Break 2 — Review & Decide (5 min)                        │
  └──────────────────────────────────────────────────────────┘

  AI: Shows full diff v1 → v2:
      Convergence metric: "+N invariants, +N terms, +N stories"
  AI: Shows Bounded Context Canvases for review.
  AI: "Round 2 complete. Ready for scenario stress-testing?"
  User: Reviews canvases. Decides continue to Round 3 or stop.

  ┌────────────────────────────────────────────────────────────────┐
  │ Pomodoro 3A — Simulate: Core Scenarios & Edge Cases (25 min)   │
  └────────────────────────────────────────────────────────────────┘

  0:00-0:05  SCENARIO GENERATION (5 min)
    AI: Generates a scenario deck from the domain model:
        - 1 happy path per Core context (derived from domain stories)
        - 1 failure mode per Core context (from Round 2 challenges)
        - 2-3 edge cases (from invariants — "what if this rule is
          almost violated?")
    AI: Presents the scenario list. User can add/remove scenarios.
    Target: 4-8 scenarios queued.

  0:05-0:10  HAPPY PATH WALKTHROUGH (5 min)
    AI: Narrates each happy path as a concrete story:
        "Alice creates a Pet Profile for her cat Miso. She searches
         for sitters available Dec 15-18 in Berlin. She finds Bob,
         who accepts cats. She sends a BookingRequest..."
    AI: Traces through contexts: BookingRequest → Booking context
        → BookingConfirmed event → Payments context → Escrow...
    AI: Flags any step where the model is silent ("your model
        doesn't define what happens between BookingConfirmed
        and EscrowCreated — is there a policy?").
    User: Confirms ✓ or fixes ✗ each step.

  0:10-0:18  FAILURE & EDGE CASE WALKTHROUGH (8 min)
    AI: Narrates failure scenarios:
        "Bob accepts Alice's booking but then cancels 2 hours
         before. What happens to the payment? Does Alice get
         auto-rebooked? What's Bob's penalto?"
    AI: Narrates edge cases:
        "Alice tries to book Bob for Dec 15-18, but Bob already
         has a booking for Dec 16. Your invariant says 'no
         overlapping bookings' — but does Dec 15-18 overlap
         with a Dec 16-only booking?"
    User: For each scenario:
        ✓ "Model handles this" (AI marks validated)
        ✗ "Model breaks" → user describes fix → AI applies it
        ⊘ "Out of scope" → AI records as future consideration

  0:18-0:23  POLICY GAP SCAN (5 min)
    AI: Checks for missing "whenever X, then Y" rules:
        "BookingConfirmed triggers EscrowPayment — but what
         triggers the sitting to actually start? Is there a
         SittingStarted event?"
        "You have a Dispute process but no policy for what
         happens when a dispute is resolved in favor of the
         owner — is the sitter penalized?"
    User: Confirms new policies or defers.

  0:23-0:25  TALLY (2 min)
    AI: Quick score:
        "Scenarios tested: 6/8"
        "Model gaps found & fixed: 3"
        "Deferred to future: 2"
        "Open questions added: 1"

  ┌──────────────────────────────────────────────────────────┐
  │ Break 3 — Converge or Continue? (5 min)                  │
  └──────────────────────────────────────────────────────────┘

  AI: Shows convergence metric:
      "Pomodoro 2A: +5 invariants, +3 terms, +2 stories"
      "Pomodoro 2B: +1 invariant, +1 term, +0 stories"
      "Pomodoro 3A: +1 invariant, +0 terms, +1 story"
  AI: "Model is stabilizing. Remaining gaps: [list]"
  AI: If zero gaps → "Model has converged. Recommend stopping."
      If 1-2 gaps → "Minor gaps remain. Optional Pomodoro 3B
      would cover cross-context and scale scenarios."
      If 3+ gaps → "Recommend continuing with Pomodoro 3B."
  User: Decides stop (→ DDD.md v3 final) or continue.

  ┌──────────────────────────────────────────────────────────────────────┐
  │ Pomodoro 3B (optional) — Advanced: Scale & Cross-Context (25 min)    │
  └──────────────────────────────────────────────────────────────────────┘

  0:00-0:03  CROSS-CONTEXT SETUP (3 min)
    AI: Identifies context interaction chains from the model:
        "Booking → Payments → Notifications is a 3-context chain.
         Let's trace a full lifecycle through all three."

  0:03-0:12  CROSS-CONTEXT SCENARIOS (9 min)
    AI: Traces end-to-end flows across multiple contexts:
        "Booking emits BookingConfirmed → Payments creates Escrow
         → Payments emits EscrowCreated → Notifications sends
         confirmation email. What if Notifications fails? Does
         Payments roll back? Does Booking know?"
    AI: Tests eventual consistency assumptions:
        "If Payments takes 30 seconds to create the escrow,
         can the user cancel the booking in the meantime?"
    User: Confirms or fixes each cross-context interaction.
    Target: 1 end-to-end trace per major workflow.

  0:12-0:19  SCALE SCENARIOS (7 min)
    AI: Tests the model under volume:
        "A popular sitter gets 50 booking requests per day.
         Does the Booking aggregate handle concurrent requests?
         Is there a queue or first-come-first-served?"
        "The review system has 10,000 reviews for top sitters.
         Does the SitterProfile aggregate load all reviews?"
    User: Identifies scalability concerns or confirms design.
    AI: Records scale notes in Open Questions or Assumptions.

  0:19-0:23  POLICY CONFLICT DETECTION (4 min)
    AI: Checks for contradicting policies:
        "Policy 1: 'Cancel within 48h → penalto for sitter.'
         Policy 2: 'Sitter can block dates at any time.'
         Conflict: What if sitter blocks a date that has an
         existing booking? Is that a cancellation?"
    User: Resolves or flags for later.

  0:23-0:25  FINAL TALLY (2 min)
    AI: Final convergence metric and artifact summary.
    AI: Generates DDD.md v3 (final).

Total deep session: ~75-120 min (2-4 pomodoros + breaks)
  Minimum: Round 1 (15 min) + 2A + break (30 min) + 3A (25 min) = ~70 min
  Maximum: All pomodoros including optional 3B = ~120 min
```

**Key design decisions:**

1. **Each pomodoro has internal structure** — Not 25 minutes of freeform conversation. Each pomodoro is divided into 3-5 timed activities with clear targets. The AI acts as timekeeper and facilitator, moving to the next activity when the current one completes or time runs out.

2. **Pomodoro cadence** — 25 min focused work + 5 min break mirrors proven productivity research. The break is not idle — AI uses it to summarize diffs and the user reviews what changed. This creates natural reflection points that improve decision quality.

3. **Breaks as review gates** — Each 5-min break shows the user what the AI changed, plus the convergence metric. This prevents "drift" where the user loses track of accumulated changes. The break diff is a mini-playback.

4. **Progressive focus** — 2A (surface gaps) → 2B (deep design) → 3A (validate with scenarios) → 3B (stress with scale). Each pomodoro builds on the previous. If the user stops early, they still have a valid artifact.

5. **Optional final pomodoro (3B)** — Not everyone needs 4 pomodoros. After Break 3, user stops if the model is stable. This respects the stopping heuristic (convergence) and user fatigue.

6. **Complexity budget gates depth** — Only **Core** subdomains get the full pomodoro treatment. Supporting gets a light pass (one challenge question per context during 2A). Generic gets nothing extra — it's off-the-shelf by definition.

7. **Concrete scenario narration** — AI doesn't ask abstract questions in Round 3. It narrates concrete stories with real names and data ("Alice books Bob's cat sitting for Dec 15-18"). This grounds the model in reality and surfaces gaps that abstract questions miss.

### 2. AI Agent Roles

#### Role 1: Facilitator (Round 1 — exists today)
- Asks the 10-question framework in dual register
- Extracts artifacts from answers
- Triggers playback checkpoints
- **No changes needed** — this is the current `DiscoverySession` aggregate

#### Role 2: Challenger (Round 2 — NEW)
- Reads DDD.md v1 and finds gaps:
  - **Missing invariants:** "You said Order can be cancelled, but what prevents cancellation after shipping?"
  - **Ambiguous language:** "You use 'account' in both Billing and Identity contexts — do they mean the same thing?"
  - **Missing failure modes:** "What happens if payment fails after booking is confirmed?"
  - **Under-specified aggregates:** "Your Booking aggregate has no invariants. What rules prevent invalid bookings?"
  - **Context boundary challenges:** "Should 'user profile' live in Identity or in the context that uses it most?"
- Presents challenges one at a time, user responds
- Each response refines the relevant artifact section
- **Implementation:** Prompt-based — feed DDD.md v1 to LLM with challenger system prompt

#### Role 3: Domain Researcher (Round 2 — NEW)
- Before Round 2 starts, AI does background research:
  - Web search for the problem domain (e.g., "pet sitting marketplace business model")
  - Competitive analysis (what do existing solutions do?)
  - Common domain patterns (e.g., "marketplace trust patterns", "escrow payment flows")
- Research informs the Challenger's questions — makes them domain-specific, not generic
- **Implementation:** Web search + summarization, stored as research context

#### Role 4: Customer Simulator (Round 3 — NEW)
- Takes the domain model and generates concrete scenarios:
  - Happy path: "Alice books a pet sitter for her cat for 3 days in December..."
  - Edge case: "What if the sitter accepts two overlapping bookings?"
  - Failure mode: "Payment processor is down when sitting completes..."
  - Scale scenario: "What if a sitter has 50 pending reviews to write?"
- For each scenario, traces through bounded contexts, commands, events, policies
- Flags where the model breaks or is silent
- User confirms fixes or says "out of scope"
- **Implementation:** Prompt-based — feed DDD.md v2 + scenario generation prompt

### 3. Artifact Iteration Protocol

```
DDD.md v1 (Round 1: Express Discovery, ~15 min)
    │
    ├── AI domain research runs (background, during Round 1)
    │
    ▼
Pomodoro 2A — Challenger: gaps, invariants, language (25 min)
    │
    ├── User responds to each challenge
    ├── Each response updates the relevant section:
    │   - New invariant → Aggregate section
    │   - Disambiguation → Ubiquitous Language
    │   - New failure mode → Domain Story section
    │   - Boundary change → Context Map
    │
    ▼
Break (5 min) — AI shows diff v1→v2-draft, user reviews
    │
    ▼
Pomodoro 2B — Deeper: aggregates, communication, canvases (25 min)
    │
    ▼
DDD.md v2 (Round 2 output, with Bounded Context Canvases)
    │
    ▼
Pomodoro 3A — Simulator: happy path, edge cases, failures (25 min)
    │
    ├── User confirms each scenario result:
    │   ✓ "Model handles this correctly"
    │   ✗ "Model breaks here" → fix applied
    │   ⊘ "Out of scope for now"
    │
    ▼
Break (5 min) — AI shows remaining gaps, user decides continue/stop
    │
    ├── STOP → DDD.md v3 (final)
    │
    ▼
Pomodoro 3B (optional) — Scale, cross-context, policy conflicts (25 min)
    │
    ▼
DDD.md v3 (final output, validated)
```

**Diff presentation:** Between rounds, show the user what changed:
```
DDD.md v1 → v2 changes:
  + Added invariant: "Sitter cannot have overlapping bookings" (Booking context)
  + Disambiguated: "account" → "UserAccount" (Identity) vs "BillingAccount" (Payments)
  + New failure story: "Payment fails after booking confirmed"
  ~ Moved: "SitterProfile" from Booking to Identity context
```

### 4. Bounded Context Canvas — Markdown Format

Based on [ddd-crew Bounded Context Canvas v5](https://github.com/ddd-crew/bounded-context-canvas):

```markdown
# Bounded Context Canvas: [Context Name]

## Purpose
[2-3 sentences: why this context exists, in business language]

## Strategic Classification
| Dimension | Value |
|-----------|-------|
| **Domain** | Core / Supporting / Generic |
| **Business Model** | Revenue / Engagement / Compliance |
| **Evolution** | Genesis / Custom / Product / Commodity |

## Domain Roles
- [ ] Execution (enforces workflows, processes commands)
- [ ] Analysis (processes data, produces insights)
- [ ] Gateway (translates between internal and external)
- [ ] Specification (defines rules consumed by others)
- [ ] Draft (exploratory, not yet stable)

## Inbound Communication
| Message | Type | Sender |
|---------|------|--------|
| [e.g., CreateBooking] | Command | [e.g., API Gateway] |
| [e.g., GetAvailability] | Query | [e.g., Search Context] |

## Outbound Communication
| Message | Type | Receiver |
|---------|------|----------|
| [e.g., BookingConfirmed] | Event | [e.g., Payments Context] |
| [e.g., NotifySitter] | Command | [e.g., Notifications] |

## Ubiquitous Language
| Term | Definition |
|------|-----------|
| [e.g., Booking] | [A confirmed reservation...] |

## Business Decisions & Rules
- [e.g., A sitter cannot have overlapping bookings]
- [e.g., Cancellation within 48h incurs a penalto]

## Assumptions
- [e.g., All sitters are pre-verified before accepting bookings]

## Open Questions
- [e.g., Should emergency rebooking be automatic or manual?]
```

**Design decision:** Keep it as a markdown section within DDD.md, one canvas per bounded context. Not a separate file — DDD.md is already the canonical document.

### 5. Express vs Deep Mode UX

```
$ alto guide

How thorough should the discovery session be?

  1. Express (15 min) — Quick 10-question flow. Good for prototypes,
     solo experiments, and well-understood domains.

  2. Deep (60-90 min) — Full iterative session with AI challenger,
     domain research, and scenario testing. Recommended for real
     projects and unfamiliar domains.

  [1/2]:
```

**Implementation:** A single flag on `DiscoverySession`:
- Express mode: existing flow unchanged (backward compatible)
- Deep mode: after Round 1 completes, enters Round 2 (challenger) and Round 3 (scenarios)
- User can exit deep mode at any time ("I'm satisfied with v1") — the express output is always valid

**No breaking changes:** Express produces exactly what the current system produces. Deep adds optional iteration on top.

### 6. Stopping Heuristic

**Breaks are the natural decision points.** Users don't need to think about when to stop — the pomodoro structure creates built-in checkpoints.

Round 2 (Challenger) — two pomodoros:
- **Pomodoro 2A** always runs (25 min). Minimum: 3 challenges per Core context.
- **Break after 2A:** AI shows diff. If diff is tiny (< 3 changes), suggest skipping 2B. User decides.
- **Pomodoro 2B** runs if user continues. Focuses on what the break review surfaced.

Round 3 (Simulator) — one mandatory + one optional pomodoro:
- **Pomodoro 3A** always runs (25 min). Minimum: 1 happy path + 1 failure per Core context.
- **Break after 3A:** AI shows remaining gap count. If zero gaps, suggest stopping. User decides.
- **Pomodoro 3B** is optional. Only runs if user wants deeper validation (scale, cross-context).

**Convergence metric:** Track `delta(invariants + terms + stories)` per pomodoro. Display at each break:
```
Pomodoro 2A: +5 invariants, +3 terms, +2 stories (active refinement)
Pomodoro 2B: +1 invariant, +1 term, +0 stories (stabilizing)
Pomodoro 3A: +0 invariants, +0 terms, +1 story (converged — consider stopping)
```

**Hard stop:** After 4 pomodoros total (excluding Round 1), session ends regardless. Diminishing returns beyond ~2 hours.

### 7. Competitive Landscape

| Tool | Approach | Strengths | Weaknesses vs alto |
|------|----------|-----------|-------------------|
| **Qlerify** | Visual AI-powered EventStorming + code gen | Visual board, real-time collab, code generation, Jira integration | SaaS (not local), visual-first (alto is CLI), generates code (alto generates structure only), no iterative challenger role |
| **Context Mapper** | DSL-based DDD modeling | Formal CML language, generates diagrams, API contracts | Manual (no AI assist), steep learning curve, no guided discovery |
| **Ruflo** | Multi-agent DDD swarms for Claude | 8 specialized agents (Domain Expert, Aggregate Designer, Context Mapper...), CLAUDE.MD templates | Agent orchestration platform (not bootstrapper), requires Claude specifically, no guided question flow, no iteration protocol |
| **Generic LLM prompting** | "Design bounded contexts for X" | Fast, flexible, any LLM | No structure, no iteration, no human-in-the-loop gates, no artifact format standard |

**alto's differentiation:**
1. **Guided iteration, not one-shot generation** — Qlerify and LLM prompting generate artifacts in one pass. alto iterates with challenger + scenarios.
2. **CLI-native, local-first** — No SaaS dependency. Works offline after initial research.
3. **Complexity-budget-aware depth** — Only Core subdomains get deep treatment. Others get appropriate shallowness.
4. **Human-in-the-loop at every step** — AI suggests, human confirms. Never auto-generates without approval.
5. **Bounded Context Canvas as standard output** — None of the competitors produce the ddd-crew canvas format.

### 8. Academic Research Landscape (2025-2026)

No arxiv papers directly study AI-assisted EventStorming or Domain Storytelling. However, adjacent research validates alto's design decisions and reveals key risks.

#### Directly Relevant Papers

| Paper | Authors | Date | Key Finding | alto Implication |
|-------|---------|------|-------------|-----------------|
| **Leveraging Generative AI for Enhancing Domain-Driven Software Design** ([arxiv:2601.20909](https://arxiv.org/abs/2601.20909)) | Wiegand, Stepniak, Baier | Jan 2026 | Fine-tuned Code Llama generates DDD metamodel JSON with 0.99 BLEU score, but 19/50 ambiguous prompts produce parsing errors. Repetitive generation until token limit is a failure mode. | DDD artifact generation works for well-specified inputs but breaks on ambiguity. **alto must not auto-generate DDD artifacts from ambiguous user answers — always confirm.** |
| **Software Architecture Meets LLMs: A Systematic Literature Review** ([arxiv:2505.16697](https://arxiv.org/abs/2505.16697)) | Schmid et al. (KIT) | May 2025 | 18 papers analyzed. Most use basic prompting. Gaps: code generation from architecture, conformance checking. LLMs "frequently surpass baselines" on architectural tasks. | Validates that LLMs can do architecture-level reasoning. **alto's challenger and simulator roles are architecturally sound.** Gap in conformance checking aligns with alto's fitness function approach. |
| **LLM-Assisted Architecture Design Using ADD** ([arxiv:2506.22688](https://arxiv.org/abs/2506.22688)) | Cervantes, Kazman, Cai | Jun 2025 | LLM + explicit methodology docs + persona definitions → architectures "closely aligned with established solutions." But: "human oversight and iterative refinement remain essential." | Directly validates alto's approach: feed the LLM explicit DDD methodology + structured prompts + human-in-the-loop. **The iterative refinement finding confirms Rounds 2-3 design.** |
| **Elicitron: LLM Agent-Based Simulation for Requirements Elicitation** ([arxiv:2404.16045](https://arxiv.org/abs/2404.16045)) | Ataei et al. | Apr 2024 (cited in 2025 surveys) | LLM agents simulate diverse users for requirements elicitation. Context-aware agent generation → more latent needs discovered than human interviews. | Validates alto's Customer Simulator (Role 4). **Context-aware generation is key — feed the simulator the full DDD.md, not just scenario prompts.** |
| **LLMs for Requirements Engineering: SLR** ([arxiv:2509.11446](https://arxiv.org/abs/2509.11446)) | Zadenoori et al. | Sep 2025 | 74 studies reviewed. GPT models dominate. Shift from defect detection → elicitation and validation. "Limited use in industry settings." | Field is still exploratory. **alto's guided approach (structured questions, not open-ended generation) is more reliable than end-to-end LLM generation.** |

#### Domain Knowledge & Hallucination Papers

| Paper | Authors | Date | Key Finding | alto Implication |
|-------|---------|------|-------------|-----------------|
| **RvLLM: LLM Runtime Verification with Domain Knowledge** ([arxiv:2505.18585](https://arxiv.org/abs/2505.18585)) | Zhang et al. | May 2025 | Domain experts define constraints in a specification language (ESL). LLM outputs are verified at runtime against these predicates. Better TPR/TNR balance than baselines. | **Domain Researcher outputs should be verified against user-confirmed facts.** User's answers are the "domain constraints"; AI research that contradicts them must be flagged, not silently applied. |
| **Beyond the Prompt: Domain Knowledge Strategies for LLM Optimization in SE** ([arxiv:2602.02752](https://arxiv.org/abs/2602.02752)) | Srinivasan, Menzies | Feb 2026 | Compares four strategies: Human-in-the-Loop (H-DKP), Adaptive Multi-Stage Prompting, Progressive Refinement, Hybrid RAG. High-dimensional SE tasks (>11 features) still favor Bayesian methods. | Confirms that pure LLM approaches struggle with complex domains. **alto's human-in-the-loop design is the right call for DDD modeling.** |
| **Key Considerations for Domain Expert Involvement in LLM Design** ([arxiv:2602.14357](https://arxiv.org/abs/2602.14357)) | Szymanski et al. | Feb 2026 | 12-week ethnographic study. Key practices: workarounds for data collection, augmentation when expert input is limited, co-developed evaluation criteria, hybrid expert-developer-LLM evaluation. | **When the user (domain expert) provides limited input, AI augments — not replaces — their knowledge.** Evaluation criteria should be co-developed with the user. |
| **Mitigating Hallucination: RAG, Reasoning, and Agentic Systems** ([arxiv:2510.24476](https://arxiv.org/abs/2510.24476)) | Survey | Oct 2025 | Three mitigation paradigms: (1) RAG grounds responses in retrieved knowledge, (2) reasoning enhancement improves logical consistency, (3) agentic systems enable self-correction through multi-agent debate. | alto uses all three: Domain Researcher does RAG, Challenger does reasoning, multi-role design enables cross-checking. **Add explicit source attribution to every AI-generated claim.** |

#### Key Academic Insight: No One Has Done This Yet

The literature search reveals that **no published paper combines AI-assisted DDD discovery with iterative human-in-the-loop refinement.** Existing work falls into two buckets:

1. **Full automation** — Generate DDD artifacts from text (arxiv:2601.20909). No iteration, no human feedback loop.
2. **Human workshops with AI support** — Use AI to transcribe, summarize, or visualize workshop outputs. No AI challenger or simulator role.

alto's three-round protocol (Express → Challenge → Simulate) with human gates at every pomodoro break is a novel combination. The closest analog is Elicitron's simulated user interviews, but applied to product requirements, not DDD modeling.

### 9. Domain Knowledge Quality & Anti-Hallucination Design

This section addresses the critical question: **How do we handle domain knowledge when internet data may be incomplete, outdated, or wrong — and AI models can hallucinate?**

#### Core Principle: User Is the Oracle, AI Is the Challenger

The fundamental design decision: **the AI never states domain facts. It only asks questions and surfaces evidence.** The user (domain expert / product owner / developer) is the single source of truth for domain-specific knowledge.

```
WRONG approach:
  AI: "In pet sitting marketplaces, sitters typically charge $15-25/hour
       and escrow periods are usually 48 hours."
  (← hallucination risk: numbers may be wrong, domain may differ)

RIGHT approach:
  AI: "I found that some pet sitting platforms use escrow periods between
       24-72 hours. What escrow period makes sense for YOUR marketplace?"
  (← AI surfaces a range, user decides the actual value)
```

#### The Knowledge Trust Hierarchy

Every piece of information in the session has a trust level:

```
Level 1: USER-STATED (highest trust)
  Source: Direct user answers to discovery questions
  Treatment: Accept as ground truth. Never contradict without evidence.
  Example: "Our escrow period is 48 hours"

Level 2: USER-CONFIRMED
  Source: AI-suggested fact that the user explicitly confirmed
  Treatment: Treat as ground truth after confirmation
  Example: AI: "Should sitters be verified before accepting bookings?"
           User: "Yes, background check required"

Level 3: AI-RESEARCHED (requires attribution)
  Source: Web search, competitive analysis, domain patterns
  Treatment: Present WITH source URL. Mark as "[researched]" in artifacts.
             User must confirm before it enters the domain model.
  Example: "[Source: Rover.com] Competing platforms use a 24h review
            window after sitting completes. Relevant for your design?"

Level 4: AI-INFERRED (lowest trust, requires challenge)
  Source: LLM reasoning from existing context, no external source
  Treatment: Present as a QUESTION, not a statement. Mark as "[inferred]"
             in artifacts. Must be confirmed or rejected by user.
  Example: "[Inferred] Based on your booking flow, it seems like you
            need a 'BookingExpired' event for requests that go unanswered.
            Is that correct?"
```

#### RLM Over RAG for Domain Research

alto already uses RLM (Recursive Language Model) patterns for knowledge retrieval (see `docs/RLM.md`). RLM is strictly better than RAG for domain research because:

- **RAG** does a one-shot retrieval: embed query → find top-k chunks → generate answer. If the first retrieval misses critical context, the answer is wrong.
- **RLM** iterates: search → read → reason → search again based on what was found. The LLM generates code to navigate dependency chains, refining context at each step.

For the Domain Researcher role, the RLM loop would look like:

```
RLM Loop for Domain Research:

Step 1: Search "{domain} business model patterns"
  → Finds: marketplace, SaaS, B2B — identifies it as a marketplace
Step 2: Search "{domain} marketplace trust mechanisms"
  → Finds: reviews, verification, escrow — matches user's answers
Step 3: Search "{domain} marketplace failure modes"
  → Finds: cancellation policies, dispute resolution, fraud
  → AI doesn't know these yet — surfaces them as challenger questions
Step 4: Search "{competing product} how they handle {gap from step 3}"
  → Finds specific competitor approaches with sources
```

**However: RLM does NOT solve hallucination.** Even with perfect retrieval, the LLM can still:
- Misinterpret retrieved content
- Conflate patterns from different domains
- Generate plausible-sounding invariants that don't apply to the user's domain
- Extrapolate beyond what sources actually say

This is why the Trust Hierarchy (Section 9) and human confirmation gates are non-negotiable. RLM improves the quality of retrieved knowledge, but every AI-sourced fact still requires user confirmation before entering the domain model.

#### Anti-Hallucination Patterns for Each AI Role

**Domain Researcher (Role 3) — uses RLM:**

```
Pattern: RLM-SEARCH → ATTRIBUTE → PRESENT → CONFIRM

1. RLM-SEARCH: Iterative search using RLM REPL loop.
   Start broad ("{domain} patterns"), narrow based on findings.
   Each iteration generates targeted follow-up searches.
   Store retrieved content as RLM-addressable knowledge chunks.

2. ATTRIBUTE: Every fact gets a source URL or "no source found"
   - If no source → don't present as fact
   - If source is >2 years old → mark as "[may be outdated]"
   - If multiple sources agree → higher confidence
   - If sources contradict → present both, let user decide

3. PRESENT: Show as "Here's what I found" with sources,
   never as "Here's how it works"

4. CONFIRM: User says "relevant" / "not relevant" / "partially relevant"
   - Only "relevant" facts enter the domain model
   - "Not relevant" facts are discarded
   - "Partially relevant" facts are adapted by the user

Fallback when internet data is insufficient:
  AI: "I couldn't find reliable information about [specific aspect].
       This seems unique to your domain. Can you describe how [X] works
       in your context?"
  → User becomes the primary source
  → AI records as Level 1 (USER-STATED)
```

**Challenger (Role 2):**

```
Pattern: CHALLENGE-AS-QUESTION, never CHALLENGE-AS-FACT

WRONG: "Your aggregate design is incorrect because aggregates
        should be small." (← presenting opinion as DDD law)

RIGHT: "Your Booking aggregate contains 8 entities. DDD
        practitioners like Vaughn Vernon recommend keeping aggregates
        small (1-3 entities) for consistency boundaries. Which entities
        MUST be consistent within the same transaction?"
        [Source: Vernon, 'Implementing DDD', Ch. 10]
        (← cites source, asks user to decide)
```

**Customer Simulator (Role 4):**

```
Pattern: SCENARIO-FROM-MODEL, not SCENARIO-FROM-IMAGINATION

1. Generate scenarios ONLY from entities, events, and workflows
   already in DDD.md — don't invent new domain concepts

2. Use concrete but FICTIONAL names and data
   (Alice, Bob, Miso the cat — not real businesses)

3. When a scenario reveals a gap, ask:
   "Your model doesn't cover what happens when [X].
    Is this: (a) a real scenario we should handle,
             (b) out of scope, or
             (c) something your model already handles
                  but I missed?"

4. Never assume the answer — always ask
```

#### Handling Domain-Specific Knowledge Gaps

For niche domains where internet data is sparse:

```
Preparation Phase (before Round 2):

1. AI searches for domain-specific resources:
   - Industry standards and regulations
   - Competitor websites and product descriptions
   - Domain-specific forums, blogs, documentation
   - Academic papers on the problem domain

2. AI categorizes findings:
   ├── HIGH confidence: Multiple independent sources agree
   ├── MEDIUM confidence: Single authoritative source
   ├── LOW confidence: Blog posts, opinions, single data points
   └── NO DATA: Nothing reliable found

3. AI presents a "Research Briefing" at the start of Pomodoro 2A:
   "I researched [domain]. Here's what I found with high confidence:
    [list with sources]. For these areas, I couldn't find reliable
    data: [list]. You'll be the primary source for those."

4. For NO DATA areas, AI switches to pure Socratic questioning:
   "I don't have reliable data on [X]. Let me ask:
    - How does [X] work in your experience?
    - What are the rules around [X]?
    - What can go wrong with [X]?"
   → These become USER-STATED facts (highest trust)
```

#### Artifact Provenance Tracking

Every fact in DDD.md v2/v3 carries its provenance:

```markdown
## Ubiquitous Language

| Term | Definition | Source | Confidence |
|------|-----------|--------|------------|
| Booking | A confirmed reservation for pet sitting services | User (Q3) | ●●● user-stated |
| Escrow | Payment held by platform until sitting completes | User (Q3) + [Rover.com](https://rover.com) | ●●● user-confirmed |
| Reliability Score | Numeric rating based on cancellation history | AI-inferred from Q7 policies | ●○○ needs confirmation |
```

This ensures:
- Users know which facts came from them vs. AI
- Reviewers can assess which parts of the model are grounded
- Facts that were never confirmed are explicitly marked
- Over time, all Level 3-4 facts should be elevated to Level 1-2 through user confirmation

#### The "I Don't Know" Protocol

A critical anti-hallucination measure: **the AI must say "I don't know" rather than guess.**

```
Triggers for "I don't know":
- Domain-specific pricing, rates, or thresholds
- Legal or regulatory requirements
- Industry-specific operational details
- Business model specifics (margins, costs, partnerships)
- Technical constraints of existing systems the user is integrating with

AI response template:
"I don't have reliable information about [X]. This is something
 specific to your domain/business. Can you tell me:
 - [specific question about X]?"
```

This is especially important because DDD modeling requires **precise domain language and rules**. A hallucinated invariant ("orders must be fulfilled within 24 hours") that sounds plausible but is wrong will poison the entire aggregate design and all downstream tickets.

## Recommendation

### Session Structure
Adopt the **three-round protocol** (Express → Challenge → Simulate) with complexity-budget gating. Express mode preserves backward compatibility. Deep mode is opt-in.

### Implementation Priority
1. **Bounded Context Canvas generation** — Highest value, lowest risk. Can ship independently. Add canvas section to DDD.md artifact generation. No new AI roles needed.
2. **AI Challenger (Round 2)** — Core innovation. Requires prompt engineering and iteration protocol. Ship after canvas.
3. **AI Domain Researcher** — Background research before Round 2. Enhances challenger quality. Ship with or after challenger.
4. **Customer Simulator (Round 3)** — Validates the model. Ship last — depends on deep enough model from Round 2.
5. **Express vs Deep mode selection** — UX wrapper. Ship when Round 2 is ready.

### Architecture Impact
- `DiscoverySession` aggregate gets a new `DiscoveryMode` enum (EXPRESS/DEEP) and new states for rounds 2-3
- New domain services: `ChallengerService`, `DomainResearchService`, `ScenarioSimulatorService`
- New port: `DomainResearchPort` (adapter does web search)
- Artifact generation handlers extended to produce Bounded Context Canvas sections
- **No breaking changes** to existing flow — Express mode = current behavior

## References

### DDD Practice & Tools
- [DDD Starter Modelling Process](https://ddd-crew.github.io/ddd-starter-modelling-process/)
- [Bounded Context Canvas v5](https://github.com/ddd-crew/bounded-context-canvas)
- [Aggregate Design Canvas](https://github.com/ddd-crew/aggregate-design-canvas)
- [Qlerify — AI-Powered DDD Tool](https://www.qlerify.com/domain-driven-design-tool)
- [Ruflo — Agent Orchestration with DDD](https://github.com/ruvnet/ruflo)
- [Context Mapper](https://contextmapper.org/)
- [Event Storming + AI — Tania Storm](https://medium.com/@tanstorm/event-storming-ia-ddd-turbo-d8f720fccfb0)
- [DDD 2026 Research](docs/research/20260305_ddd_collaborative_modeling_2026.md)
- [Current 10-Question Framework](docs/research/20260222_guided_ddd_question_framework_consolidated.md)

### Academic Papers (2025-2026)
- [Wiegand et al. — Generative AI for DDD Metamodel Generation](https://arxiv.org/abs/2601.20909) (Jan 2026)
- [Schmid et al. — Software Architecture Meets LLMs: SLR](https://arxiv.org/abs/2505.16697) (May 2025)
- [Cervantes, Kazman, Cai — LLM-Assisted Architecture Design Using ADD](https://arxiv.org/abs/2506.22688) (Jun 2025)
- [Ataei et al. — Elicitron: LLM Agent-Based Simulation for Requirements](https://arxiv.org/abs/2404.16045) (2024, cited in 2025 surveys)
- [Zadenoori et al. — LLMs for Requirements Engineering: SLR](https://arxiv.org/abs/2509.11446) (Sep 2025)
- [Zhang et al. — RvLLM: Runtime Verification with Domain Knowledge](https://arxiv.org/abs/2505.18585) (May 2025)
- [Srinivasan & Menzies — Domain Knowledge Strategies for LLM Optimization in SE](https://arxiv.org/abs/2602.02752) (Feb 2026)
- [Szymanski et al. — Domain Expert Involvement in LLM Design: Ethnographic Study](https://arxiv.org/abs/2602.14357) (Feb 2026)
- [Hallucination Mitigation Survey — RAG, Reasoning, Agentic Systems](https://arxiv.org/abs/2510.24476) (Oct 2025)

## Follow-up Tickets

1. **Task: Bounded Context Canvas generation in DDD artifact output** — Add canvas markdown section per context to `alto generate` DDD.md output. Uses existing domain model data. No new AI roles.
2. **Task: AI Challenger service for Round 2 discovery** — New `ChallengerService` domain service. Reads DDD.md v1, generates challenges, records user responses, produces DDD.md v2. Requires `DomainResearchPort`.
3. **Task: Domain Research port + adapter** — `DomainResearchPort` protocol + web search adapter. Background research on problem domain before Round 2.
4. **Task: Customer Simulator for Round 3 scenario testing** — New `ScenarioSimulatorService`. Generates and traces scenarios through domain model. Flags gaps.
5. **Task: Express vs Deep mode in `alto guide`** — `DiscoveryMode` enum, CLI prompt, mode-aware session flow. Express = current, Deep = rounds 2+3.
6. **Task: Iteration diff display and user approval flow** — Show v1→v2→v3 diffs between rounds. User approves/rejects each change.
