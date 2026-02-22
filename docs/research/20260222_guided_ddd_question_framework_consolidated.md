---
last_reviewed: 2026-02-22
owner: researcher
status: complete
spike: vibe-seed-k7m.2
---

# Guided DDD Question Framework — Consolidated Research Report

## Research Question

What is the minimal effective set of DDD / Domain Storytelling questions to take a 4-5 sentence idea and produce bounded contexts, ubiquitous language, aggregate design, and subdomain classification — for both technical and non-technical users?

## Decision

**Hybrid approach**: Domain Storytelling's concrete-scenario question flow + Event Storming's artifact taxonomy + Khononov's complexity budget decision tree. 10 core questions in 5 phases, delivered in dual-register (technical + non-technical), producing all DDD artifacts needed for architecture in 15-20 minutes.

## Supporting Research Reports

| Report | Focus |
|--------|-------|
| `20260222_ddd_question_framework.md` | Methodology deep-dive: Domain Storytelling, Event Storming, DDD-Crew, solo-dev adaptation, tool landscape |
| `20260222_subdomain_classification_complexity_budget.md` | Classification heuristics: Evans/Vernon/Khononov/Tune, decision tree, treatment levels, real-world examples |
| `20260222_ddd_persona_adapted_discovery.md` | Plain language mapping, dual-register question pairs, non-technical question flow, persona detection |

---

## 1. The 10-Question Framework (Dual Register)

### Persona Detection (before questions start)

```
"Before we start, which best describes you?
 1. Developer or technical lead
 2. Product owner or business person
 3. Domain expert (you know the problem area deeply)
 4. Not sure / mixed"
```

Options 1 → technical register. Options 2-4 → non-technical register.

### Phase 1: Seed (automated, no questions)

AI reads the 4-5 sentence README, extracts candidate actors, work objects, and events. Presents initial understanding for confirmation.

### Phase 2: Actors & Work Objects (2 questions)

| # | Technical Register | Non-Technical Register | Produces |
|---|-------------------|----------------------|----------|
| Q1 | "Who are the actors in this domain? List people and external systems." | "Who are the people involved in this process? Who starts it? Who else participates?" | Actor list → User roles, external systems |
| Q2 | "What are the key domain entities — the things actors create, modify, or exchange?" | "What things do people work with? Documents, orders, items, records, messages?" | Work Object list → Entity/Aggregate candidates |

### Phase 3: Primary Story (3 questions)

| # | Technical Register | Non-Technical Register | Produces |
|---|-------------------|----------------------|----------|
| Q3 | "Describe the primary use case as a flow: Actor → Command → Event → Actor." | "Walk me through the most common scenario from start to finish. Use: '[Who] [does what] [with what].'" | Primary domain story, commands, events |
| Q4 | "What is the primary failure mode? What invariant can be violated?" | "What is the most important thing that can go wrong in this process?" | Error story, invariant candidates |
| Q5 | "What other significant workflows exist? (Names only, we'll detail Core ones later.)" | "Are there other important processes? Just name them — we can explore them later." | Story inventory, scope assessment |

### Phase 4: Events & Rules (3 questions)

| # | Technical Register | Non-Technical Register | Produces |
|---|-------------------|----------------------|----------|
| Q6 | "What domain events occur? List state transitions in past tense." | "What are the key moments where the state of something changes? (e.g., 'order was placed', 'payment was received')" | Domain Events list |
| Q7 | "What policies exist? 'Whenever [event], then [command].'" | "Are there any automatic rules? Whenever X happens, Y must happen?" | Policies, event handlers, domain services |
| Q8 | "What read models does each actor need to make decisions?" | "What information does [actor] need to see before making a decision at step [N]?" | Read Models, query requirements |

### Phase 5: Boundaries & Classification (2 questions)

| # | Technical Register | Non-Technical Register | Produces |
|---|-------------------|----------------------|----------|
| Q9 | "Where do bounded context boundaries fall? Where does the same term mean different things?" | "Are there parts of this that could work completely independently? Parts where different people are responsible, or where the same word means something different?" | Bounded Context candidates |
| Q10 | "Classify each subdomain: Core (build + invest), Supporting (build simple), Generic (buy)." | "Which part is the thing that makes your idea unique? Which parts are standard stuff every system needs?" | Subdomain classification → Complexity Budget |

### Post-Questions: Playback & Validation

After every 3-4 questions, summarize using the user's own words:

> "Let me play back what I heard. [Summary in user's language]. Did I get that right? What did I miss?"

---

## 2. Question-to-Artifact Mapping

| DDD Artifact | Produced By | DDD.md Section |
|-------------|------------|----------------|
| **Ubiquitous Language Glossary** | Q1, Q2, Q3 (exact terms captured from all answers) | Section 2 |
| **Domain Stories** (2-3) | Q3, Q4, Q5 (primary flow, failure flow, story inventory) | Section 1 |
| **Actor List** | Q1 | Section 1 (story actors) |
| **Work Object / Entity List** | Q2 | Section 5 (aggregate contents) |
| **Command List** | Q3 (verbs from stories), Q8 (decision triggers) | Section 6 (commands table) |
| **Domain Events** | Q6 | Section 6 (events table) |
| **Policies** | Q7 | Section 6 (inferred domain services) |
| **Read Models** | Q8 | Section 6 (queries table) |
| **Bounded Context Map** | Q9 | Section 4 |
| **Subdomain Classification** | Q10 → Classification Flow (Section 3 below) | Section 3 |
| **Aggregate Candidates** | Q2 + Q6 (work objects grouped by events they emit) | Section 5 |
| **Invariant Candidates** | Q4, Q7 | Section 5 (aggregate invariants) |

---

## 3. Subdomain Classification Flow (Complexity Budget)

Triggered by Q10 answers. For each identified subdomain:

```
Step 1: BUY question
  "Could you use an existing product for [subdomain] without losing
   what makes your business special?"
  → YES: GENERIC

Step 2: COMPLEXITY question
  "When your team talks about [subdomain], is it mostly storing and
   retrieving data, or are there complex rules and special cases?"
  → Simple/CRUD: SUPPORTING
  → Complex rules: CORE

Step 3: COMPETITOR validation (mandatory)
  "If a competitor copied your [subdomain] exactly, would that
   threaten your business?"
  → YES: confirms CORE
  → NO + was CORE: reconsider as SUPPORTING
```

Source: Khononov's three-step decision tree (LDDD Ch. 10)

### Treatment Levels per Classification

| Dimension | Core | Supporting | Generic |
|-----------|------|------------|---------|
| **Architecture** | Hexagonal (Ports & Adapters) | Simple layered | ACL wrapping external lib |
| **DDD Artifacts** | Full: stories, aggregates, glossary, canvas | Light: entity list, basic flow | None: integration notes |
| **Fitness Functions** | Strict: layers + forbidden + independence + acyclic | Moderate: layers + basic forbidden | Minimal: ACL boundary only |
| **Ticket Detail** | Full AC, TDD phases, SOLID mapping, edge cases | Standard AC, basic tests | Stub: "integrate X", verify boundary |
| **Test Coverage** | >= 90% domain, >= 80% overall | >= 80% | >= 60% (boundary only) |
| **import-linter** | `layers` + `forbidden` + `independence` + `acyclic_siblings` | `layers` + basic `forbidden` | Single `forbidden` (no direct access) |

### Complexity Budget Output Format

```yaml
subdomains:
  - name: "Guided DDD Discovery"
    classification: core
    rationale: "The conversational question framework is the primary differentiator"
    treatment:
      architecture: hexagonal
      testing: comprehensive
      fitness_functions: strict
      ticket_detail: full

  - name: "Knowledge Base"
    classification: supporting
    rationale: "RLM-addressable docs, custom but not the differentiator"
    treatment:
      architecture: layered
      testing: standard
      fitness_functions: moderate
      ticket_detail: standard

  - name: "CLI Framework"
    classification: generic
    rationale: "Use click/typer — commodity"
    treatment:
      architecture: acl_wrapper
      testing: boundary
      fitness_functions: minimal
      ticket_detail: stub
```

---

## 4. Plain Language Mapping (Key Terms)

| DDD Term | Plain Language | Source |
|----------|---------------|--------|
| Bounded Context | "Area of your business" | Bourgau: "nobody understands BC, everyone gets 'Functional Area'" |
| Aggregate | "Key business object" | EventStorming deprecated "aggregate" → "Constraint" for business |
| Entity | "Something you track by its identity" | Business people think in "records" |
| Value Object | "A description or measurement" | "An address is a description, not a tracked thing" |
| Domain Event | "Something important that happened" | EventStorming's core premise |
| Command | "An action someone takes" | EventStorming uses "Action" for business audiences |
| Policy | "Whenever X happens, do Y" | Maps to EventStorming lilac stickies |
| Ubiquitous Language | "The words your team actually uses" | Bourgau: use "shared vocabulary" until concept clicks |
| Core Subdomain | "Your secret sauce" | "What you do that nobody else does" |
| Supporting Subdomain | "Necessary plumbing" | "You need it, but it's not what customers pay for" |
| Generic Subdomain | "Off-the-shelf commodity" | "Everyone needs this, everyone does it the same way" |

---

## 5. Example Application: Pet Sitting Marketplace

### Input (4-5 sentence README)

> "PawWatch connects pet owners who need sitters with trusted, verified pet sitters in their area. Owners describe their pet, choose dates, and browse available sitters with reviews. Sitters set their availability, rates, and accepted pet types. Payments are held in escrow until the sitting is complete. Both sides leave reviews afterward."

### Phase 1: Seed (AI extracts from README)

**Candidate actors:** Pet Owner, Pet Sitter, Payment System
**Candidate work objects:** Pet Profile, Booking, Sitter Listing, Review, Payment
**Candidate events:** Booking Requested, Booking Confirmed, Sitting Completed, Payment Released, Review Posted

### Phase 2: Actors & Work Objects

**Q1 (non-technical):** "Who are the people involved in this process?"
> **Answer:** Pet Owner (searches, books, pays, reviews), Pet Sitter (lists availability, accepts bookings, performs sitting), Admin (handles disputes, verifies sitters), Payment Processor (Stripe — holds and releases funds).

**Q2:** "What things do people work with?"
> **Answer:** Pet Profile (species, breed, needs, medical notes), Sitter Listing (rates, availability, accepted pets, location), Booking (dates, pet, sitter, status), Review (rating, text, photos), Payment (amount, escrow status).

### Phase 3: Primary Story

**Q3:** "Walk me through the most common scenario."
> 1. Pet Owner creates a Pet Profile with details about their pet
> 2. Pet Owner searches for available sitters by date, location, and pet type
> 3. Pet Owner browses Sitter Listings and reads Reviews
> 4. Pet Owner sends a Booking Request to a Sitter
> 5. Sitter reviews the Booking Request and Pet Profile
> 6. Sitter accepts the Booking (Booking Confirmed)
> 7. Pet Owner pays — Payment held in Escrow
> 8. Sitter performs the sitting (Sitting In Progress)
> 9. Sitter marks sitting complete (Sitting Completed)
> 10. Payment released to Sitter (Payment Released)
> 11. Both parties leave Reviews

**Q4:** "What's the most important thing that can go wrong?"
> Sitter cancels last-minute. Owner needs emergency rebooking. Payment must be refunded from escrow. Sitter takes a reliability penalty.

**Q5:** "Other important workflows?"
> Sitter verification (background check, identity), Dispute resolution (owner claims sitter didn't show up), Sitter onboarding (set up profile, availability, rates).

### Phase 4: Events & Rules

**Q6 (events):** Booking Requested, Booking Confirmed, Booking Cancelled (by either side), Sitting Started, Sitting Completed, Payment Escrowed, Payment Released, Payment Refunded, Review Posted, Sitter Verified, Dispute Opened, Dispute Resolved.

**Q7 (policies):**
- Whenever Booking Confirmed → Escrow Payment
- Whenever Sitting Completed → Release Payment (after 24h grace period)
- Whenever Sitter Cancels within 48h → Reliability Penalty applied
- Whenever 3+ Dispute Resolved against Sitter → Sitter Suspended

**Q8 (read models):**
- Owner needs: Sitter availability, ratings, accepted pet types, distance
- Sitter needs: Pet details, owner reviews, booking dates
- Admin needs: Dispute history, sitter verification status

### Phase 5: Boundaries & Classification

**Q9 (bounded contexts):**
- **Booking** — booking lifecycle, matching, scheduling (different "sitter" than in Reviews)
- **Payments** — escrow, release, refund (different "booking" than in Booking — just an amount + reference)
- **Trust & Reviews** — reviews, ratings, sitter verification, dispute resolution
- **Identity** — owner profiles, sitter profiles, authentication

**Q10 (subdomain classification):**

| Subdomain | Classification | Rationale |
|-----------|---------------|-----------|
| Booking & Matching | **Core** | The matching algorithm (availability × pet type × location × ratings) is the differentiator |
| Trust & Reviews | **Core** | Trust is THE marketplace problem; verification + review system is competitive advantage |
| Payments & Escrow | **Supporting** | Escrow logic is custom (can't buy off-shelf for pet sitting) but not the differentiator |
| Identity & Auth | **Generic** | Use Auth0/Clerk; standard user management |
| Notifications | **Generic** | Use SendGrid/Twilio; commodity |

### Generated Artifacts (automated from answers)

**Ubiquitous Language:** Pet Owner, Pet Sitter, Pet Profile, Sitter Listing, Booking Request, Booking Confirmed, Escrow, Sitting Completed, Review, Reliability Penalty, Dispute, Sitter Verified.

**Bounded Contexts:** Booking (Core), Trust (Core), Payments (Supporting), Identity (Generic).

**Aggregates:**
- `Booking` (root) — contains BookingRequest, BookingDates, PetReference, SitterReference. Invariant: "A sitter cannot have overlapping bookings."
- `SitterProfile` (root) — contains Availability, Rates, AcceptedPetTypes. Invariant: "Rates must be > 0."
- `Review` (root) — contains Rating, ReviewText. Invariant: "Review can only be posted after Sitting Completed."

---

## 6. Key Risks

| Risk | Mitigation |
|------|-----------|
| Solo developers over-simplify boundaries (no workshop disagreements to surface them) | AI must probe: "Could [term] mean something different in another part of the system?" |
| Non-technical users over-classify as Core ("everything is important!") | Mandatory competitor-copy validation question |
| 10 questions may feel like too many for impatient users | Minimum viable: 5 questions (Q1, Q3, Q4, Q9, Q10) produce basic bounded context map |
| Text-based loses spatial patterns from sticky notes | AI auto-groups artifacts and presents structured summary |

---

## 7. Follow-Up Implementation Tickets

1. **Question flow engine** — Build the conversational flow (10 questions, 5 phases, dual register, persona detection). Branching logic: if user lists > 5 actors, probe for context boundaries earlier.

2. **Artifact generator** — Transform answers into DDD_STORY_TEMPLATE.md format (ubiquitous language glossary, bounded context map, aggregate candidates, domain events, subdomain classification).

3. **Complexity budget classifier** — From Q10 + classification flow, assign treatment levels. Store as YAML in DDD artifacts. Feed into ticket pipeline and fitness function generation.

4. **Integration: ticket pipeline (k7m.11)** — Generated DDD artifacts become input for auto-generating dependency-ordered beads tickets with treatment-level-appropriate detail.

5. **Integration: fitness functions (k7m.10)** — Bounded context map + subdomain classification drive import-linter contract strictness and pytestarch rule generation.

6. **Question branching / adaptive depth** — Optional deepening questions for complex domains. If user names > 3 workflows in Q5, detail Core ones. If classification is ambiguous, ask Tune's differentiation questions.

---

## Sources

See supporting reports for full source lists:
- `20260222_ddd_question_framework.md` — Domain Storytelling, Event Storming, DDD-Crew sources
- `20260222_subdomain_classification_complexity_budget.md` — Evans, Vernon, Khononov, Tune sources
- `20260222_ddd_persona_adapted_discovery.md` — Bourgau, DDD-Crew glossary, SingleStone sources
