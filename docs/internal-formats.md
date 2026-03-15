# .alty/ Internal File Formats

This document describes every file that alty creates inside the `.alty/` directory.
These schemas are the source of truth for manual creation, debugging, and the `alty import` command.

## File Inventory

### Created by `alty init` (BootstrapHandler)

Source: `internal/bootstrap/application/bootstrap_handler.go:34-42` (`plannedFiles` list)
Content generators: `internal/bootstrap/infrastructure/content.go`

| File | Format | Generator Function | Source Line |
|------|--------|--------------------|-------------|
| `.alty/config.toml` | TOML | `AltyConfigContent()` | `content.go:35` |
| `.alty/knowledge/_index.toml` | TOML | `KnowledgeIndexContent()` | `content.go:43` |
| `.alty/maintenance/doc-registry.toml` | TOML | `DocRegistryContent()` | `content.go:67` |

### Created by `alty guide` (ArtifactGenerationHandler)

Source: `internal/discovery/application/artifact_generation_handler.go:127`

| File | Format | Generator Function | Source Line |
|------|--------|--------------------|-------------|
| `.alty/bounded_context_map.yaml` | YAML | `renderBoundedContextMapYAML()` | `artifact_generation_handler.go:391` |

### Files that do NOT exist

- ~~`session.json`~~ — zero references in Go source. No such file is created by any code path.

---

## 1. `.alty/config.toml`

**Purpose:** Project-level alty configuration. Created during `alty init` with detected project settings.

**Generator:** `internal/bootstrap/infrastructure/content.go` — `AltyConfigContent(config domain.ProjectConfig)`

### Schema

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `project.name` | string | yes | Project name (directory name during `alty init`) |
| `project.language` | string | no | Detected language (`"go"`, `"python"`, `"typescript"`). Omitted if not detected. |
| `project.module_path` | string | no | Module path extracted from manifest (e.g. `go.mod`). Omitted if not detected. |
| `tools.detected` | string[] | yes | AI coding tools found in project directory. Empty array if none. |
| `discovery.completed` | boolean | yes | Whether guided discovery has been completed. |
| `llm.provider` | string | no | LLM provider. Commented out by default. |
| `llm.model` | string | no | LLM model. Commented out by default. |
| `llm.api_key_env` | string | no | Environment variable name for API key. Commented out by default. |

### Example

```toml
# alty project configuration

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

## 2. `.alty/knowledge/_index.toml`

**Purpose:** Index of the knowledge base. Maps section names to subdirectories under `.alty/knowledge/`. Each section contains RLM-addressable documents.

**Generator:** `internal/bootstrap/infrastructure/content.go:43` — `KnowledgeIndexContent()`

### Schema

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `knowledge.version` | integer | yes | Knowledge index schema version |
| `sections` | array of table | yes | List of knowledge sections |
| `sections[].name` | string | yes | Section directory name under `.alty/knowledge/` |
| `sections[].description` | string | yes | Human-readable section purpose |

### Example

```toml
# alty knowledge base index
#
# Sections map to subdirectories under .alty/knowledge/.
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

## 3. `.alty/maintenance/doc-registry.toml`

**Purpose:** Tracks which project documents to monitor for freshness, their owners, and review cadence. Used by `alty doc-health`.

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
# alty document registry
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

## 4. `.alty/bounded_context_map.yaml`

**Purpose:** Machine-readable map of bounded contexts, their subdomain classifications, layers, and inter-context relationships. Used by `alty fitness generate` to validate architecture conformance.

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
