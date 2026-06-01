---
last_reviewed: 2026-06-01
---
# alto-scaffold/ Internal File Formats

This document describes every file that alto creates inside the `alto-scaffold/` directory.
These schemas are the source of truth for manual creation, debugging, and the `alto import` command.

## File Inventory

### Created by `alto init` (BootstrapHandler)

Source: `internal/bootstrap/application/bootstrap_handler.go:34-42` (`plannedFiles` list)
Content generators: `internal/bootstrap/infrastructure/content.go`

| File | Format | Generator Function | Source Line |
|------|--------|--------------------|-------------|
| `alto-scaffold/config.toml` | TOML | `AltoConfigContent()` | `content.go:35` |
| `alto-scaffold/knowledge/_index.toml` | TOML | `KnowledgeIndexContent()` | `content.go:43` |
| `alto-scaffold/maintenance/doc-registry.toml` | TOML | `DocRegistryContent()` | `content.go:67` |

### Created by `alto guide` (ArtifactGenerationHandler / DiscoveryHandler)

Source: `internal/discovery/application/artifact_generation_handler.go:127`

| File | Format | Generator Function | Source Line |
|------|--------|--------------------|-------------|
| `alto-scaffold/bounded_context_map.yaml` | YAML | `renderBoundedContextMapYAML()` | `artifact_generation_handler.go:391` |
| `alto-scaffold/stories/<name>.story.yaml` | YAML | TBD (Phase 1) | -- |
| `alto-scaffold/glossary.yaml` | YAML | TBD (Phase 1) | -- |
| `alto-scaffold/context-map.yaml` | YAML | TBD (Phase 1) | -- |

### Files that do NOT exist

- ~~`session.json`~~ — zero references in Go source. No such file is created by any code path.

---

## 1. `alto-scaffold/config.toml`

**Purpose:** Project-level alto configuration. Created during `alto init` with detected project settings.

**Generator:** `internal/bootstrap/infrastructure/content.go` — `AltoConfigContent(config domain.ProjectConfig)`

### Schema

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `project.name` | string | yes | Project name (directory name during `alto init`) |
| `project.language` | string | no | Detected language (`"go"`, `"python"`, `"typescript"`). Omitted if not detected. |
| `project.module_path` | string | no | Module path extracted from manifest (e.g. `go.mod`). Omitted if not detected. |
| `tools.detected` | string[] | yes | AI coding tools found in project directory. Empty array if none. |
| `discovery.completed` | boolean | yes | Whether guided discovery has been completed. |
| `llm.provider` | string | no | LLM provider. Commented out by default. |
| `llm.model` | string | no | LLM model. Commented out by default. |
| `llm.api_key_env` | string | no | Environment variable name for API key. Commented out by default. |

### Example

```toml
# alto project configuration

[project]
name = "my-service"
language = "go"
module_path = "github.com/user/my-service"

[tools]
detected = ["claude", "cursor"]

[discovery]
completed = false

# [llm]
# provider = ""
# model = ""
# api_key_env = ""
# Uncomment and configure when LLM features are enabled.
```

---

## 2. `alto-scaffold/knowledge/_index.toml`

**Purpose:** Index of the knowledge base. Maps section names to subdirectories under `alto-scaffold/knowledge/`. Each section contains RLM-addressable documents.

**Generator:** `internal/bootstrap/infrastructure/content.go:43` — `KnowledgeIndexContent()`

### Schema

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `knowledge.version` | integer | yes | Knowledge index schema version |
| `sections` | array of table | yes | List of knowledge sections |
| `sections[].name` | string | yes | Section directory name under `alto-scaffold/knowledge/` |
| `sections[].description` | string | yes | Human-readable section purpose |

### Example

```toml
# alto knowledge base index
#
# Sections map to subdirectories under alto-scaffold/knowledge/.
# Each section contains RLM-addressable documents.

[knowledge]
version = 1

[[sections]]
name = "ddd"
description = "DDD patterns, tactical and strategic references"

[[sections]]
name = "tools"
description = "AI coding tool conventions (versioned per tool)"

[[sections]]
name = "conventions"
description = "TDD, SOLID, quality gate references"
```

---

## 3. `alto-scaffold/maintenance/doc-registry.toml`

**Purpose:** Tracks which project documents to monitor for freshness, their owners, and review cadence. Used by `alto doc-health`.

**Generator:** `internal/bootstrap/infrastructure/content.go:67` — `DocRegistryContent()`

### Schema

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `registry.version` | integer | yes | Registry schema version |
| `docs` | array of table | yes | List of monitored documents |
| `docs[].path` | string | yes | Relative path from project root |
| `docs[].owner` | string | yes | Ownership category (e.g. `"product"`, `"architecture"`) |
| `docs[].review_days` | integer | yes | Maximum days between reviews before flagging as stale |

### Example

```toml
# alto document registry
#
# Tracks which docs to monitor, their owners, and review cadence.

[registry]
version = 1

[[docs]]
path = "docs/PRD.md"
owner = "product"
review_days = 90

[[docs]]
path = "docs/DDD.md"
owner = "architecture"
review_days = 90

[[docs]]
path = "docs/ARCHITECTURE.md"
owner = "architecture"
review_days = 90
```

---

## 4. `alto-scaffold/bounded_context_map.yaml`

**Purpose:** Machine-readable map of bounded contexts, their subdomain classifications, layers, and inter-context relationships. Used by `alto fitness generate` to validate architecture conformance.

**Generator:** `internal/discovery/application/artifact_generation_handler.go:391` — `renderBoundedContextMapYAML()`
**Parser:** `internal/fitness/infrastructure/bounded_context_map_parser.go:15-31`

### Generator vs Parser Gap

The **generator** emits a subset of the full schema:
- Includes: `project`, `bounded_contexts[].{name, module_path, classification, rationale, layers}`
- Omits: `bounded_contexts[].relationships` — the generator does not produce relationship data

The **parser** accepts the full schema including `relationships`. Users adding relationships must do so manually.

### Schema

#### `project` (required)

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | yes | Project name. Validated: must be non-empty. |
| `root_package` | string | yes | Go module root package path. Validated: must be non-empty. |

#### `bounded_contexts[]` (required)

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | yes | Bounded context name (e.g. `"Bootstrap"`, `"Discovery"`) |
| `module_path` | string | yes | Snake_case directory name under `internal/` (e.g. `"bootstrap"`) |
| `classification` | string | yes | Subdomain classification. See enum below. |
| `rationale` | string | no | Why this classification was chosen. Generator emits this; parser ignores it. |
| `layers` | string[] | yes | DDD layers present (typically `["domain", "application", "infrastructure"]`) |
| `relationships` | array | no | Inter-context relationships. See below. **Not emitted by generator.** |

#### `bounded_contexts[].relationships[]` (optional, parser-only)

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `target` | string | yes | Name of the related bounded context |
| `direction` | string | yes | Relationship direction. See enum below. |
| `pattern` | string | yes | Integration pattern. See enum below. |

### Enums

#### `SubdomainClassification`

Source: `internal/shared/domain/valueobjects/domain_values.go:14-16`

| Value | Description |
|-------|-------------|
| `"core"` | Core subdomain — competitive advantage, highest investment |
| `"supporting"` | Supporting subdomain — necessary but not differentiating |
| `"generic"` | Generic subdomain — commodity, buy or adopt off-the-shelf |

#### `RelationshipDirection`

Source: `internal/fitness/domain/bounded_context_map.go:12-13`

| Value | Description |
|-------|-------------|
| `"upstream"` | This context is upstream (provides data/events) |
| `"downstream"` | This context is downstream (consumes data/events) |

#### `RelationshipPattern`

Source: `internal/fitness/domain/bounded_context_map.go:26-29`

| Value | Description |
|-------|-------------|
| `"domain_event"` | Communication via domain events |
| `"shared_kernel"` | Shared code/types between contexts |
| `"acl"` | Anti-corruption layer isolates this context |
| `"open_host"` | Published API/protocol for consumers |

### Example (generator output)

```yaml
project:
  name: my-service
  root_package: github.com/project/my_service
bounded_contexts:
  - name: Bootstrap
    module_path: bootstrap
    classification: supporting
    rationale: Scaffolding — necessary but not core business logic
    layers:
      - domain
      - application
      - infrastructure
  - name: Discovery
    module_path: discovery
    classification: core
    rationale: Primary value proposition — guided DDD discovery
    layers:
      - domain
      - application
      - infrastructure
```

### Example (with relationships — manual addition)

```yaml
project:
  name: my-service
  root_package: github.com/project/my_service
bounded_contexts:
  - name: Discovery
    module_path: discovery
    classification: core
    layers:
      - domain
      - application
      - infrastructure
    relationships:
      - target: Bootstrap
        direction: downstream
        pattern: domain_event
  - name: Fitness
    module_path: fitness
    classification: supporting
    layers:
      - domain
      - application
      - infrastructure
    relationships:
      - target: Discovery
        direction: upstream
        pattern: acl
```

### Validation Rules

The parser (`bounded_context_map_parser.go:71-78`) enforces:

1. `project.name` must be non-empty
2. `project.root_package` must be non-empty
3. Each `classification` must be one of: `core`, `supporting`, `generic`
4. Each `direction` must be one of: `upstream`, `downstream`
5. Each `pattern` must be one of: `domain_event`, `shared_kernel`, `acl`, `open_host`

---

## 5. `alto-scaffold/stories/<name>.story.yaml`

**Purpose:** Machine-readable domain story captured during `alto guide` (Discovery). One file per story. Used by boundary detection, DDD.md generation, and ticket pipeline.

**Generator:** TBD (Phase 1 tickets will create story capture and persistence)
**Parser:** TBD (Phase 1)

**Schema source:** `docs/research/20260323_6_story_format_validation.md` Section 3 (finalized)

### Top-level Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `title` | string | yes | Human-readable story title |
| `type` | StoryType | yes | Granularity of the story |
| `time` | TimeType | yes | Whether the story describes current or future state |
| `purity` | PurityType | yes | Whether the story is domain-pure or includes digital systems |
| `trigger` | string | yes | What initiates this story |
| `actors` | []Actor | yes | List of story participants |
| `work_objects` | []WorkObject | yes | List of things acted upon |
| `sentences` | []Sentence | yes | Ordered story sentences |
| `annotations` | []Annotation | no | Business rules and constraints |
| `variations` | []string | no | Pointers to alternative stories or scenario descriptions |

### `actors[]`

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | yes | Actor's name in ubiquitous language |
| `type` | ActorType | yes | Classification of the actor |
| `trust` | TrustLevel | yes | Provenance of this element |
| `source` | string | conditional | Citation. Required when `trust` is `ai_researched`. |

### `work_objects[]`

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | yes | Work object's name in ubiquitous language |
| `type` | WorkObjectType | yes | Classification of the work object |
| `trust` | TrustLevel | yes | Provenance of this element |
| `source` | string | conditional | Citation. Required when `trust` is `ai_researched`. |

### `sentences[]`

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `step` | int | yes | Sequence number (1-based, sequential) |
| `subject` | string | yes | Actor name performing the action. Must reference a name from `actors`. |
| `activity` | string | yes | Verb phrase in domain language |
| `object` | string | yes | Work object or actor being acted upon. Must reference a name from `work_objects` or `actors`. |
| `preposition` | string | no | Connecting word (for, to, via, using, from, with, in, about, based on, on). If set, `indirect_object` must also be set. |
| `indirect_object` | string | no | Second work object or actor. Must reference a name from `work_objects` or `actors`. If set, `preposition` must also be set. |
| `trust` | TrustLevel | yes | Provenance of this sentence |
| `source` | string | conditional | Citation. Required when `trust` is `ai_researched`. |

### `annotations[]`

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `text` | string | yes | The business rule or constraint text |
| `type` | AnnotationType | yes | Classification of the annotation |
| `sentence` | int | no | Step number this annotation applies to. Omit for story-wide annotations. |
| `trust` | TrustLevel | no | Provenance. Defaults to story-level trust if omitted. |
| `source` | string | no | Citation when `trust` is `ai_researched` |

### Enums

#### `StoryType`

| Value | Description |
|-------|-------------|
| `coarse_grained` | High-level overview, focuses on the main flow |
| `fine_grained` | Detailed sub-flow, zooms into a specific part of the domain |

#### `TimeType`

| Value | Description |
|-------|-------------|
| `as_is` | Describes the current state of the domain process |
| `to_be` | Describes the desired future state |

#### `PurityType`

| Value | Description |
|-------|-------------|
| `pure` | Domain-only story, no digital systems involved |
| `digitalized` | Story includes digital systems (e.g., payment gateways, databases) |

#### `ActorType`

| Value | Description |
|-------|-------------|
| `person` | A human participant |
| `system` | An automated or external system |
| `group` | A team or organizational unit |

#### `WorkObjectType`

| Value | Description |
|-------|-------------|
| `document` | A structured record (order, invoice, medical record) |
| `folder` | A collection of documents |
| `call` | A phone or video call |
| `email` | An email message |
| `conversation` | A verbal or chat conversation |
| `info` | An informational element (e.g., a pet, a status) |
| `data` | Raw or structured data (e.g., available slots, metrics) |

#### `AnnotationType`

| Value | Description |
|-------|-------------|
| `constraint` | A business rule that must be enforced |
| `invariant` | A condition that must always be true |
| `assumption` | An element that needs future validation |

#### `TrustLevel`

| Value | Description |
|-------|-------------|
| `user_stated` | User explicitly said this during the story |
| `user_confirmed` | AI proposed, user confirmed |
| `ai_researched` | AI discovered via domain research, not yet confirmed. `source` is required. |
| `ai_inferred` | AI inferred from context, lowest confidence |

### Validation Rules

1. `title` must be non-empty
2. Every `subject` in sentences must reference a name from `actors`
3. Every `object` and `indirect_object` in sentences must reference a name from either `work_objects` or `actors`
4. Step numbers must be sequential starting from 1
5. If `preposition` is set, `indirect_object` must also be set (and vice versa)
6. `source` is required when `trust` is `ai_researched`
7. At least 1 actor, 1 work object, and 1 sentence required

### Example

```yaml
# E-commerce Marketplace: Customer Purchases Product
# alto Domain Story Format v0.1

title: "Customer Purchases Product from Marketplace"
type: coarse_grained
time: to_be
purity: digitalized
trigger: "Customer searches for a product"

actors:
  - name: Customer
    type: person
    trust: user_stated
  - name: Payment Gateway
    type: system
    trust: ai_researched
    source: "Stripe/PayPal integration patterns"
  - name: Platform
    type: system
    trust: user_confirmed

work_objects:
  - name: Product Listing
    type: document
    trust: user_stated
  - name: Shopping Cart
    type: document
    trust: user_stated
  - name: Order
    type: document
    trust: user_stated
  - name: Payment
    type: document
    trust: user_confirmed

sentences:
  - step: 1
    subject: Customer
    activity: browses
    object: Product Listing
    trust: user_stated
  - step: 2
    subject: Customer
    activity: adds
    object: Product Listing
    preposition: to
    indirect_object: Shopping Cart
    trust: user_stated
  - step: 3
    subject: Customer
    activity: initiates checkout of
    object: Shopping Cart
    trust: user_stated
  - step: 4
    subject: Customer
    activity: submits
    object: Payment
    preposition: via
    indirect_object: Payment Gateway
    trust: user_confirmed
  - step: 5
    subject: Platform
    activity: creates
    object: Order
    preposition: from
    indirect_object: Shopping Cart
    trust: user_confirmed

annotations:
  - text: "Customer must be authenticated before checkout"
    sentence: 3
    type: constraint
    trust: user_stated
  - text: "Payment must be authorized before Order is created"
    type: invariant
    trust: user_stated

variations:
  - "Customer pays via PayPal instead of credit card"
  - "Payment is declined by gateway"
  - "Seller is out of stock after order placed"
```

---

## 6. `alto-scaffold/glossary.yaml`

**Purpose:** Ubiquitous language glossary with trust levels and bounded context ownership. Used by DDD.md generation and ticket pipeline. Terms are extracted from domain stories captured during `alto guide`.

**Generator:** TBD (Phase 1 tickets will create glossary extraction)
**Parser:** TBD (Phase 1)

**Schema source:** `docs/research/20260323_6_story_format_validation.md` Section 4 (finalized)

### Top-level Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `version` | int | yes | Schema version (currently `1`) |
| `terms` | []Term | yes | List of ubiquitous language terms |

### `terms[]`

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `term` | string | yes | The term in ubiquitous language |
| `definition` | string | yes | What this term means in the domain |
| `context` | string | yes | Which bounded context this term belongs to |
| `trust` | TrustLevel | yes | Provenance of this term |
| `source` | string | conditional | Citation. Required when `trust` is `ai_researched`. |
| `stories` | []string | yes | Story file references where this term appears |
| `aliases` | []string | no | Alternative names for this term in other contexts |
| `note` | string | no | Additional notes (e.g., language difference signals across contexts) |

### Glossary Extraction Rules

1. Every `work_objects[].name` and `actors[].name` from stories is a candidate glossary term
2. Every `sentences[].activity` verb is a candidate for the domain activity vocabulary
3. `aliases` capture language difference signals across bounded contexts (e.g., "Shopping Cart" vs "Basket")
4. `note` captures cross-context naming conflicts that signal boundaries

### Validation Rules

1. `term` must be non-empty
2. `definition` must be non-empty
3. `context` must be non-empty
4. `trust` must be a valid TrustLevel enum value
5. `source` is required when `trust` is `ai_researched`

### Example

```yaml
# Ubiquitous Language Glossary
# alto Glossary Format v0.1

version: 1

terms:
  # --- Catalog Context ---
  - term: Product Listing
    definition: "A seller's published offering including title, description, price, and images"
    context: Catalog
    trust: user_stated
    stories:
      - "ecommerce/01-customer-purchases-product"
    aliases: []

  - term: Inventory
    definition: "The quantity of a specific product available for sale, managed by the seller"
    context: Catalog
    trust: user_stated
    stories:
      - "ecommerce/01-customer-purchases-product"
    aliases:
      - "stock"

  # --- Ordering Context ---
  - term: Shopping Cart
    definition: "A temporary collection of product listings selected by a customer before checkout"
    context: Ordering
    trust: user_stated
    stories:
      - "ecommerce/01-customer-purchases-product"
    note: "Called 'Basket' in UK English -- language difference signal between markets"
    aliases:
      - "Basket"

  # --- Payment Context ---
  - term: Commission
    definition: "The platform's fee calculated as a percentage (8-15%) of each order's total"
    context: Payment
    trust: user_stated
    stories:
      - "ecommerce/01-customer-purchases-product"
    note: "Commission percentage is a business rule, not a technical constant"
    aliases:
      - "Platform Fee"
```

---

## 7. `alto-scaffold/context-map.yaml`

**Purpose:** Bounded context map with boundary signals, relationships, and subdomain classification. Derived from domain story analysis during `alto guide`. Used by fitness function generation and ticket pipeline.

**Generator:** TBD (Phase 1 tickets will create context map generation)
**Parser:** TBD (Phase 1)

**Schema source:** `docs/research/20260323_6_story_format_validation.md` Section 5 (finalized)

**Note:** This file is distinct from `alto-scaffold/bounded_context_map.yaml` (Section 4). The bounded_context_map is generated by the existing ArtifactGenerationHandler and focused on module paths and DDD layers. This context-map is generated from domain stories and captures boundary signals, actor/work-object ownership, and DDD relationship patterns.

### Top-level Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `version` | int | yes | Schema version (currently `1`) |
| `project` | string | yes | Project name |
| `contexts` | []Context | yes | List of bounded contexts |
| `relationships` | []Relationship | yes | How contexts relate to each other |

### `contexts[]`

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | yes | Context name |
| `classification` | SubdomainClassification | yes | Subdomain classification |
| `confidence` | float | yes | 0.0-1.0, confidence in this boundary |
| `actors` | []string | yes | Actors that appear in this context |
| `work_objects` | []string | yes | Work objects that belong to this context |
| `boundary_signals` | []BoundarySignal | yes | Evidence for why this is a separate context |
| `stories` | []string | yes | Story:step references (e.g., `"ecommerce/01:1-5"`) |
| `trust` | TrustLevel | yes | Provenance of this boundary |

### `contexts[].boundary_signals[]`

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `type` | BoundarySignalType | yes | Type of boundary evidence |
| `description` | string | yes | Human-readable explanation of the signal |

### `relationships[]`

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `upstream` | string | yes | Upstream context name |
| `downstream` | string | yes | Downstream context name |
| `type` | RelationshipType | yes | DDD relationship pattern |
| `shared` | []string | yes | Work objects shared across this boundary |
| `description` | string | no | Human-readable relationship description |

### Enums

#### `SubdomainClassification`

| Value | Description |
|-------|-------------|
| `core` | Core subdomain -- competitive advantage, highest investment |
| `supporting` | Supporting subdomain -- necessary but not differentiating |
| `generic` | Generic subdomain -- commodity, buy or adopt off-the-shelf |

#### `BoundarySignalType`

| Value | Description |
|-------|-------------|
| `different_trigger` | Contexts are triggered by different events |
| `one_way_flow` | Data flows in one direction between contexts |
| `language_difference` | Same concept has different names across contexts |
| `different_lifecycle` | Entities in each context change independently |
| `external_system` | A third-party system creates a natural boundary |
| `different_actor` | Different actors operate in each context |
| `complex_rules` | Complex business rules concentrate in one context |

#### `RelationshipType`

| Value | Description |
|-------|-------------|
| `shared_kernel` | Shared code or types between contexts |
| `customer_supplier` | Downstream context's needs influence upstream |
| `conformist` | Downstream conforms to upstream's model |
| `anticorruption_layer` | Downstream translates upstream's model via ACL |
| `open_host_service` | Upstream provides a published API/protocol |
| `published_language` | Communication via a well-documented shared language |
| `partnership` | Two contexts evolve together cooperatively |
| `separate_ways` | Contexts have no integration, operate independently |

### Validation Rules

1. `project` must be non-empty
2. Each context must have a non-empty `name`
3. `classification` must be one of: `core`, `supporting`, `generic`
4. `type` on relationships must be a valid RelationshipType enum value
5. `confidence` must be between 0.0 and 1.0 inclusive

### Example

```yaml
# Bounded Context Map
# alto Context Map Format v0.1

version: 1
project: "E-commerce Marketplace"

contexts:
  - name: Catalog
    classification: supporting
    confidence: 0.85
    actors:
      - Seller
      - Customer
    work_objects:
      - Product Listing
      - Inventory
    boundary_signals:
      - type: different_lifecycle
        description: "Product Listings are managed independently of Orders"
      - type: language_difference
        description: "'Product Listing' in Catalog becomes 'Line Item' in Ordering"
    stories:
      - "ecommerce/01-customer-purchases-product:1,9"
    trust: user_stated

  - name: Ordering
    classification: core
    confidence: 0.90
    actors:
      - Customer
      - Platform
    work_objects:
      - Shopping Cart
      - Order
    boundary_signals:
      - type: different_trigger
        description: "Ordering is triggered by checkout; Catalog by browsing"
      - type: one_way_flow
        description: "Product data flows from Catalog to Ordering, never back"
      - type: complex_rules
        description: "Order creation involves payment, inventory, commission"
    stories:
      - "ecommerce/01-customer-purchases-product:2-8"
    trust: user_stated

relationships:
  - upstream: Catalog
    downstream: Ordering
    type: conformist
    shared:
      - Product Listing
    description: "Ordering conforms to Catalog's product model; snapshots at order time"

  - upstream: Ordering
    downstream: Payment
    type: customer_supplier
    shared:
      - Order
    description: "Payment serves Ordering's needs; Ordering defines payment requirements"
```
