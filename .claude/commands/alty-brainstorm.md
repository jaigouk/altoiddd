---
name: alty-brainstorm
description: Guided DDD discovery — turn an idea into PRD, DDD.md, and ARCHITECTURE.md
---

# /alty-brainstorm

Guide the user through structured DDD discovery to produce three foundational documents:
PRD, DDD.md, and ARCHITECTURE.md. Follow the 7 phases below **in order**. Never skip phases.

## Ground Rules

- **One question at a time.** Never ask multiple questions in a single message.
- **Multiple-choice when possible.** Easier to answer than open-ended.
- **Playback every 3 questions.** Summarize in the user's own words, ask for confirmation.
- **Hard gate.** Do NOT generate artifacts until discovery (Phase 3) is complete and confirmed.
- **Ubiquitous language.** Use the user's exact words and terminology — never impose jargon.
- **YAGNI.** Only capture what the user describes. Do not invent features, actors, or events.

---

## Phase 0: Context Scan

Before asking any questions:

1. Read the project README (if it exists).
2. Scan `docs/` for existing PRD.md, DDD.md, ARCHITECTURE.md.
3. Scan for any existing domain models or bounded context references.

**If existing DDD.md is found:**
> "I found an existing DDD.md in this project. Would you like to:
> 1. **Extend** the existing domain model (add new bounded contexts, stories, etc.)
> 2. **Start fresh** (replace existing artifacts)
> 3. **Cancel** (keep everything as-is)"

**If no README exists:**
> "I don't see a README yet. Before we start, describe your project idea in 4-5 sentences.
> What problem does it solve? Who is it for?"

**If README exists:**
Proceed to Phase 1 with the README content as seed context.

---

## Phase 1: Persona Detection

Ask the user:

> "Before we dive in, which best describes you? This helps me adjust my language."
>
> 1. **Developer or technical lead** — I'm comfortable with DDD terminology
> 2. **Product owner or business person** — I think in business outcomes
> 3. **Domain expert** — I know the problem area deeply but I'm not a developer
> 4. **Not sure / mixed** — I wear multiple hats

**Register mapping:**
- Choice 1 → **TECHNICAL** register (use Technical column in Phase 3)
- Choice 2, 3, or 4 → **NON_TECHNICAL** register (use Non-Technical column in Phase 3)

Remember the chosen register for all subsequent phases.

---

## Phase 2: Seed Extraction

Using the README or user description from Phase 0:

1. Extract candidate actors, key concepts, and potential workflows.
2. Present a summary:

> "Here's what I understand so far:
>
> **Problem:** [summarize the problem]
> **Users/Actors:** [list detected actors]
> **Key concepts:** [list detected nouns/entities]
> **Main workflow:** [describe the primary flow if detectable]
>
> Does this capture the core idea? Anything to add or correct?"

Wait for confirmation before proceeding to Phase 3.

---

## Phase 3: Guided Questions (10Q Framework)

Ask questions one at a time using the correct register. Track answers for artifact generation.

**Quick mode:** If the user says they're experienced or want to go fast, use only the 5 MVP
questions marked with ★. Otherwise ask all 10.

### Actors & Work Objects

**Q1** ★
- **Technical:** "Who are the actors (users, external systems) that interact with your system?"
- **Non-Technical:** "Who will use this product, and what other systems does it talk to?"
- *Produces:* actors, external_systems

**Q2**
- **Technical:** "What are the core entities (nouns) in your domain?"
- **Non-Technical:** "What are the main things or concepts your product deals with?"
- *Produces:* entities, value_objects

**🔄 Playback after Q2 (or Q1 in quick mode):**
> "Let me play back what I've heard so far:
> - **Actors:** [list]
> - **Key concepts:** [list]
>
> Did I get that right? Anything to add or correct?"

### Primary Story

**Q3** ★
- **Technical:** "Describe the primary use case as a domain story: actor -> command -> event -> outcome."
- **Non-Technical:** "Walk me through the most important thing a user does, step by step."
- *Produces:* commands, events, domain_story

**Q4** ★
- **Technical:** "What is the most critical failure mode? What invariants must hold?"
- **Non-Technical:** "What could go wrong that would be a serious problem? What rules must never be broken?"
- *Produces:* invariants, failure_modes

**Q5**
- **Technical:** "What other workflows or use cases exist beyond the primary one?"
- **Non-Technical:** "What else can users do with the product besides the main thing?"
- *Produces:* secondary_stories, commands

**🔄 Playback after Q5 (or Q4 in quick mode):**
> "Here's what I understand about the workflows:
> - **Primary story:** [Actor] does [action] which triggers [event] resulting in [outcome]
> - **Critical rules:** [list invariants]
> - **Other workflows:** [list if Q5 was asked]
>
> Did I capture this correctly?"

### Events & Rules

**Q6**
- **Technical:** "What domain events are published when state changes occur?"
- **Non-Technical:** "What important things happen in the system that other parts need to know about?"
- *Produces:* domain_events

**Q7**
- **Technical:** "What policies (event -> command reactions) exist in the system?"
- **Non-Technical:** "When something happens, what automatic actions should follow?"
- *Produces:* policies, reactions

**Q8**
- **Technical:** "What read models or projections does the system need?"
- **Non-Technical:** "What views or reports do users need to see?"
- *Produces:* read_models, projections

**🔄 Playback after Q8:**
> "Here's what I understand about events and automation:
> - **Key events:** [list]
> - **Automatic reactions:** [list policies]
> - **Views/reports needed:** [list]
>
> Sound right?"

### Boundaries

**Q9** ★
- **Technical:** "How would you partition the domain into bounded contexts?"
- **Non-Technical:** "If you split the product into independent teams, what would each team own?"
- *Produces:* bounded_contexts

**Q10** ★
- **Technical:** "Classify each context: core (competitive advantage), supporting (necessary but not differentiating), or generic (commodity)."
- **Non-Technical:** "Which parts are your secret sauce, which are necessary plumbing, and which are off-the-shelf?"
- *Produces:* subdomain_classification

For Q10, guide the user through **Khononov's decision tree** for each bounded context:

> For each area you identified, let's classify it:
>
> 1. **Could you buy it or use an off-the-shelf solution?**
>    - YES → **Generic** (use existing library, SaaS, or standard CRUD)
>    - NO → continue to question 2
>
> 2. **Does it have complex business rules?**
>    - NO → **Supporting** (necessary but straightforward — simpler architecture)
>    - YES → continue to question 3
>
> 3. **If a competitor copied it exactly, would that threaten your business?**
>    - YES → **Core** (your competitive advantage — invest heavily here)
>    - NO → **Supporting**

**🔄 Final playback after Q10:**
> "Here's the complete picture:
> - **Bounded contexts:** [list each with classification]
> - **Core (secret sauce):** [list] — these get full DDD treatment
> - **Supporting (necessary plumbing):** [list] — simpler architecture
> - **Generic (off-the-shelf):** [list] — buy or use existing
>
> Does this classification feel right?"

### Handling Edge Cases

- **"Just build it"** → Explain why 10 minutes of discovery saves days of rework. Offer quick mode (5 questions).
- **User provides a URL or doc** → Read it and use as seed context in Phase 2.
- **Single bounded context** → Accept it. Skip context map details. Still classify it.
- **Non-English domain terms** → Preserve original language in the ubiquitous language glossary.
- **User wants to go back** → Allow revisiting any phase. Re-do the relevant playback.
- **Very large domain (10+ contexts)** → Focus on Core subdomains first. Stub the rest.
- **"I don't know"** → That's fine. Mark the area as needing a spike. Move on.

---

## Phase 4: Approach Proposals

Based on the discovered bounded contexts and their classifications, present 2-3
architecture approaches. Consider:

- Number of bounded contexts
- Team size (ask if not known)
- Complexity budget (how many Core subdomains)

Example approaches:

> Based on what we've discovered, here are the architecture options:
>
> **A. Modular monolith** (recommended for ≤5 contexts, small team)
> - Single deployable with clear module boundaries per bounded context
> - Clean separation via Python packages with import rules
> - Easiest to start, can split later
>
> **B. Modular monolith with event bus** (recommended for 5-10 contexts)
> - Same as A, but contexts communicate via domain events
> - Better isolation, easier to extract services later
>
> **C. Microservices** (recommended for 10+ contexts, multiple teams)
> - Independent deployables per bounded context
> - Highest operational complexity — only if you have the team for it
>
> I'd recommend **[A/B/C]** based on [reasoning]. Which approach works for you?

Wait for the user to choose before proceeding.

---

## Phase 5: Artifact Generation

Generate three documents **section by section**, asking for approval after each section.
Use the templates as structural guides:

### 5.1 PRD (`docs/PRD.md`)

**Template reference:** `docs/templates/PRD_TEMPLATE.md`

Generate each section from discovery answers:

| PRD Section | Source |
|-------------|--------|
| Problem Statement | README + Phase 2 seed |
| Vision | Phase 2 seed + user corrections |
| Users & Personas | Q1 actors |
| User Scenarios | Q3 primary story + Q5 secondary stories |
| Capabilities | All questions — map to Must Have / Should Have / Nice to Have |
| Constraints | Q10 classification + Phase 4 architecture choice |
| Out of Scope | Explicitly excluded during discovery |
| Success Metrics | Derived from Q4 invariants + capabilities |
| Risks & Unknowns | Areas marked "I don't know" + spikes needed |

Present each section and ask: "Does this section look right? Any changes?"

### 5.2 DDD.md (`docs/DDD.md`)

**Template reference:** `docs/templates/DDD_STORY_TEMPLATE.md`

Generate each section:

| DDD Section | Source |
|-------------|--------|
| Domain Stories | Q3 primary + Q5 secondary, written as "[Actor] [verb] [work object]" |
| Ubiquitous Language Glossary | All terms from user's answers — use their exact words |
| Subdomain Classification | Q10 with Khononov rationale |
| Bounded Contexts | Q9 contexts with responsibilities, key objects |
| Context Map | Relationships between contexts from Q6 events + Q7 policies |
| Aggregate Design | For Core subdomains: root, invariants (Q4), commands (Q3), events (Q6) |

**Ambiguous terms:** If the same word appeared with different meanings in different contexts,
document both meanings in the glossary.

Present each section and ask: "Does this section look right? Any changes?"

### 5.3 ARCHITECTURE.md (`docs/ARCHITECTURE.md`)

**Template reference:** `docs/templates/ARCHITECTURE_TEMPLATE.md`

Generate each section:

| Architecture Section | Source |
|---------------------|--------|
| Design Principles | Phase 4 approach + DDD alignment |
| System Overview | Bounded contexts from Q9 as components |
| Layer Architecture | Hexagonal for Core, layered for Supporting, ACL for Generic |
| Bounded Context Integration | Q6 events + Q7 policies → context map relationships |
| Data Model | Q2 entities + aggregate designs from Core subdomains |
| Security | Trust boundaries between contexts |
| Deployment | Phase 4 architecture choice |
| Constraints | From PRD constraints |

**Fitness Functions:** For each bounded context, generate architecture test rules:

- **Core subdomains:** Layer purity (domain has zero external deps), dependency direction
  (inward only), aggregate isolation (reference by ID only), no cross-context imports
- **Supporting subdomains:** Layer purity + basic dependency direction
- **Generic subdomains:** ACL boundary enforcement only

Include these as a "Fitness Functions" section in ARCHITECTURE.md with concrete rules
per bounded context. Do NOT leave this as a template placeholder.

Present each section and ask: "Does this section look right? Any changes?"

---

## Phase 6: Ticket Generation (Optional)

After all three artifacts are approved:

> "Would you like me to generate beads tickets from these DDD artifacts?
>
> 1. **Yes** — Generate dependency-ordered epics and tasks
> 2. **Not now** — I can do this later with `/prd-traceability`
> 3. **No** — I'll create tickets manually"

If yes:
1. Generate one epic per bounded context
2. Order tasks by aggregate dependencies (foundation first)
3. Core subdomains get full ticket detail (AC, TDD phases, SOLID mapping, edge cases)
4. Supporting subdomains get standard detail
5. Generic subdomains get stub tickets ("integrate X", verify boundary)
6. Preview the ticket structure before creating
7. Set formal dependencies via `bd dep add`

---

## Phase 7: Save & Next Steps

After all approved artifacts are written:

> "All artifacts have been saved:
> - `docs/PRD.md`
> - `docs/DDD.md`
> - `docs/ARCHITECTURE.md`
>
> **Recommended next steps:**
> 1. Run `/doc-health` to verify document completeness
> 2. Review and commit the generated artifacts
> 3. Start implementation with the generated tickets (if created)
> 4. Run `/prd-traceability all` to verify coverage"

Ask if the user would like to commit the changes.
