---
last_reviewed: 2026-03-23
owner: researcher
status: complete
spike: alty-cli-jcf
---

# Spike Report: gstack UX Patterns + Domain Storytelling for alto Discovery Redesign

**Date:** 2026-03-23
**Spike:** alty-cli-jcf
**Thesis:** gstack owns "what to build" (scope, product vision, ambition). alto owns "how to structure it" (DDD, bounded contexts, tickets).

---

## Table of Contents

1. [Executive Summary](#1-executive-summary)
2. [gstack Interaction Patterns (RQ1)](#2-gstack-interaction-patterns-rq1)
3. [Domain Storytelling for AI-Guided Discovery (RQ2)](#3-domain-storytelling-for-ai-guided-discovery-rq2)
4. [AI Domain Expert Agent (RQ3)](#4-ai-domain-expert-agent-rq3)
5. [gstack -> alto -> beads Pipeline (RQ4)](#5-gstack---alto---beads-pipeline-rq4)
6. [alto Current Flow: What to Keep, Change, Add](#6-alto-current-flow-what-to-keep-change-add)
7. [Proposed Discovery Flow](#7-proposed-discovery-flow)
8. [Domain Model Changes](#8-domain-model-changes)
9. [Documentation Artifacts (Critical)](#9-documentation-artifacts-critical)
10. [Follow-Up Tickets](#10-follow-up-tickets)
11. [Supporting Research](#11-supporting-research)

---

## 1. Executive Summary

### The Problem

alto's README promises "20 minutes of questions saves you 20 hours of rewrites." But the current 10-question fixed flow feels like a survey, not a conversation. Users with vague ideas hit a wall. Users with domain knowledge feel interrogated rather than guided.

### What We Learned

**From gstack:** The interaction model matters more than the questions. gstack's skills feel natural because they follow a consultant pattern — propose a complete answer, invite adjustment. The key patterns: (1) auto-infer from codebase before asking, (2) one decision per interaction with opinionated recommendation, (3) progressive phases that activate only when needed, (4) coherence validation when the user changes something.

**From Domain Storytelling research:** Domain Storytelling is the right primary technique for alto. It is sentence-based ("Actor does Activity using Work Object"), text-native, requires no special notation training, and maps directly to DDD concepts. The 2026 breakthrough: Annegret Junker (codecentric, March 2026) proved that DST artifacts fed directly to LLMs produce measurably better code — but DST alone misses business rules. You need a complementary technique for invariants.

**From alto's current code:** The domain model (Session, Question, Flow) is well-structured but designed for a questionnaire, not a conversation. The key missing abstractions: conversation narrative, domain story as value object, intermediate synthesis checkpoints, and bounded context sketches built incrementally.

### The Design

```
gstack /plan-ceo-review          alto discovery              alto ticket pipeline
========================    ======================    ==========================
Scope challenge              Domain Storytelling       Beads epics + tasks
10-star product vision       Bounded contexts          Dependency ordering
Mode selection               Subdomain classification  Template compliance
CEO plan artifact            DDD artifacts             AC from invariants
         |                          |                          |
         v                          v                          v
   "what to build"         "how to structure it"       "what to implement"
```

---

## 2. gstack Interaction Patterns (RQ1)

### 5 Patterns to Adopt

#### Pattern 1: Consultant Proposal (Replace Form Wizard)

**gstack (design-consultation):**
> "Design consultant, not form wizard. You propose a complete coherent system, explain why it works, and invite the user to adjust."

- Phase 0: Auto-infer from README, package.json, codebase
- Phase 1: Single compound question pre-filled from inference
- Phase 3: Complete proposal with SAFE/RISK breakdown
- User adjusts sections, not answer blanks

**alto today:** "Who are the actors that interact with your system?" (asks into void)
**alto should:** "From your README, I see this is a veterinary clinic system. The main actors are probably: Pet Owner, Veterinarian, Receptionist, and your Billing System. Sound right? Anything I missed?"

#### Pattern 2: One Decision Per Interaction

**gstack (plan-eng-review):**
> "STOP. For each issue found, call AskUserQuestion individually. One issue per call. Present options, state your recommendation, explain WHY. Do NOT batch."

**Escape hatch:** "If an issue has an obvious fix with no real alternatives, state what you'll do and move on — don't waste a question."

**alto today:** Playback batches 3 answers and asks "confirm?"
**alto should:** After each story sentence, validate individually. After each phase, synthesize and confirm.

#### Pattern 3: Progressive Phase Activation

**gstack (design-consultation):**
- Phase 2 (Research): "only if user said yes"
- Phase 4 (Drill-downs): "only if user requests adjustments"
- Graceful degradation: rich research -> web search -> built-in knowledge

**alto today:** All 10 questions asked regardless.
**alto should:** Offer modes upfront. Skip optional phases. Research domain only if user needs it.

#### Pattern 4: Opinionated Recommendations with Rationale

**gstack (shared preamble):**
> "RECOMMENDATION: Choose [X] because [one-line reason]. Include Completeness: X/10 for each option."

**alto today:** No recommendations. Just records answers.
**alto should:** "I'd model Appointment as an aggregate root because it has clear lifecycle rules (created -> confirmed -> completed -> cancelled) and enforces the invariant that a vet can't have overlapping appointments. Agree?"

#### Pattern 5: Coherence Validation

**gstack (design-consultation):**
> "When user overrides one section, check if rest still coheres. Flag mismatches with a gentle nudge — never block."

**alto today:** Each answer stands alone.
**alto should:** "You mentioned 'Order' in both the Sales story and the Fulfillment story, but with different properties. That's a classic bounded context signal — 'Order' means different things in each context. Should I split them?"

### AskUserQuestion Format (from gstack preamble)

Every user interaction should follow this structure:

1. **Re-ground:** State what we're working on and where we are in the flow
2. **Simplify:** Plain English a smart 16-year-old could follow
3. **Recommend:** Opinionated, with rationale
4. **Options:** Lettered, self-contained

> Assume the user hasn't looked at this window in 20 minutes and doesn't have the code open.

---

## 3. Domain Storytelling for AI-Guided Discovery (RQ2)

### Why Domain Storytelling Over Event Storming

| Dimension | Domain Storytelling | Event Storming |
|-----------|-------------------|----------------|
| Format | Narrative sentences | Sticky note timeline |
| Starting point | "Who does what with what?" | "What events happen?" |
| Cognitive load | Low — natural speech | Medium — learn color scheme |
| Text-native | Yes — sentences are the artifact | No — spatial arrangement is key |
| AI moderator fit | Excellent — structured Q&A | Poor — chaotic brainstorm phase |
| Domain knowledge needed | Low — tell stories you know | Medium — think in events |
| Bounded context discovery | Story groups, handoffs, vocabulary shifts | Event clusters, pivot events |

**Community consensus (2025-2026):** Domain Storytelling first, then Event Storming for complex areas.

**Source:** Kalele.io, Axxes, Junker (codecentric), DDD Europe 2026 program

### The Sentence Structure

```
Actor  -->  Activity  -->  Work Object  [-->  (with) Actor]
(who)       (does what)    (with what)        (with whom)
```

Each story: ONE concrete scenario. No branching. Alternatives = separate stories.

### Three Scope Dimensions

| Dimension | Options | alto Usage |
|-----------|---------|------------|
| **Granularity** | Coarse (overview) / Fine (detailed) | Coarse for context discovery; fine for core subdomains |
| **Time** | As-Is (current) / To-Be (future) | To-Be for new projects; As-Is for rescue mode |
| **Domain Purity** | Pure (no software) / Digitalized (includes systems) | Pure first, digitalized for architecture |

### Minimum Viable Story Set

Stefan Hofer: "Maybe two, three examples are enough to really understand a business process."

| Story | Type | Purpose | DDD Artifacts |
|-------|------|---------|--------------|
| **Story 1** | Primary happy path (coarse) | Actors, work objects, main flow | Actor list, entity candidates, primary contexts |
| **Story 2** | Primary failure case (coarse) | Invariants, error handling | Business rules, invariant candidates |
| **Story 3** | Secondary workflow (coarse) | Additional contexts, relationships | Context map, relationship types |
| **Story 4** (optional) | Core context happy path (fine) | Aggregates, commands, VOs | Tactical DDD design for core |

### Mapping to DDD Concepts

| Story Element | DDD Artifact | Detection |
|--------------|-------------|-----------|
| Actors | User roles, context entry points | Who appears in stories |
| Work Objects | Entities / Aggregates | Created, modified, exchanged |
| Work Object properties | Value Objects | Attributes mentioned |
| Activities | Commands / Use Cases | Verbs in sentences |
| Sequence numbers | Domain Events (implicit) | State transitions between steps |
| Groups / clusters | Bounded Context candidates | Activity clusters |
| Handoffs between groups | Context Map relationships | Work objects crossing boundaries |
| Vocabulary differences | Context boundary signals | Same term, different meaning |
| Annotations | Business rules / Invariants | "Only if...", "Must be..." |

### Bounded Context Detection Heuristics

Three primary signals (Hofer & Schwentner, DDD Europe 2018):

1. **One-way information flow** — info flows in one direction between story sections
2. **Language differences** — same concept has different names/meanings
3. **Different triggers** — work patterns vary by time, event, frequency

Additional signals:
- Organizational boundary (different departments)
- Handoff points (work objects passed between actor groups)
- Different granularity needs
- Same work object, different properties in different stories

### 2026 Breakthrough: DST as LLM Input

Annegret Junker (codecentric, March 4, 2026) demonstrated:

```
Domain Storytelling -> Event Storming -> OpenAPI -> LLM Code Generation
```

| Version | Input | LLM Output Quality |
|---------|-------|-------------------|
| V1 (Stories only) | Story diagrams | Missed business rules, collapsed concepts. 3 schema types. |
| V2 (+ Event Storming) | + Events, commands, policies | 9 schema types. Much better. Still needed refinement. |
| V3 (+ Bounded OpenAPI) | + Machine-readable contracts | Faithful generation without extra instruction. |

**Key technique:** Constrain the LLM: "Do not add any features or concepts not visible in the artifacts." This prevents hallucination.

**The headline:** "A well-facilitated Domain Storytelling session is the first prompt for your prototype."

**Implication for alto:** Domain stories should be structured as LLM context, not just documentation. The format must be both human-readable AND machine-parseable.

**Source:** [codecentric.de — From Stories to Code](https://www.codecentric.de/en/knowledge-hub/blog/from-stories-to-code-how-domain-storytelling-and-eventstorming-give-llms-the-context-they-need)

### What DST Does NOT Capture

DST captures structure (actors, objects, activities) but misses:
- **Business rules** — "Cook cannot rate own Recipe"
- **Policies** — event -> command reactions
- **Read models** — what views/reports users need
- **Domain events** — implicit in sequence numbers, not explicit

alto's Challenge phase (P1 iterative discovery) should extract these. This maps to our existing Q6-Q8 (Events phase) — those questions don't go away, they become a targeted follow-up for core subdomains.

---

## 4. AI Domain Expert Agent (RQ3)

### The Problem

gstack works because Garry Tan bakes his extensive product/startup experience into the skill prompts. The agent draws on that knowledge to make opinionated proposals.

Our users' domains vary: veterinary clinics, HR systems, logistics, fintech. We can't pre-bake expertise. But we can bootstrap it.

### The Pattern (modeled on design-consultation Phase 2)

```
1. User: "I want to build a veterinary clinic management system"

2. alto: "Want me to research this domain first so I can propose
   better stories? I'll look at how veterinary practices typically
   work, common workflows, and existing software in the space."

   A) Yes, research first (takes 2-3 min)
   B) No, I know the domain — let's start telling stories
   C) I have partial knowledge — research and I'll correct

3. If yes: AI researches via web search
   - Industry practices (appointment scheduling, medical records, billing)
   - Regulatory requirements (HIPAA/privacy, prescription tracking)
   - Common workflows (check-in, examination, treatment, follow-up)
   - Existing software (Vetter, ezyVet, IDEXX Neo)
   - Typical entities (Patient/Pet, Owner, Veterinarian, Appointment, MedicalRecord)

4. alto proposes first story:
   "Based on my research, here's the primary flow in most vet clinics:

   Story: 'Pet Owner Brings Pet for Examination'
   1. Pet Owner calls Receptionist to book Appointment
   2. Receptionist checks Vet Schedule for available slots
   3. Receptionist creates Appointment for Pet
   4. Pet Owner arrives with Pet
   5. Receptionist checks in Pet Owner using Appointment
   6. Veterinarian examines Pet using Medical Record
   7. Veterinarian records Findings in Medical Record
   8. Veterinarian prescribes Treatment
   9. Receptionist generates Invoice
   10. Pet Owner pays Invoice

   Does this match how you envision it? What would you change?"

5. User corrects: "Actually, there's a triage step before examination..."
6. alto adapts — never argues, always refines
```

### Knowledge Trust Hierarchy

| Level | Source | Confidence | Example |
|-------|--------|-----------|---------|
| USER_STATED | User's own words | Highest | "In our clinic, the vet tech does triage" |
| USER_CONFIRMED | AI proposed, user agreed | High | "Yes, that flow is correct" |
| AI_RESEARCHED | Web research, cited | Medium | "Most vet clinics use appointment scheduling" |
| AI_INFERRED | Pattern matching, uncited | Low | "This looks like a core subdomain" |

Generated artifacts should tag each element with its trust level:

```
Actors:
  - Pet Owner [user_stated]
  - Veterinarian [user_stated]
  - Vet Tech [user_stated]  # user corrected from "Receptionist"
  - Billing System [ai_researched, source: ezyVet workflow docs]
```

### Academic Validation (2026)

**"Towards Human-in-the-Loop LLM-Enabled Domain Modeling"** (Springer LNCS, 2026):
- LLM generates initial draft model from text
- Rule-based agent engages user through Q&A
- Questions selected based on potential to clarify most uncertain aspects
- "Automated approaches did not surpass human experts but demonstrated superior performance compared to novices"

**Eric Evans (DDD creator, June 2024):**
- "AI-generated models serve as a powerful starting point" but require iterative refinement
- Generic subdomains work well with AI; core domains need human expertise
- Supports alto's complexity budget: AI handles more in Generic/Supporting, humans do more in Core

---

## 5. gstack -> alto -> beads Pipeline (RQ4)

### Handoff Points

```
Phase 1: SCOPE (gstack, optional)       Phase 2: DESIGN (alto)           Phase 3: TICKETS (alto)
================================    ========================    ==========================

/plan-ceo-review                    alto guide                  alto generate tickets
  - Challenge the idea              - Read README + CEO plan    - Epics from contexts
  - Find 10-star product            - Research domain (opt)     - Tasks from stories
  - 4 modes: expand/hold/reduce     - Domain Storytelling       - Spikes for unknowns
  - Per-proposal opt-in             - Boundary detection        - Dependencies via bd dep
                                    - Subdomain classification  - Template compliance
Output:                             - Challenge phase (opt)
~/.gstack/projects/{slug}/                                      Output:
  ceo-plans/{date}-{slug}.md        Output:                     .beads/issues.jsonl
                                    DDD.md                      (beads epics + tasks)
                                    PRD.md
                                    ARCHITECTURE.md
                                    .alto/stories/*.story
```

### alto Must Work Standalone

gstack is optional upstream enrichment. If the user doesn't use gstack:
- alto reads README directly
- alto's own "scope check" phase (lighter than gstack CEO review) asks: "Is this the whole picture or just one piece?"
- Everything else works the same

### Consuming gstack Output

If alto detects gstack CEO plan output:
```bash
# Check for gstack CEO plans
ls ~/.gstack/projects/${SLUG}/ceo-plans/*.md 2>/dev/null
```

If found, read and extract:
- Accepted scope items -> inform story proposals
- Deferred items -> inform "NOT in scope" section of DDD.md
- Vision statement -> inform PRD generation
- Mode used (expansion/hold/reduction) -> adjust discovery depth

---

## 6. alto Current Flow: What to Keep, Change, Add

### Keep (Working Well)

1. **Phase-based organization** (Actors -> Story -> Events -> Boundaries) — maps to DST progression
2. **Session state machine** with explicit status transitions
3. **Skip tracking with reasons** — audit trail for decisions
4. **Flow strategy pattern** (EXPRESS/DEEP/CONVERSATIONAL)
5. **ConversationalFlow** supports dynamic question registration
6. **Dual-register phrasing** — good for different personas
7. **Event emission** on completion (feeds downstream generation)
8. **JSONL agent mode** for AI tool consumption

### Change (Rigid)

1. **10 questions hardcoded** -> Domain Storytelling moderator questions (dynamic, story-driven)
2. **MVP set hardcoded {Q1,Q3,Q4,Q9,Q10}** -> Minimum viable story set (3 stories)
3. **Phase ordering enforcement** -> Story-driven progression (primary -> failure -> secondary -> boundaries)
4. **Playback as Q/A transcript** -> Synthesis checkpoints ("here's what I understand about your domain")
5. **Each answer isolated** -> Narrative being built (each sentence extends the shared understanding)
6. **Persona selection upfront** -> Auto-detect register from language, offer to switch
7. **No follow-up questions** -> Moderator follow-ups ("What happens next?", "Who does that?")

### Add (Missing)

1. **DomainStory value object** — structured story with actors, work objects, sentences, annotations
2. **ConversationNarrative** — turn-by-turn exchange record with synthesis checkpoints
3. **PreliminaryDomainModel** — incrementally built actors, contexts, glossary, invariants
4. **BoundedContextSketch** — tentative context map refined as stories accumulate
5. **UbiquitousLanguageGlossary** — terminology captured as stories are told
6. **StoryCard** — individual sentence with actor, activity, work object, sequence number
7. **TrustLevel** — USER_STATED / USER_CONFIRMED / AI_RESEARCHED / AI_INFERRED per element
8. **Consultant proposal mechanism** — propose then refine, not ask then record
9. **Coherence validation** — flag contradictions between stories
10. **Domain research phase** — optional web research to bootstrap domain knowledge

---

## 7. Proposed Discovery Flow

```
Phase 0: Context Seed (automated)
  ├── Read README (4-5 sentences)
  ├── Check for gstack CEO plan output (optional)
  ├── Auto-infer: domain, probable actors, probable work objects
  └── Present inference: "From your README, I see X for Y. Sound right?"

Phase 1: Mode Selection
  ├── A) RAPID (3 stories, ~15 min) — enough for MVP project setup
  ├── B) THOROUGH (5+ stories, ~30-45 min) — comprehensive domain model
  └── C) EXISTING (codebase inference) — reverse-engineer from code

Phase 2: Domain Research (optional, only if user says yes)
  ├── Web search for industry practices, common workflows, existing software
  ├── Build AI domain knowledge base
  └── Present findings: "Here's what I learned about [domain]. Ready to tell stories?"

Phase 3: Primary Story (Domain Storytelling — coarse-grained)
  ├── AI proposes initial story based on README + research
  ├── User refines sentence by sentence
  ├── AI records, numbers, builds glossary
  ├── Coherence check after each sentence
  └── Produces: Actor list, work object inventory, primary story, glossary v1

Phase 4: Failure Story (Domain Storytelling — coarse-grained)
  ├── AI: "What's the most important thing that can go wrong?"
  ├── User tells the failure story
  ├── AI extracts invariant candidates from failure points
  └── Produces: Invariant candidates, error stories, business rules

Phase 5: Secondary Story (Domain Storytelling — coarse-grained)
  ├── AI: "What else happens in your domain beyond [primary workflow]?"
  ├── User tells secondary story
  ├── AI identifies boundary signals across all stories
  └── Produces: Additional actors/objects, boundary signal inventory

Phase 6: Boundary Discovery (synthesis)
  ├── AI analyzes all stories for boundary signals
  │   ├── One-way information flow
  │   ├── Language differences (same term, different meaning)
  │   ├── Different triggers
  │   └── Organizational boundaries
  ├── AI proposes bounded contexts with rationale
  ├── User validates or adjusts (one context per interaction)
  └── Produces: Bounded context map, context relationships

Phase 7: Subdomain Classification
  ├── AI proposes Core/Supporting/Generic per context
  ├── Uses Khononov decision tree + story complexity signals
  ├── User confirms (one context per interaction)
  └── Produces: Complexity budget

Phase 8 (THOROUGH only): Fine-Grained Stories for Core
  ├── For each Core bounded context:
  │   ├── AI proposes detailed story
  │   ├── User refines
  │   └── Extract: aggregates, commands, value objects
  └── Produces: Tactical DDD design for core subdomains

Phase 9 (THOROUGH only): Business Rules Extraction
  ├── For Core contexts, extract what DST misses:
  │   ├── Domain events (explicit, not just implied by sequence)
  │   ├── Policies (event -> command reactions)
  │   ├── Read models / projections
  │   └── Maps to existing Q6-Q8 but targeted, not blanket
  └── Produces: Event catalog, policy catalog, read model list

Phase 10: Artifact Generation
  ├── Generate DDD.md (bounded contexts, ubiquitous language, invariants)
  ├── Generate PRD.md (user scenarios from stories)
  ├── Generate ARCHITECTURE.md (context map, layer structure)
  ├── Store domain stories in .alto/stories/ (machine-readable)
  ├── Optional: export PlantUML DST (MIT-licensed)
  ├── Optional: export Egon .egn JSON (for visualization)
  └── User reviews and approves
```

---

## 8. Domain Model Changes

### New Value Objects

```go
// DomainStory — the core artifact of Domain Storytelling
type DomainStory struct {
    Title       string           // "Pet Owner Brings Pet for Examination"
    StoryType   StoryType        // coarse_grained | fine_grained
    TimeFrame   TimeFrame        // as_is | to_be
    Purity      DomainPurity     // pure | digitalized
    Trigger     string           // "Pet Owner calls to book appointment"
    Actors      []StoryActor     // Named actors with types
    WorkObjects []WorkObject     // Named objects with types
    Sentences   []StorySentence  // Numbered activities
    Annotations []Annotation     // Business rules, assumptions
    Variations  []string         // Pointers to separate stories
}

// StorySentence — one numbered activity in a story
type StorySentence struct {
    Number        int            // Sequential: 1, 2, 3...
    Subject       string         // Actor performing the activity
    Activity      string         // Domain verb ("books", "examines", "pays")
    Object        string         // Work object acted upon
    Preposition   string         // Optional: "with", "using", "for"
    IndirectObject string        // Optional: second actor or object
    TrustLevel    TrustLevel     // user_stated | user_confirmed | ai_researched | ai_inferred
}

// BoundedContextSketch — tentative context built incrementally
type BoundedContextSketch struct {
    Name            string
    Actors          []string         // Actors that operate in this context
    WorkObjects     []string         // Objects owned by this context
    StorySentences  []SentenceRef    // Which sentences belong here
    BoundarySignals []BoundarySignal // Evidence for this boundary
    Classification  SubdomainType    // Core | Supporting | Generic
    Confidence      float64          // How sure we are about this boundary
}

// BoundarySignal — evidence for a bounded context boundary
type BoundarySignal struct {
    Type        SignalType  // one_way_flow | language_difference | different_trigger | org_boundary
    Description string
    StoryRefs   []string   // Which stories exhibit this signal
}

// UbiquitousLanguageEntry — glossary term captured during storytelling
type UbiquitousLanguageEntry struct {
    Term       string
    Definition string
    Context    string      // Which bounded context owns this term
    Source     TrustLevel  // How we learned this term
    StoryRefs  []string    // Stories where this term appears
}

// ConversationTurn — one exchange in the discovery conversation
type ConversationTurn struct {
    Phase            DiscoveryPhase
    ConsultantAction string         // What alto said/proposed
    UserResponse     string         // What user answered
    Synthesis        string         // "So you're saying..."
    Confirmed        bool           // User agreed with synthesis
    Clarifications   []string       // Any corrections
    ArtifactsProduced []string      // What was captured from this turn
}
```

### Modified Aggregates

```go
// DiscoverySession — updated to support consultation model
type DiscoverySession struct {
    // Existing (keep)
    sessionID    SessionID
    status       DiscoveryStatus
    persona      DiscoveryPersona
    register     DiscoveryRegister
    mode         DiscoveryMode       // RAPID | THOROUGH | EXISTING

    // Replace
    // answers []Answer              // OLD: flat list of Q/A pairs
    stories      []DomainStory       // NEW: structured domain stories
    narrative    []ConversationTurn   // NEW: turn-by-turn record

    // Add
    domainModel  PreliminaryDomainModel  // Built incrementally
    glossary     []UbiquitousLanguageEntry
    contexts     []BoundedContextSketch
    researchBase *DomainResearchBase     // Optional: web research results
}
```

### New Ports

```go
// DomainResearcher — research a domain via web search
type DomainResearcher interface {
    Research(ctx context.Context, domain string) (*DomainResearchBase, error)
}

// StoryProposer — propose domain stories from context
type StoryProposer interface {
    ProposeStory(ctx context.Context, readme string, research *DomainResearchBase,
        existingStories []DomainStory) (*DomainStory, error)
}

// BoundaryDetector — analyze stories for context boundaries
type BoundaryDetector interface {
    DetectBoundaries(stories []DomainStory) ([]BoundedContextSketch, error)
}

// StoryExporter — export stories to various formats
type StoryExporter interface {
    ExportPlantUML(story DomainStory) (string, error)
    ExportEgnJSON(story DomainStory) ([]byte, error)
    ExportText(story DomainStory) (string, error)
}
```

---

## 9. Documentation Artifacts (Critical)

This section documents exactly what alto must capture and persist from the discovery process. These are not optional — they are the source of truth that downstream processes (ticket generation, AI coding tools, architecture fitness) consume.

### 9.1 Domain Story Files (.alto/stories/)

Each story persisted as a structured file:

```yaml
# .alto/stories/01-pet-owner-brings-pet.story.yaml
title: "Pet Owner Brings Pet for Examination"
type: coarse_grained
time: to_be
purity: pure
trigger: "Pet Owner calls to book appointment"

actors:
  - name: Pet Owner
    type: person
    trust: user_stated
  - name: Receptionist
    type: person
    trust: user_confirmed
  - name: Veterinarian
    type: person
    trust: user_stated
  - name: Billing System
    type: system
    trust: ai_researched
    source: "ezyVet workflow documentation"

work_objects:
  - name: Appointment
    type: document
    trust: user_stated
  - name: Medical Record
    type: document
    trust: user_stated
  - name: Invoice
    type: document
    trust: ai_researched

sentences:
  - step: 1
    subject: Pet Owner
    activity: calls
    object: Receptionist
    trust: user_stated
  - step: 2
    subject: Receptionist
    activity: checks
    object: Vet Schedule
    preposition: for
    indirect_object: available slots
    trust: user_confirmed
  - step: 3
    subject: Receptionist
    activity: creates
    object: Appointment
    preposition: for
    indirect_object: Pet
    trust: user_stated

annotations:
  - text: "Only during business hours"
    sentence: 1
    type: constraint
  - text: "Appointment must not overlap with existing appointments for same vet"
    type: invariant
    trust: user_stated

variations:
  - "Pet Owner cancels Appointment" # -> separate story
  - "Emergency walk-in (no appointment)" # -> separate story
```

### 9.2 Ubiquitous Language Glossary (.alto/glossary.yaml)

```yaml
# .alto/glossary.yaml
terms:
  - term: Appointment
    definition: "A scheduled time slot for a pet to see a veterinarian"
    context: Scheduling
    trust: user_stated
    stories: ["01-pet-owner-brings-pet"]

  - term: Medical Record
    definition: "Complete health history of a pet including diagnoses, treatments, vaccinations"
    context: Clinical
    trust: user_confirmed
    stories: ["01-pet-owner-brings-pet", "02-emergency-walk-in"]

  - term: Invoice
    definition: "Billing document generated after treatment"
    context: Billing
    trust: ai_researched
    note: "In Billing context, this is called 'Charge Sheet' not 'Invoice' (language difference signal)"
    stories: ["01-pet-owner-brings-pet"]
```

### 9.3 Bounded Context Map (.alto/context-map.yaml)

```yaml
# .alto/context-map.yaml
contexts:
  - name: Scheduling
    classification: supporting
    confidence: 0.85
    actors: [Pet Owner, Receptionist]
    work_objects: [Appointment, Vet Schedule]
    boundary_signals:
      - type: different_trigger
        description: "Scheduling operates on calendar time; Clinical operates on visit events"
      - type: one_way_flow
        description: "Appointments flow from Scheduling to Clinical, never back"
    stories: ["01-pet-owner-brings-pet:1-3"]

  - name: Clinical
    classification: core
    confidence: 0.90
    actors: [Veterinarian, Vet Tech]
    work_objects: [Medical Record, Treatment, Prescription]
    boundary_signals:
      - type: language_difference
        description: "'Patient' in Clinical = 'Pet' in Scheduling"
    stories: ["01-pet-owner-brings-pet:6-8", "03-follow-up-visit"]

relationships:
  - upstream: Scheduling
    downstream: Clinical
    type: customer_supplier
    shared: [Appointment]
```

### 9.4 Discovery Report (.alto/discovery-report.md)

```markdown
# Discovery Report

**Mode:** RAPID
**Date:** 2026-03-23
**Stories told:** 3
**Bounded contexts identified:** 4
**Invariants captured:** 7

## Stories Summary
1. Pet Owner Brings Pet for Examination (primary happy path)
2. Emergency Walk-In (primary failure/edge case)
3. Follow-Up Visit and Prescription Refill (secondary workflow)

## Boundary Decisions
| Context | Classification | Confidence | Evidence |
|---------|---------------|-----------|----------|
| Scheduling | Supporting | 85% | One-way flow, different trigger |
| Clinical | Core | 90% | Language differences, complex stories |
| Billing | Generic | 80% | Standard process, could be off-the-shelf |
| Inventory | Supporting | 70% | Only appeared in story 3, needs more evidence |

## Trust Distribution
- USER_STATED: 23 elements (58%)
- USER_CONFIRMED: 11 elements (28%)
- AI_RESEARCHED: 4 elements (10%)
- AI_INFERRED: 2 elements (5%)

## What DST Did NOT Capture (needs follow-up)
- Domain events (implicit in sequences, not yet explicit)
- Policies (e.g., auto-notify owner when prescription ready)
- Read models (e.g., daily appointment dashboard)
- These are targeted by Phase 9 (THOROUGH mode) or manual follow-up

## Generated Artifacts
- DDD.md (bounded contexts, ubiquitous language, invariants)
- PRD.md (user scenarios from stories)
- ARCHITECTURE.md (context map, layer structure)
```

### 9.5 Alto Text Format for Domain Stories

For CLI display and conversation, stories render as:

```
Domain Story: "Pet Owner Brings Pet for Examination"
Type: coarse-grained, to-be, pure
Trigger: Pet Owner calls to book appointment

Actors: Pet Owner, Receptionist, Veterinarian, Billing System
Work Objects: Appointment, Vet Schedule, Medical Record, Invoice

1. Pet Owner calls Receptionist
2. Receptionist checks Vet Schedule for available slots
3. Receptionist creates Appointment for Pet
4. Pet Owner arrives with Pet
5. Receptionist checks in Pet Owner using Appointment
6. Veterinarian examines Pet using Medical Record
7. Veterinarian records Findings in Medical Record
8. Veterinarian prescribes Treatment
9. Receptionist generates Invoice
10. Pet Owner pays Invoice

Annotations:
  [1] Only during business hours
  [invariant] Appointment must not overlap for same vet
  [invariant] Payment before treatment release

Variations:
  -> "Pet Owner cancels Appointment" (separate story)
  -> "Emergency walk-in" (separate story)
```

### 9.6 Export Formats

| Format | License | Purpose | Implementation |
|--------|---------|---------|---------------|
| alto text (.story.yaml) | N/A (our format) | Primary storage, machine-readable, human-readable | Native |
| PlantUML DST | MIT (DomainStory-PlantUML) | Diagram generation, documentation | Optional export |
| Egon .egn JSON | N/A (just JSON serialization) | Visualization in Egon.io | Optional export |
| Markdown (in DDD.md) | N/A | Human-readable in docs | Auto-generated |

---

## 10. Follow-Up Tickets

Based on this research, these implementation tickets should be created:

1. **Redesign discovery domain model** — Add DomainStory, StorySentence, BoundedContextSketch, UbiquitousLanguageEntry VOs. Refactor DiscoverySession to hold stories instead of Q/A pairs.

2. **Implement Domain Storytelling moderator flow** — Replace FixedQuestionFlow with story-driven flow. AI proposes stories, user refines sentence by sentence.

3. **Implement boundary detection engine** — Analyze stories for boundary signals (one-way flow, language differences, different triggers). Propose bounded contexts.

4. **Implement AI domain research phase** — DomainResearcher port + web search adapter. Bootstrap domain knowledge when user lacks expertise.

5. **Define and implement .story.yaml format** — Structured story persistence with trust levels, annotations, variations.

6. **Implement glossary extraction** — Build ubiquitous language glossary incrementally from story sentences.

7. **Implement PlantUML DST export** — Optional export using MIT-licensed DomainStory-PlantUML syntax.

8. **Implement Egon .egn JSON export** — Optional export for Egon.io visualization.

9. **Implement gstack CEO plan consumption** — Detect and read ~/.gstack/projects/{slug}/ceo-plans/ as optional input context.

10. **Implement constrained artifact generation** — LLM generates DDD.md/PRD.md/ARCHITECTURE.md using only concepts from domain stories (Junker's technique).

---

## 11. Supporting Research

This report consolidates findings from:

- [20260323_1_domain_storytelling_methodology.md](20260323_1_domain_storytelling_methodology.md) — Foundational DST research (notation, workshop format, DDD mapping, minimum viable stories, AI suitability)
- [20260323_2_domain_storytelling_2025_2026_developments.md](20260323_2_domain_storytelling_2025_2026_developments.md) — 2025-2026 developments (Junker's LLM pipeline, new books, tooling updates, conference activity, competitive landscape)
- gstack skills analysis (plan-ceo-review, plan-eng-review, design-consultation) — Interaction patterns
- alto codebase analysis (discovery bounded context) — Current flow assessment
