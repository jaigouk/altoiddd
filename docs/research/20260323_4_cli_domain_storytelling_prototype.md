---
last_reviewed: 2026-03-23
owner: researcher
status: complete
spike: alty-cli-gox
---

# Spike Report: CLI-based Domain Storytelling Moderator Prototype

**Date:** 2026-03-23
**Spike:** alty-cli-gox
**Risk de-risked:** Can Domain Storytelling work in a CLI terminal without visuals?

---

## Table of Contents

1. [Executive Summary](#1-executive-summary)
2. [Prototype Description](#2-prototype-description)
3. [Decision 1: Primary Interaction Pattern](#3-decision-1-primary-interaction-pattern)
4. [Decision 2: Story Proposal Mechanism](#4-decision-2-story-proposal-mechanism)
5. [Decision 3: Prompter Interface for Consultant Flow](#5-decision-3-prompter-interface-for-consultant-flow)
6. [Correction Flow Testing](#6-correction-flow-testing)
7. [Synthesis Checkpoint Testing](#7-synthesis-checkpoint-testing)
8. [Mode Selection UX](#8-mode-selection-ux)
9. [Key Findings](#9-key-findings)
10. [Follow-Up Tickets](#10-follow-up-tickets)
11. [Sources](#11-sources)

---

## 1. Executive Summary

**Can Domain Storytelling work in a CLI terminal?** Yes. The prototype validates that the moderator conversation works in text-only. The critical decisions:

| Decision | Recommendation | Rationale |
|----------|---------------|-----------|
| **Primary interaction pattern** | C) Hybrid | Consultant proposes first story (requires LLM), user narrates subsequent stories. Best of both worlds. |
| **Story proposal for MVP** | B) User-driven only | Template matching from README fails catastrophically (see Section 4). For MVP, user is the domain knowledge source. LLM proposal is a Phase 2 enhancement. |
| **Prompter interface** | New `StorytellingPrompter` | 8 methods replacing current 4. See Section 5 for signatures. |

**Key finding:** The biggest risk was whether template-matching from README nouns/verbs could propose usable stories without an LLM. It cannot. The parser found 1 out of 5 actors in the e-commerce README, 1 out of 4 in the vet clinic README, and generated grammatically broken sentences ("Veterinarians management Pet"). User-driven narration (Option B) produces high-quality stories with zero external dependencies.

---

## 2. Prototype Description

### Location

- Prototype code: `docs/research/prototype/cli-dst/`
- Sample stories: `docs/research/samples/`
- This report: `docs/research/20260323_4_cli_domain_storytelling_prototype.md`

### What Was Built

A standalone Go program (`cli-dst-prototype`) that simulates the moderator conversation. It tests:

1. **Two interaction patterns:**
   - Consultant-proposes: AI generates a story from README, user refines sentence by sentence
   - User-narrates: AI asks moderator questions, user tells the story, AI structures into sentences
2. **Correction flows:** reject, edit, insert, reorder, branching detection
3. **Synthesis checkpoints:** mid-story (every 3 sentences) and final
4. **Mode selection:** RAPID (3 stories) vs THOROUGH (5+ stories)
5. **Template-matching parser:** regex-based noun/verb extraction from README text

### Sample Stories Created

Three stories in alto text format, reusable by other spikes:

| File | Domain | Sentences | Actors | Work Objects |
|------|--------|-----------|--------|-------------|
| `docs/research/samples/01-ecommerce-customer-purchases-product.story.txt` | E-commerce marketplace | 11 | 5 | 6 |
| `docs/research/samples/02-vet-clinic-pet-examination.story.txt` | Veterinary clinic | 10 | 4 | 6 |
| `docs/research/samples/03-alto-new-project-bootstrap.story.txt` | alto itself | 15 | 3 | 9 |

---

## 3. Decision 1: Primary Interaction Pattern

### Recommendation: C) Hybrid

**First story:** Consultant-proposes (when LLM is available) or user-narrates (MVP/fallback).
**Subsequent stories:** User-narrates.

### Evidence

**Consultant-proposes strengths:**
- Lower activation energy for users who don't know DST methodology
- Follows gstack's proven "propose then refine" pattern (design-consultation Phase 1/3)
- Users spend less time thinking about structure, more time correcting substance
- Trust level tracking shows clear provenance (ai_inferred -> user_confirmed)

**Consultant-proposes weaknesses:**
- Requires either an LLM or a reliable parser to generate the initial proposal
- Template matching fails (see Decision 2) — the initial proposal is only as good as the input parsing
- If the proposal is poor, the user spends more time rejecting/editing than composing from scratch
- Without LLM: empty or nonsensical proposals destroy trust ("Veterinarians management Pet")

**User-narrates strengths:**
- Zero external dependencies — works today, no LLM needed
- The user IS the domain expert; asking them to narrate leverages their knowledge directly
- Moderator questions are well-defined (Hofer & Schwentner's facilitator questions, source: [domainstorytelling.org/quick-start-guide](https://domainstorytelling.org/quick-start-guide))
- Every sentence has `user_stated` trust level from the start
- Feels natural — "tell me a story" is more intuitive than "correct my mistakes"

**User-narrates weaknesses:**
- Higher activation energy — user must think of the answer, not just confirm/reject
- Some users (especially non-technical POs, domain experts) may struggle to start from zero
- The user must provide structure (who/what/order) that a consultant could propose

**Hybrid rationale:**
- First story benefits most from a proposal because it establishes the vocabulary and actors
- Subsequent stories can reference actors/objects already established, making narration easier
- "Here are the actors we've identified: Pet Owner, Receptionist, Veterinarian. For your next story, who starts the process?" — the context is set
- Matches gstack Phase 0 (auto-infer) -> Phase 1 (compound question with pre-fill) -> Phase 4 (user drives drill-downs)

**For MVP:** Use user-narrates as the sole pattern. Add consultant-proposes when LLM integration is available.

### Source

- gstack design-consultation SKILL.md: Phase 0 auto-infer, Phase 1 compound question, Phase 3 complete proposal
- DST workshop format: moderator asks structured questions, domain expert narrates (source: [domainstorytelling.org/quick-start-guide](https://domainstorytelling.org/quick-start-guide))
- Prototype testing: `docs/research/prototype/cli-dst/`

---

## 4. Decision 2: Story Proposal Mechanism for MVP

### Recommendation: B) User-driven only

Template matching (Option A) fails. LLM (Option C) is the correct long-term solution but is not needed for MVP. User-driven narration (Option B) is a valid, complete MVP.

### Template Matching Failure Evidence

Tested the regex-based README parser on all three sample domains:

**E-commerce README:**
```
Input: "An online marketplace where customers browse products from multiple sellers,
        add items to their cart, pay via credit card or PayPal, and receive home
        delivery. Sellers manage their own inventory and pricing. The platform takes
        a commission on each sale."

Actors found: [Sellers]          (missed: customers, platform)
Work Objects found: []           (missed: cart, products, inventory, commission, delivery)
Sentences generated: 0           (not enough data to construct any)
```

**Vet Clinic README:**
```
Input: "A management system for a veterinary clinic. Pet owners book appointments
        by phone or online. The receptionist manages the schedule. Veterinarians
        examine animals, record diagnoses, prescribe treatments. The clinic tracks
        medical history per animal and bills owners after visits."

Actors found: [Veterinarians]    (missed: Pet owners, receptionist, clinic)
Work Objects found: []           (missed: appointments, schedule, diagnoses, treatments, medical history)
Sentences generated:
  1. Veterinarians management Pet  [ai_inferred]  <- NONSENSICAL
  2. Veterinarians book Pet        [ai_inferred]  <- WRONG ACTOR
```

**Alto README:**
```
Actors found: []
Work Objects found: []
Sentences generated: 0
```

**Root cause:** Natural language READMEs use:
- Lowercase domain terms ("customers", "receptionist", "appointments") that regex capitalization patterns miss
- Complex sentence structures where nouns/verbs can't be extracted by word-boundary matching
- Compound terms ("medical history", "home delivery") that single-word extraction fragments
- Context-dependent word roles ("management system" where "management" is an adjective, not a verb)

**Conclusion:** Template matching cannot produce usable story proposals from natural-language README text. An LLM with natural language understanding is required for consultant-proposes to work. This is not surprising — it's the core value proposition of LLMs.

### Why User-Driven is a Valid MVP

The user IS the domain knowledge source. In a traditional DST workshop, the domain expert tells the story and the moderator records it. alto's moderator role is to:

1. Ask structured questions ("Who starts the process?", "What do they do first?", "What happens next?")
2. Structure answers into StorySentence format
3. Number activities automatically
4. Validate understanding by replaying the story
5. Detect boundary signals across stories

None of these require proposing stories. The moderator is a facilitator, not a domain expert.

gstack's CEO review (upstream from alto) handles the "scope challenge" — ensuring the user has thought through what they're building. By the time the user reaches alto, they should know their domain well enough to narrate stories.

### Source

- Parser test: `docs/research/prototype/cli-dst/test_parser_only.go` (run via `go run test_parser_only.go domain.go readme_parser.go`)
- Prototype consultant flow: `docs/research/prototype/cli-dst/consultant_flow.go`
- Prototype narrator flow: `docs/research/prototype/cli-dst/narrator_flow.go`
- DST workshop role of moderator: [domainstorytelling.org/quick-start-guide](https://domainstorytelling.org/quick-start-guide)

---

## 5. Decision 3: Prompter Interface for Consultant Flow

### Current Prompter Interface

```go
// Current: internal/discovery/application/ports.go
type Prompter interface {
    SelectPersona(ctx context.Context) (string, error)
    AskQuestion(ctx context.Context, question string) (string, error)
    AskSkipReason(ctx context.Context) (string, error)
    ConfirmPlayback(ctx context.Context, summary string) (bool, error)
}
```

This interface is designed for a fixed-question questionnaire. It cannot support:
- Proposing a multi-sentence story and getting per-sentence feedback
- Asking moderator follow-up questions dynamically
- Displaying structured story text with numbered sentences
- Offering lettered choices (A/B/C) with descriptions
- Capturing structured sentence components (subject/activity/object)

### Recommended: StorytellingPrompter Interface

```go
// New: replaces Prompter for the discovery redesign
type StorytellingPrompter interface {
    // SelectMode presents RAPID/THOROUGH/EXISTING mode options.
    // Returns the selected DiscoveryMode.
    SelectMode(ctx context.Context) (DiscoveryMode, error)

    // ProposeStory presents a complete proposed story and collects
    // per-sentence feedback. Returns the refined story.
    // Used in consultant-proposes pattern.
    ProposeStory(ctx context.Context, proposed *DomainStory) (*DomainStory, error)

    // AskNarration asks a moderator question and returns the user's
    // free-text response. Used in user-narrates pattern.
    // The question parameter follows the re-ground/simplify/recommend format.
    AskNarration(ctx context.Context, question string, context string) (string, error)

    // ConfirmSentence presents a structured sentence and asks the user
    // to accept, reject, or edit it. Returns the (possibly edited) sentence
    // and whether it was accepted.
    ConfirmSentence(ctx context.Context, sentence StorySentence) (StorySentence, bool, error)

    // AskChoice presents lettered options with descriptions and an
    // opinionated recommendation. Returns the selected option key.
    AskChoice(ctx context.Context, prompt string, options []Choice, recommended string) (string, error)

    // DisplayStory renders a complete story in alto text format for review.
    // This is display-only; no input captured.
    DisplayStory(ctx context.Context, story *DomainStory) error

    // SynthesisCheckpoint presents the accumulated understanding and asks
    // for confirmation. Returns true if confirmed, false if user wants changes.
    // The checkpoint includes stories told so far, actors, work objects, and
    // any boundary signals detected.
    SynthesisCheckpoint(ctx context.Context, synthesis SynthesisSummary) (bool, error)

    // AskAnnotation prompts the user for a business rule or constraint.
    // Returns the annotation text and optional sentence number, or empty
    // string if the user is done adding annotations.
    AskAnnotation(ctx context.Context) (string, int, error)
}

// Supporting types

type Choice struct {
    Key         string // "A", "B", "C"
    Label       string // Short description
    Description string // Longer explanation
}

type SynthesisSummary struct {
    StoriesSoFar    []DomainStory
    ActorInventory  []StoryActor
    ObjectInventory []WorkObject
    BoundarySignals []string
    GlossaryTerms   []string
}
```

### How Methods Map to Interaction Patterns

| Method | Consultant-Proposes | User-Narrates | Both |
|--------|-------------------|---------------|------|
| `SelectMode` | | | X |
| `ProposeStory` | X | | |
| `AskNarration` | | X | |
| `ConfirmSentence` | X | X | |
| `AskChoice` | | | X |
| `DisplayStory` | | | X |
| `SynthesisCheckpoint` | | | X |
| `AskAnnotation` | | | X |

### Migration Path from Current Prompter

| Current Method | New Equivalent | Notes |
|----------------|---------------|-------|
| `SelectPersona` | `AskChoice` (generic) | Persona selection becomes one of many choice interactions |
| `AskQuestion` | `AskNarration` | Free-text input with richer context |
| `AskSkipReason` | Removed | Skipping doesn't apply in storytelling; user says "done" |
| `ConfirmPlayback` | `SynthesisCheckpoint` | Richer payload with structured data |

### Source

- Current Prompter: `internal/discovery/application/ports.go:69-84`
- Current HuhPrompter implementation: `internal/discovery/infrastructure/huh_prompter.go`
- gstack AskUserQuestion format: `.claude/skills/gstack/design-consultation/SKILL.md` (re-ground, simplify, recommend, options)
- Prototype flows: `docs/research/prototype/cli-dst/consultant_flow.go`, `narrator_flow.go`

---

## 6. Correction Flow Testing

### Results

All correction flows tested successfully in the prototype:

| Flow | Mechanism | Result |
|------|-----------|--------|
| **Reject sentence** | `[n]` at confirmation prompt | Sentence removed, remaining renumbered. Works cleanly. |
| **Edit sentence** | `[e]` at confirmation prompt | Component-by-component editing (subject, activity, object, preposition, indirect object). Feels natural — user edits only what's wrong. |
| **Insert sentence** | `[i]` at confirmation prompt | New sentence composed and inserted before current. Renumbering works. |
| **Add missing actor** | Post-refinement prompt | Comma-separated actor names added to story. Trust level set to `user_stated`. |
| **Reorder activities** | Edit + renumber | Renumbering after insert/remove handles reordering implicitly. No explicit "move" needed. |
| **Detect branching** | Keyword scan in activities | Detects "sometimes", "or", "alternatively", "if", "optionally", "either". Suggests splitting into separate story. Works well. |

### Observations

1. **Component-by-component editing is superior to free-text re-entry.** When editing, showing the current value in brackets `[current]` and accepting Enter for "keep" reduces friction. The user only types what changed.

2. **Branching detection is a high-value feature.** Domain experts naturally say "sometimes the customer pays by credit card or PayPal." DST requires separate stories for alternatives. Automatic detection + suggestion to split is a genuine moderator contribution.

3. **Renumbering must be automatic.** After any insert/remove, all sentence numbers must be recalculated. Users should never manually manage sequence numbers.

---

## 7. Synthesis Checkpoint Testing

### Mid-Story Checkpoints (Every 3 Sentences)

Tested in the narrator flow. The checkpoint replays all sentences so far and asks "Does this look right?"

**Finding:** Every-3-sentences feels right for CLI. It prevents drift without being annoying. The gstack pattern of "one decision per interaction" maps to "one sentence confirmation" in the per-sentence loop, and "one checkpoint per phase" maps to the periodic synthesis.

### Final Synthesis

Both flows end with a full story replay and "Does this capture your primary workflow?" prompt.

**Finding:** Phase-based synthesis (replay complete story after all sentences are done) is better than per-question playback. Reasons:
- The story has a narrative arc — reading it end-to-end reveals gaps and ordering issues
- Per-question playback (alto's current every-3-answers approach) fragments the narrative
- The full replay is the DST equivalent of the moderator reading back the diagram — it's the moment the domain expert says "yes, that's how it works" or "no, you missed step X"

### Recommendation

Use a two-tier synthesis:
1. **Per-sentence confirmation:** After each sentence is captured/proposed, confirm individually (lightweight, Y/n/edit)
2. **Phase synthesis:** After the story is complete, replay the whole thing for narrative-level review

---

## 8. Mode Selection UX

### Testing

Mode selection was tested as the first interaction in the prototype:

```
How deep should we go?

  A) RAPID   -- 3 stories, ~15 min
     Enough for MVP project setup. Covers primary workflow,
     main failure case, and one secondary workflow.

  B) THOROUGH -- 5+ stories, ~30 min
     Comprehensive domain model. Adds fine-grained stories for
     core subdomains and explicit business rules extraction.
```

### Finding

Mode selection as first interaction works well. It:
- Sets expectations immediately (time commitment, depth)
- Maps to the gstack pattern of progressive phase activation (Phase 2/3 only if THOROUGH)
- Replaces the current EXPRESS/DEEP/CONVERSATIONAL flow strategy with something more meaningful
- The descriptions are actionable — user knows what each mode produces

### Recommendation

Keep mode selection as the first interaction. Add EXISTING mode when rescue mode is implemented:

```
  C) EXISTING -- Analyze codebase, ~20 min
     Reverse-engineer domain model from existing code.
     For projects that need structure applied after the fact.
```

---

## 9. Key Findings

### Finding 1: CLI Domain Storytelling Works

The sentence-based format ("Actor does Activity using Work Object") is inherently text-native. No visual diagram is needed for the core storytelling flow. The moderator role (ask structured questions, record sentences, validate understanding) maps directly to a CLI conversation.

**Source:** Prototype testing on vet clinic domain, both interaction patterns.

### Finding 2: Template Matching Cannot Replace LLM for Story Proposal

Regex-based noun/verb extraction from README text produces unusable output. The parser found 20% of actors and 0% of work objects across all test domains. Sentences generated were grammatically broken and semantically nonsensical.

**Source:** `docs/research/prototype/cli-dst/test_parser_only.go`, run on all three sample domains.

### Finding 3: User-Narrates is the Natural MVP

User-driven narration with moderator questions requires zero external dependencies and produces stories with 100% `user_stated` trust level. The moderator's structured questions (from Hofer & Schwentner's workshop format) guide the user through storytelling naturally.

**Source:** Prototype narrator flow tested with vet clinic domain.

### Finding 4: Per-Sentence Confirmation + Phase Synthesis is the Right Checkpoint Model

Two-tier synthesis: lightweight Y/n/edit after each sentence, full story replay after each complete story. This replaces alto's current every-3-answers playback with a model that respects the narrative structure.

**Source:** Prototype testing of both synthesis approaches.

### Finding 5: Branching Detection is High-Value

Automatic detection of branching language ("sometimes", "or", "alternatively") with a suggestion to split into separate stories is a genuine moderator contribution that doesn't require an LLM. This enforces DST's "one scenario per story" rule automatically.

**Source:** Prototype branching detection tested in both flows.

### Finding 6: The StorytellingPrompter Needs 8 Methods, Not 4

The current Prompter interface (4 methods: SelectPersona, AskQuestion, AskSkipReason, ConfirmPlayback) cannot support consultant-proposes or structured sentence editing. The new StorytellingPrompter needs: SelectMode, ProposeStory, AskNarration, ConfirmSentence, AskChoice, DisplayStory, SynthesisCheckpoint, AskAnnotation.

**Source:** Interface design derived from prototype flows (see Section 5).

---

## 10. Follow-Up Tickets

Based on this spike, these implementation tickets are recommended:

1. **Implement StorytellingPrompter port interface** -- Define the 8-method interface in `internal/discovery/application/ports.go`. Create HuhStorytellingPrompter adapter using charmbracelet/huh v2.

2. **Implement user-narrates moderator flow** -- Build the narrator flow as the MVP discovery pattern. Uses StorytellingPrompter methods: AskNarration, ConfirmSentence, AskChoice, DisplayStory, SynthesisCheckpoint, AskAnnotation.

3. **Implement DomainStory and StorySentence value objects** -- Add to `internal/discovery/domain/`. Include FormatText(), validation, trust level tracking, branching detection.

4. **Implement consultant-proposes flow (requires LLM)** -- Add ProposeStory method implementation. Requires LLM port for story generation from README. Phase 2 work, not MVP.

5. **Implement branching detection in domain layer** -- Keyword-based detection of branching language in story sentences. Suggests splitting into variation stories. Pure domain logic, no external deps.

6. **Implement mode selection (RAPID/THOROUGH)** -- Replace EXPRESS/DEEP/CONVERSATIONAL flow strategy with RAPID/THOROUGH/EXISTING modes. Controls story count and phase activation.

---

## 11. Sources

### Primary

- **Prototype code:** `docs/research/prototype/cli-dst/` (main.go, domain.go, consultant_flow.go, narrator_flow.go, readme_parser.go)
- **Sample stories:** `docs/research/samples/` (3 stories in alto text format)
- **Parser test:** `docs/research/prototype/cli-dst/test_parser_only.go`

### Prior Research (Inputs to This Spike)

- `docs/research/20260323_3_gstack_ux_and_domain_storytelling.md` -- Section 2 (gstack patterns), Section 7 (proposed flow)
- `docs/research/20260323_1_domain_storytelling_methodology.md` -- Section 1 (notation), Section 2 (workshop format), Section 7 (AI suitability)
- `.claude/skills/gstack/design-consultation/SKILL.md` -- Consultant interaction pattern

### Existing Code (Baseline)

- `internal/discovery/application/ports.go:69-84` -- Current Prompter interface
- `internal/discovery/infrastructure/huh_prompter.go` -- Current HuhPrompter implementation
- `internal/discovery/infrastructure/cli_discovery_adapter.go` -- Current CLI discovery flow

### Domain Storytelling Methodology

- [domainstorytelling.org/quick-start-guide](https://domainstorytelling.org/quick-start-guide) -- Moderator questions, workshop format
- [InfoQ Podcast -- Domain Storytelling](https://www.infoq.com/podcasts/domain-storytelling/) -- "One scenario per story" rule
- [Tech Lead Journal #75](https://techleadjournal.dev/episodes/75/) -- "Maybe two, three examples are enough"
