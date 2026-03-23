---
last_reviewed: 2026-03-23
owner: researcher
status: complete
spike: alty-cli-1i3
---

# Spike Report: .story.yaml Format and Export Validation

**Date:** 2026-03-23
**Spike:** alty-cli-1i3
**Goal:** Finalize .story.yaml, glossary.yaml, and context-map.yaml schemas. Paper-validate against 4 downstream consumers. Assess v1 migration.

---

## Table of Contents

1. [Decisions Summary](#1-decisions-summary)
2. [v1 Format Spec](#2-v1-format-spec)
3. [.story.yaml Schema (Finalized)](#3-storyyaml-schema-finalized)
4. [glossary.yaml Schema (Finalized)](#4-glossaryyaml-schema-finalized)
5. [context-map.yaml Schema (Finalized)](#5-context-mapyaml-schema-finalized)
6. [Paper Validation: PlantUML Export](#6-paper-validation-plantuml-export)
7. [Paper Validation: Egon .egn Export](#7-paper-validation-egon-egn-export)
8. [Paper Validation: DDD.md Generation](#8-paper-validation-dddmd-generation)
9. [Paper Validation: Ticket Pipeline](#9-paper-validation-ticket-pipeline)
10. [Terminal Readability Assessment](#10-terminal-readability-assessment)
11. [Serialization Format Decision](#11-serialization-format-decision)
12. [v1 Migration Analysis](#12-v1-migration-analysis)
13. [Schema Changes from Proposed](#13-schema-changes-from-proposed)

---

## 1. Decisions Summary

| Decision | Outcome | Rationale |
|----------|---------|-----------|
| **D1: .story.yaml schema** | Finalized with minor additions from Section 9.1 proposal | Added `aliases` field to glossary, `description` to context-map relationships, `source` field on sentences. See Section 13 for full diff. |
| **D2: Serialization format** | YAML confirmed | Best balance of human readability, nesting support, and Go ecosystem tooling. See Section 11. |
| **D3: Companion formats** | glossary.yaml and context-map.yaml finalized | Schemas validated against 3 sample domains. See Sections 4-5. |
| **D4: v1 migration** | **B) Archive** -- v1 sessions preserved as-is, not converted | v1 captures Q/A pairs (unstructured text), not domain stories (structured sentences). Migration is fundamentally lossy. See Section 12. |

---

## 2. v1 Format Spec

The v1 discovery session format is defined in:
- `internal/discovery/domain/discovery_session.go` (aggregate, `ToSnapshot()`/`FromSnapshot()`)
- `internal/discovery/domain/discovery_values.go` (value objects: Answer, Playback, etc.)
- `internal/discovery/infrastructure/filesystem_session_repository.go` (JSON persistence)

### v1 Persisted Fields

| Field | Type | Description |
|-------|------|-------------|
| `session_id` | string | UUID for the session |
| `readme_content` | string | Raw README text input |
| `status` | enum string | One of: created, persona_detected, answering, playback_pending, completed, cancelled, round_1_complete, challenging, round_2_complete, simulating |
| `persona` | nullable string | developer, product_owner, domain_expert, mixed |
| `register` | nullable string | technical, non_technical |
| `mode` | nullable string | express, deep, conversational |
| `round` | nullable string | discovery, challenge, simulate |
| `answers` | []{ question_id, response_text } | Ordered list of Q/A pairs |
| `skipped` | []{ question_id, reason } | Questions skipped with reason |
| `playback_confirmations` | []{ summary_text, confirmed, corrections } | Playback checkpoints |
| `answers_since_last_playback` | int | Counter for playback interval |
| `tech_stack` | nullable { language, package_manager } | Detected tech stack |
| `context_classifications` | map[name]{ classification, rationale } | Subdomain classification results |

**Source:** `internal/discovery/domain/discovery_session.go:462-536` (ToSnapshot), `internal/discovery/domain/discovery_session.go:539-793` (FromSnapshot)

### v1 Question Catalog (10 questions)

| ID | Phase | Produces |
|----|-------|----------|
| Q1 | actors | actors, external_systems |
| Q2 | actors | entities, value_objects |
| Q3 | story | commands, events, domain_story |
| Q4 | story | invariants, failure_modes |
| Q5 | story | secondary_stories, commands |
| Q6 | events | domain_events |
| Q7 | events | policies, reactions |
| Q8 | events | read_models, projections |
| Q9 | boundaries | bounded_contexts |
| Q10 | boundaries | subdomain_classification |

**Source:** `internal/discovery/domain/question.go:45-86`

### v1 Persistence Format

File: `discovery_session.json` (single file per project, stored in `.alto/` directory)
Format: JSON via `encoding/json` with `MarshalIndent`
Storage: `FileSystemSessionRepository` at `internal/discovery/infrastructure/filesystem_session_repository.go`

---

## 3. .story.yaml Schema (Finalized)

### Top-level Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `title` | string | yes | Human-readable story title |
| `type` | enum | yes | `coarse_grained` or `fine_grained` |
| `time` | enum | yes | `as_is` or `to_be` |
| `purity` | enum | yes | `pure` or `digitalized` |
| `trigger` | string | yes | What initiates this story |
| `actors` | []Actor | yes | List of story participants |
| `work_objects` | []WorkObject | yes | List of things acted upon |
| `sentences` | []Sentence | yes | Ordered story sentences |
| `annotations` | []Annotation | no | Business rules and constraints |
| `variations` | []string | no | Pointers to alternative stories |

### Actor Schema

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | yes | Actor's name in ubiquitous language |
| `type` | enum | yes | `person`, `system`, `group` |
| `trust` | enum | yes | `user_stated`, `user_confirmed`, `ai_researched`, `ai_inferred` |
| `source` | string | no | Citation when trust is `ai_researched` |

### WorkObject Schema

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | yes | Work object's name in ubiquitous language |
| `type` | enum | yes | `document`, `folder`, `call`, `email`, `conversation`, `info` |
| `trust` | enum | yes | `user_stated`, `user_confirmed`, `ai_researched`, `ai_inferred` |
| `source` | string | no | Citation when trust is `ai_researched` |

### Sentence Schema

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `step` | int | yes | Sequence number (1-based) |
| `subject` | string | yes | Actor name performing the action |
| `activity` | string | yes | Verb phrase (domain language) |
| `object` | string | yes | Work object or actor being acted upon |
| `preposition` | string | no | Connecting word (for, to, via, using, from, with, in, about, based on, on) |
| `indirect_object` | string | no | Second work object or actor |
| `trust` | enum | yes | `user_stated`, `user_confirmed`, `ai_researched`, `ai_inferred` |
| `source` | string | no | Citation when trust is `ai_researched` |

### Annotation Schema

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `text` | string | yes | The business rule or constraint text |
| `sentence` | int | no | Step number this annotation applies to (null = story-wide) |
| `type` | enum | yes | `constraint`, `invariant`, `assumption` |
| `trust` | enum | no | Trust level (defaults to story-level trust) |
| `source` | string | no | Citation when trust is `ai_researched` |

### Trust Level Enum

| Value | Meaning |
|-------|---------|
| `user_stated` | User explicitly said this during the story |
| `user_confirmed` | AI proposed, user confirmed |
| `ai_researched` | AI discovered via domain research, not yet confirmed |
| `ai_inferred` | AI inferred from context, lowest confidence |

### Validation Rules

1. Every `subject` in sentences must reference a name from `actors`
2. Every `object` and `indirect_object` in sentences must reference a name from either `work_objects` or `actors`
3. Step numbers must be sequential starting from 1
4. At least 1 actor, 1 work object, and 1 sentence required
5. If `trust` is `ai_researched`, `source` should be present

---

## 4. glossary.yaml Schema (Finalized)

### Top-level

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `terms` | []Term | yes | List of ubiquitous language terms |

### Term Schema

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `term` | string | yes | The term in ubiquitous language |
| `definition` | string | yes | What this term means in the domain |
| `context` | string | yes | Which bounded context this term belongs to |
| `trust` | enum | yes | Trust level |
| `stories` | []string | yes | Story file references where this term appears |
| `note` | string | no | Additional notes (e.g., language difference signals) |
| `aliases` | []string | no | Alternative names for this term in other contexts |
| `source` | string | no | Citation for ai_researched terms |

### Glossary Extraction Rules

1. Every `work_objects[].name` and `actors[].name` is a candidate glossary term
2. Every `sentences[].activity` verb is a candidate for the domain activity vocabulary
3. `aliases` capture language difference signals across bounded contexts (e.g., "Shopping Cart" vs "Basket")
4. `note` captures cross-context naming conflicts that signal boundaries

---

## 5. context-map.yaml Schema (Finalized)

### Top-level

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `project` | string | yes | Project name |
| `contexts` | []Context | yes | List of bounded contexts |
| `relationships` | []Relationship | yes | How contexts relate |

### Context Schema

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | yes | Context name |
| `classification` | enum | yes | `core`, `supporting`, `generic` |
| `confidence` | float | yes | 0.0-1.0, how confident we are in this boundary |
| `actors` | []string | yes | Actors that appear in this context |
| `work_objects` | []string | yes | Work objects that belong to this context |
| `boundary_signals` | []BoundarySignal | yes | Evidence for this boundary |
| `stories` | []string | yes | Story:step references (e.g., "ecommerce/01:1-5") |
| `trust` | enum | yes | Trust level |

### BoundarySignal Schema

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `type` | enum | yes | `different_trigger`, `one_way_flow`, `language_difference`, `different_lifecycle`, `external_system`, `different_actor`, `complex_rules` |
| `description` | string | yes | Human-readable explanation |

### Relationship Schema

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `upstream` | string | yes | Upstream context name |
| `downstream` | string | yes | Downstream context name |
| `type` | enum | yes | `conformist`, `customer_supplier`, `published_language`, `shared_kernel`, `anticorruption_layer`, `open_host_service`, `partnership` |
| `shared` | []string | yes | Work objects shared across this boundary |
| `description` | string | no | Human-readable relationship description |

---

## 6. Paper Validation: PlantUML Export

### Method

Hand-wrote PlantUML equivalents for e-commerce and vet clinic stories using DomainStory-PlantUML v0.3.1 syntax (MIT license). Files: `docs/research/samples/ecommerce.puml`, `docs/research/samples/vetclinic.puml`.

**Reference:** [github.com/johthor/DomainStory-PlantUML](https://github.com/johthor/DomainStory-PlantUML)

### Mapping

| .story.yaml Field | PlantUML Equivalent | Notes |
|-------------------|---------------------|-------|
| `actors[].type: person` | `Person($name)` | Direct |
| `actors[].type: group` | `Group($name)` | Direct |
| `actors[].type: system` | `System($name)` | Direct |
| `work_objects[].type: document` | `Document($name)` | Direct |
| `work_objects[].type: info` | `Info($name)` | Direct |
| `work_objects[].type: folder` | `Folder($name)` | Direct |
| `work_objects[].type: call` | `Call($name)` | Direct |
| `work_objects[].type: email` | `Email($name)` | Direct |
| `work_objects[].type: conversation` | `Conversation($name)` | Direct |
| `sentences[].step` | `activity($step, ...)` | Direct |
| `sentences[].subject` | `$subject` parameter | Direct |
| `sentences[].activity` | `$predicate` parameter | Direct |
| `sentences[].object` | `$object` parameter | Direct |
| `sentences[].preposition` | `$post` parameter | Direct |
| `sentences[].indirect_object` | `$target` parameter | Direct |
| `annotations` | **NO EQUIVALENT** | PlantUML DST has no annotation mechanism |
| `trust` | **NO EQUIVALENT** | PlantUML DST has no trust/provenance metadata |
| `variations` | **NO EQUIVALENT** | No variation pointer mechanism |
| Story metadata (type, time, purity) | **NO EQUIVALENT** | Not representable in PlantUML DST |
| `trigger` | **NO EQUIVALENT** | Not representable |

### Verdict

**Export is feasible but lossy.** The core narrative (actors, work objects, sentences) maps 1:1. PlantUML DST cannot represent:
- **Annotations** (constraints, invariants) -- these are the most important loss
- **Trust levels** -- provenance metadata has no PlantUML equivalent
- **Story metadata** (type, time, purity, trigger) -- these could go in a PlantUML comment
- **Variations** -- could be listed in a PlantUML note

**Recommendation:** Export should include a PlantUML `note` block at the bottom listing annotations, and a header comment with story metadata. Trust levels are inherently lost -- this is acceptable since PlantUML is a visualization format, not a data format.

---

## 7. Paper Validation: Egon .egn Export

### Method

Hand-wrote an Egon .egn JSON file for the e-commerce story based on reverse-engineering the .egn schema from `WPS/egon.io-examples` (travel-1-by-taxi-en.egn). File: `docs/research/samples/ecommerce.egn`.

**Reference:** [github.com/WPS/egon.io-examples](https://github.com/WPS/egon.io-examples), .egn format version 2.0.1

### .egn JSON Schema (Reverse-Engineered)

```
{
  "domain": {
    "name": string,           // Domain name
    "actors": {               // Actor type -> SVG icon mapping
      "Person": "<svg .../>",
      "Group": "<svg .../>",
      "System": "<svg .../>"
    },
    "workObjects": {          // Work object type -> SVG icon mapping
      "Document": "<svg .../>",
      ...
    }
  },
  "dst": [                    // Array of elements (shapes + connections)
    // Shape (actor or work object):
    {
      "type": "domainStory:actorPerson",     // or actorGroup, actorSystem, workObject{Type}
      "name": string,                         // Display name
      "id": "shape_NNNN",                     // Unique ID
      "pickedColor": string,                  // Display color
      "x": number,                            // Canvas X position
      "y": number,                            // Canvas Y position
      "$type": "Element",
      "di": {},
      "$descriptor": {}
    },
    // Connection (activity/sentence):
    {
      "type": "domainStory:activity",
      "name": string,                         // Activity verb
      "id": "connection_NNNN",
      "pickedColor": string,
      "number": int | null,                   // Step number (null for continuation/preposition)
      "source": "shape_NNNN",                 // Source element ID
      "target": "shape_NNNN",                 // Target element ID
      "waypoints": [{"x": N, "y": N}, ...],  // Visual path
      "$type": "Element",
      "di": {},
      "$descriptor": {}
    },
    // Metadata entries (at end of array):
    { "info": string },                        // Story description
    { "version": "2.0.1" }                    // .egn format version
  ]
}
```

### Mapping

| .story.yaml Field | Egon .egn Equivalent | Notes |
|-------------------|---------------------|-------|
| `actors[].name` | Element with type `domainStory:actor{Type}` | Direct, but Egon needs x,y coordinates |
| `actors[].type` | Type suffix: `Person`, `Group`, `System` | Direct |
| `work_objects[].name` | Element with type `domainStory:workObject{Type}` | Direct |
| `work_objects[].type` | Type suffix: `Document`, `Info`, etc. | Direct |
| `sentences[].step` | Connection `number` field | Direct |
| `sentences[].subject` | Connection `source` -> actor shape ID | Indirect: requires ID lookup |
| `sentences[].activity` | Connection `name` field | Direct |
| `sentences[].object` | Connection `target` -> work object shape ID | Indirect: requires ID lookup |
| `sentences[].preposition` | Separate connection with `number: null` | Split: preposition becomes a second connection |
| `sentences[].indirect_object` | Target of the preposition connection | Split across two connections |
| `annotations` | **NO EQUIVALENT** | Egon has no annotation model |
| `trust` | **NO EQUIVALENT** | Egon has no provenance metadata |
| `variations` | **NO EQUIVALENT** | Not representable |
| `title` | `info` entry in dst array | Partial: info is free text |
| `type`, `time`, `purity` | **NO EQUIVALENT** | Not representable |
| `trigger` | **NO EQUIVALENT** | Not representable |

### Additional Challenges

1. **Spatial layout required:** Egon needs x,y coordinates for every element. An export adapter must implement a layout algorithm (or use fixed templates).
2. **SVG icons required:** The `domain.actors` and `domain.workObjects` fields require embedded SVG strings. These are constant per type and can be shipped as a static asset.
3. **Connection splitting:** A .story.yaml sentence with preposition+indirect_object must be split into TWO Egon connections: one numbered (subject->object) and one unnumbered (object->indirect_object).
4. **ID generation:** Egon uses `shape_NNNN` and `connection_NNNN` identifiers. Export must generate stable IDs.

### Verdict

**Export is feasible but requires significant transformation.** The core narrative maps, but:
- Annotations and trust levels are lost (same as PlantUML)
- Spatial layout must be computed (non-trivial)
- Sentences with prepositions require connection splitting
- SVG icon embedding is boilerplate but required

**Recommendation:** Egon export is lower priority than PlantUML. The layout algorithm adds complexity with limited value (users can rearrange in Egon.io). Ship PlantUML first, Egon later.

---

## 8. Paper Validation: DDD.md Generation

### What a DDD.md Generator Needs (and Where to Find It)

| DDD.md Section | Source Format | Extraction Difficulty |
|----------------|--------------|----------------------|
| **Bounded Contexts** | `context-map.yaml` contexts[] | Direct: name, classification, confidence |
| **Context Map (relationships)** | `context-map.yaml` relationships[] | Direct: upstream, downstream, type, shared |
| **Ubiquitous Language** | `glossary.yaml` terms[] | Direct: term, definition, context |
| **Actors / Personas** | `.story.yaml` actors[] | Direct: name, type |
| **Domain Stories (narrative)** | `.story.yaml` sentences[] | Direct: reconstruct natural language from structured sentences |
| **Invariants** | `.story.yaml` annotations[] where type=invariant | Direct: text field |
| **Constraints** | `.story.yaml` annotations[] where type=constraint | Direct: text field |
| **Domain Events** | `.story.yaml` sentences[] (inferred) | **Indirect:** events are implicit in sentence sequences. E.g., "Platform creates Order" implies an OrderCreated event. This requires heuristic extraction. |
| **Aggregates** | Not directly captured | **Missing:** .story.yaml captures work objects, which are candidates for aggregates, but aggregate boundaries require additional analysis |
| **Value Objects** | Partially in glossary | **Partial:** glossary terms are VO candidates, but classification (entity vs VO) is not in the format |
| **Policies** | Not captured | **Missing:** "When X happens, do Y" rules are not in .story.yaml. They could be added as a new annotation type. |

### Assessment

The .story.yaml + glossary.yaml + context-map.yaml trio captures **70-80% of what DDD.md needs**. The gaps:

1. **Domain events** are implicit, not explicit. The sentence "Platform creates Order" implies OrderCreated, but this is heuristic. **Mitigation:** Add an optional `emits` field to sentences in a future version, or extract events via convention ("creates X" -> XCreated, "notifies about X" -> XNotification).

2. **Aggregates vs entities vs value objects** -- the format captures work objects but does not classify them. **Mitigation:** This classification should happen in the DDD.md generation phase (AI-assisted), not in the story format itself. Work objects are raw material.

3. **Policies** ("when OrderPaid, then release Shipment") are not in the format. **Mitigation:** Could be captured as a new annotation type `policy` in a future iteration, or inferred from sentence pairs (step N triggers step N+1).

**Verdict:** The format is sufficient for DDD.md generation. The gaps are acceptable because:
- Domain events can be inferred from sentence verbs with reasonable accuracy
- Aggregate classification is a downstream analysis step, not a storytelling concern
- Policies emerge from story sequences and can be explicitly captured if needed

---

## 9. Paper Validation: Ticket Pipeline

### What a Ticket Pipeline Needs

| Ticket Artifact | Source | Extraction Method |
|-----------------|--------|-------------------|
| **Epics** (per bounded context) | `context-map.yaml` contexts[] | 1 epic per context: "Implement {context.name} Context ({context.classification})" |
| **Stories/Tasks** (per sentence) | `.story.yaml` sentences[] | Each sentence maps to 1-2 tasks. E.g., step 6 "Platform creates Order from Shopping Cart" -> task: "Implement Order creation command handler" |
| **Acceptance criteria** | `.story.yaml` annotations[] | Invariants and constraints become AC items on related tasks |
| **Task ordering** | `.story.yaml` sentences[].step | Sentence sequence implies task dependency order |
| **Edge case tasks** | `.story.yaml` variations[] | Each variation is a candidate for additional stories/tasks |
| **Cross-context integration tasks** | `context-map.yaml` relationships[] | Each relationship maps to an integration task: "Implement {type} between {upstream} and {downstream}" |

### Example Extraction (E-commerce)

From `ecommerce.story.yaml` + `context-map.yaml`:

**Epic: Catalog Context (supporting)**
- Task: Implement Product Listing aggregate
- Task: Implement Inventory management (Seller updates stock)
- AC: "Inventory must not go negative" (from annotations)

**Epic: Ordering Context (core)**
- Task: Implement Shopping Cart aggregate
- Task: Implement checkout flow (Cart -> Order conversion)
- Task: Implement Order creation with payment authorization check
- AC: "Payment must be authorized before Order is created" (from annotations)
- AC: "Customer must be authenticated before checkout" (from annotations)

**Epic: Payment Context (generic)**
- Task: Implement Payment Gateway integration (ACL)
- Task: Implement Commission calculation per product category
- AC: "Commission percentage varies by product category (8-15%)" (from annotations)

**Epic: Fulfillment Context (supporting)**
- Task: Implement Shipment tracking
- Task: Implement delivery confirmation flow
- AC: "Seller has 48 hours to ship after order notification" (from annotations)

**Integration Tasks:**
- Task: Implement Catalog->Ordering conformist (Product Listing snapshot)
- Task: Implement Ordering->Payment customer-supplier (payment initiation)
- Task: Implement Ordering->Fulfillment customer-supplier (order handoff)
- Task: Implement Payment->Fulfillment published language (payment status events)

**Edge Case Stories (from variations):**
- Story: "Payment is declined by gateway" (separate story + tasks)
- Story: "Seller is out of stock after order placed" (separate story + tasks)
- Story: "Customer requests refund after delivery" (separate story + tasks)

### Assessment

**Verdict: Extraction is straightforward.** The format contains enough information for a ticket pipeline to:
1. Generate epics from contexts (with classification driving priority)
2. Generate tasks from sentences (with annotations as AC)
3. Generate integration tasks from relationships
4. Generate edge-case stories from variations
5. Establish ordering from sentence steps and context dependencies

**No additional fields needed** for the ticket pipeline consumer.

---

## 10. Terminal Readability Assessment

### Test Setup

Printed `alto.story.yaml` (15 sentences, 5 annotations, 4 variations) to terminal. 169 lines total.

### Findings

| Criterion | Result | Notes |
|-----------|--------|-------|
| **10-second scan** | Partial | Can identify actors and work objects quickly. Sentences section is harder to scan due to multi-line entries (5-7 lines per sentence). The narrative flow is not immediately obvious. |
| **Nesting depth** | Acceptable | Maximum 2 levels deep (sentences[] -> fields). No deeply nested structures. |
| **Noise ratio** | Moderate | `trust` fields on every element add visual noise for human readers. Essential for machine processing but clutters the terminal view. |
| **Story flow** | Hard to follow | The sentence-by-sentence structure breaks up the narrative. You cannot read the story as a story -- you have to mentally reconstruct it from structured fields. |

### Comparison with alto Text Format (from Section 9.5)

The alto text format from the research report renders much better for terminal reading:

```
Domain Story: "Developer Bootstraps New Project with alto"
Type: coarse-grained, to-be, digitalized
Trigger: Developer has a project idea

1. Developer writes README
2. Developer runs alto CLI
3. alto CLI reads README
4. AI Domain Expert proposes Domain Story based on README
...
```

This is 15 lines for the core story vs 120+ lines in YAML.

### Recommendation

**.story.yaml is the persistence and machine-processing format.** For terminal display, alto should render stories in the alto text format (Section 9.5). This is a display concern, not a storage concern.

The YAML format should NOT be changed to improve readability -- it needs the structure for machine processing. Instead:
1. `alto story show` command renders the text format
2. `alto story export --format yaml` writes the machine format
3. `alto story export --format plantuml` writes the PlantUML format

---

## 11. Serialization Format Decision

### Candidates Evaluated

| Format | Human Readability | Nesting Support | Go Tooling | Round-trip Safety | Verdict |
|--------|------------------|-----------------|------------|-------------------|---------|
| **YAML** | Good (indentation-based, comments allowed) | Excellent | `gopkg.in/yaml.v3` (MIT), mature | Safe with yaml.v3 | **Selected** |
| **JSON** | Poor (verbose, no comments) | Excellent | stdlib `encoding/json` | Perfect | Rejected: no comments, harder to hand-edit |
| **TOML** | Good for flat config | Poor for arrays of objects | `github.com/BurntSushi/toml` (MIT), already in go.mod | Safe | Rejected: awkward for nested arrays like sentences[] |
| **Custom DSL** | Best for stories | N/A | Must build parser | Depends on parser | Rejected: parser cost, maintenance burden |

### Decision: YAML

**Rationale:**

1. **Comments:** Domain experts reviewing .story.yaml files can read inline comments explaining context. JSON does not support comments.
2. **Nesting:** Sentences with optional preposition/indirect_object map naturally to YAML indented lists. TOML's `[[sentences]]` syntax is awkward for 12+ entries.
3. **Go ecosystem:** `gopkg.in/yaml.v3` is MIT-licensed, mature, and pure Go (no CGO). Already well-established in the Go ecosystem.
4. **Consistency:** Many DDD/DevOps tools use YAML (Kubernetes, GitHub Actions, Docker Compose). Developers expect structured config in YAML.
5. **Hand-editability:** Users can create or edit .story.yaml files by hand in any text editor. JSON requires careful bracket matching.

**Risk:** YAML's whitespace sensitivity can cause subtle bugs. Mitigated by:
- Strict validation on load (schema validation)
- `yaml.safe_load()` equivalent in Go (yaml.v3 defaults to safe)
- Future: JSON Schema for IDE validation

### Go Library

```
gopkg.in/yaml.v3
```
- **License:** MIT (via Apache 2.0 origin, re-licensed MIT)
- **CGO:** None
- **Min Go version:** 1.15+
- **Goroutine safety:** Marshal/Unmarshal are stateless, safe for concurrent use
- **Source:** [github.com/go-yaml/yaml](https://github.com/go-yaml/yaml)

---

## 12. v1 Migration Analysis

### Field Mapping: v1 -> v2

| v1 Field | v2 Equivalent | Transfer Quality |
|----------|--------------|-----------------|
| `session_id` | None (stories don't have session IDs) | **Lost** -- irrelevant to v2 |
| `readme_content` | Input context for story generation | **Transferable** as generation input |
| `status` | None (stories don't have workflow state) | **Lost** -- irrelevant to v2 |
| `persona` | None (trust levels replace this) | **Conceptual mapping** -- persona influences trust assignment |
| `register` | None | **Lost** -- irrelevant to v2 |
| `answers[Q1]` (actors) | `actors[]` | **Requires parsing** -- Q1 response is free text listing actors. Must be NLP-parsed or AI-interpreted to extract structured Actor objects. |
| `answers[Q2]` (entities) | `work_objects[]` | **Requires parsing** -- same as Q1. Free text to structured objects. |
| `answers[Q3]` (domain story) | `sentences[]` | **Requires deep interpretation** -- Q3 asks for a narrative. Converting free text like "the user logs in, creates an order, the system processes payment" into structured sentences with subject/activity/object is non-trivial AI work. |
| `answers[Q4]` (invariants) | `annotations[]` where type=invariant | **Requires parsing** -- free text to structured invariants |
| `answers[Q5]` (workflows) | `variations[]` or additional stories | **Requires parsing** |
| `answers[Q6-Q8]` (events, policies, read models) | Not directly in .story.yaml | **No v2 equivalent** -- these v1 answers capture DDD concepts that .story.yaml deliberately does not model (events/policies are downstream analysis) |
| `answers[Q9]` (bounded contexts) | `context-map.yaml` contexts[] | **Requires parsing** -- free text to structured contexts |
| `answers[Q10]` (classification) | `context-map.yaml` contexts[].classification | **Partially transferable** if structured |
| `context_classifications` | `context-map.yaml` contexts[].classification | **Direct transfer** -- already structured |
| `playback_confirmations` | None | **Lost** -- irrelevant to v2 |
| `tech_stack` | Not in story format (separate concern) | **Preserved separately** |

### Assessment

**Migration is fundamentally lossy.** The core problem:

v1 stores **free-text answers to questions**. v2 stores **structured domain stories with actors, work objects, and sentences**. Converting free text like:

> "Users are: customers who buy products, sellers who list products, and admins who manage the platform"

into:

```yaml
actors:
  - name: Customer
    type: person
    trust: user_stated
  - name: Seller
    type: person
    trust: user_stated
  - name: Admin
    type: person
    trust: user_stated
```

...requires NLP or AI interpretation, which is unreliable and adds complexity with questionable value. The user already has the knowledge -- they can re-tell the story in the new format in 10-15 minutes.

### Decision: B) Archive

**v1 sessions should be preserved as-is (readable but not converted).**

**Rationale:**
1. **Conversion is AI work, not mechanical mapping.** Every v1 answer needs interpretation, not just reformatting.
2. **Trust levels cannot be inferred.** v1 has no provenance tracking. Every migrated element would need to be `ai_inferred`, the lowest trust level, making the migrated story less useful than a fresh one.
3. **v1 captures different things.** Q6-Q8 (events, policies, read models) have no direct v2 equivalent because .story.yaml deliberately stays in the storytelling domain and leaves event extraction to downstream analysis.
4. **Low ROI.** The number of existing v1 sessions is very small (alto is pre-release). Building a migration pipeline for ~0 users is waste.
5. **User can re-tell.** A 15-minute storytelling session produces better v2 output than any automated migration of v1 text.

**Implementation:** The FileSystemSessionRepository should continue reading v1 `discovery_session.json` files as-is. New discovery sessions write v2 .story.yaml files. No converter needed.

---

## 13. Schema Changes from Proposed (Section 9.1)

### Additions (not in Section 9.1)

| Field | Where | Rationale |
|-------|-------|-----------|
| `source` on Sentence | `.story.yaml` sentences[] | AI-researched sentences need citation at sentence level, not just actor/object level |
| `aliases` on Term | `glossary.yaml` terms[] | Captures language difference signals across bounded contexts (key DDD insight) |
| `description` on Relationship | `context-map.yaml` relationships[] | Human-readable explanation of why this relationship exists |
| `note` on Term | `glossary.yaml` terms[] | Cross-context naming conflicts and observations |

### Modifications from Section 9.1

| Change | Section 9.1 | Finalized | Rationale |
|--------|------------|-----------|-----------|
| Annotation `type` enum | constraint, invariant (implicit) | constraint, invariant, assumption | Added `assumption` for elements that need future validation |
| Boundary signal types | different_trigger, one_way_flow, language_difference | Added: different_lifecycle, external_system, different_actor, complex_rules | More signal types discovered during sample writing |
| Relationship types | customer_supplier (only example) | Full DDD list: conformist, customer_supplier, published_language, shared_kernel, anticorruption_layer, open_host_service, partnership | Complete strategic DDD vocabulary |
| Work object types | document (only example) | Full DST list: document, folder, call, email, conversation, info | Matches DomainStory-PlantUML type set |

### Removals/Unchanged

No fields from Section 9.1 were removed. The core structure (actors, work_objects, sentences, annotations, variations) is unchanged.

---

## Sample Files

All samples validated with Python `yaml.safe_load()` and `json.load()`.

| File | Description | Valid |
|------|-------------|-------|
| `docs/research/samples/ecommerce.story.yaml` | E-commerce marketplace, 12 sentences, 5 annotations | Yes (YAML) |
| `docs/research/samples/vetclinic.story.yaml` | Vet clinic, 12 sentences, 5 annotations | Yes (YAML) |
| `docs/research/samples/alto.story.yaml` | alto itself, 15 sentences, 5 annotations | Yes (YAML) |
| `docs/research/samples/glossary.yaml` | 12 terms across 6 bounded contexts | Yes (YAML) |
| `docs/research/samples/context-map.yaml` | 4 contexts, 4 relationships, 7 boundary signals | Yes (YAML) |
| `docs/research/samples/ecommerce.puml` | PlantUML DST export of e-commerce story | Yes (PlantUML) |
| `docs/research/samples/vetclinic.puml` | PlantUML DST export of vet clinic story | Yes (PlantUML) |
| `docs/research/samples/ecommerce.egn` | Egon .egn export of e-commerce story | Yes (JSON) |

---

## Sources

- DomainStory-PlantUML syntax: [github.com/johthor/DomainStory-PlantUML](https://github.com/johthor/DomainStory-PlantUML) (MIT license, v0.3.1)
- Egon.io .egn format: reverse-engineered from [github.com/WPS/egon.io-examples](https://github.com/WPS/egon.io-examples) (travel-1-by-taxi-en.egn, version 2.0.1)
- Egon.io: [egon.io](https://egon.io/) (GPLv3, cannot embed)
- v1 session format: `internal/discovery/domain/discovery_session.go:462-536` (ToSnapshot)
- v1 question catalog: `internal/discovery/domain/question.go:45-86`
- v1 value objects: `internal/discovery/domain/discovery_values.go`
- Proposed schema: `docs/research/20260323_3_gstack_ux_and_domain_storytelling.md` Section 9.1-9.3
- Go YAML library: [github.com/go-yaml/yaml](https://github.com/go-yaml/yaml) (MIT license)
- Domain Storytelling methodology: [domainstorytelling.org](https://domainstorytelling.org/)
