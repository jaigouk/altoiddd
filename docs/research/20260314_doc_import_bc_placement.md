# Spike: Doc Import Bounded Context Placement

**Ticket:** alty-cli-2pg
**Date:** 2026-03-14
**Status:** Complete

## Decision

**New "DocImport" bounded context** at `internal/docimport/` with a `DocImporter` port interface in its application layer, producing `*ddd.DomainModel` (from shared kernel).

## Rationale Summary

Doc import is a distinct capability with its own domain language (parsing, extraction, inference confirmation), its own input source (markdown files, not conversational Q&A), and its own invariants (document structure validation, section presence checks). It does not fit in any existing context without violating SRP or bounded context autonomy.

---

## Analysis of Each Option

### Option 1: New "DocImport" Bounded Context (RECOMMENDED)

**Cohesion:** High. The capability is self-contained: read markdown files, parse structure, extract domain elements, construct a `*ddd.DomainModel`. All operations share the same lifecycle and failure modes (file I/O, parsing errors, incomplete documents).

**Coupling:** Minimal. Depends only on shared kernel types (`*ddd.DomainModel`, value objects in `internal/shared/domain/valueobjects/`, `internal/shared/domain/ddd/`). No imports from discovery, fitness, or ticket contexts. No new cross-context dependencies.

**DDD alignment:** DDD.md already lists "DocImport" and "DocParser" as terms pending this spike (lines 232-234). The glossary explicitly marks them as "TBD (pending spike alty-cli-2pg)" without assigning a bounded context. Creating `internal/docimport/` gives them a home.

**Practicality:**
- 3 new directories: `internal/docimport/{domain,application,infrastructure}/`
- 1 new entry in `composition/app.go` (handler wiring)
- 0 changes to existing bounded contexts
- 0 import cycle risk (depends only on shared kernel, nothing depends on it)

**Files affected:**
- `internal/composition/app.go` — add `DocImportHandler` field and wiring (~10 lines)
- `docs/DDD.md` — update DocImport/DocParser entries from "TBD" to "DocImport"
- `.alty/bounded_context_map.yaml` (if applicable) — add DocImport context

### Option 2: Extend Shared Kernel with DomainModel Factory

**Rejected.** The shared kernel (`internal/shared/domain/ddd/`) owns the `DomainModel` aggregate and its invariant checks. Adding markdown parsing logic here would:

1. Violate the shared kernel's purpose. The shared kernel holds types used across contexts; it does not contain use-case-specific logic. Markdown parsing is use-case-specific.
2. Introduce infrastructure concerns (file I/O, parsing) into a domain package. `internal/shared/domain/ddd/domain_model.go` currently has zero external dependencies beyond the shared kernel itself. Adding `os.ReadFile` or a YAML/markdown parser would break the "domain has zero external deps" rule enforced by depguard (source: `.golangci.yml`, `arch-go.yml`).
3. Create a coupling magnet. Every context that touches DomainModel would transitively depend on parsing logic.

### Option 3: Anti-Corruption Layer (Between Docs and DomainModel)

**Partially correct, but incomplete.** An ACL is a *pattern*, not a *placement decision*. The doc import adapter will indeed function as an ACL — translating external markdown structure into internal domain types. But the ACL must live *inside* a bounded context. The question is: which one?

If we place the ACL in the shared kernel, we get Option 2's problems. If we place it in discovery, we get Option 4's problems. The ACL pattern naturally leads to Option 1: a dedicated context that owns the translation boundary.

**Verdict:** The ACL is a design pattern to use *within* Option 1, not a standalone option.

### Option 4: Extend Guided Discovery

**Rejected.** Guided Discovery's domain is the conversational Q&A flow. Every domain concept is session-oriented:

- `DiscoverySession` — tracks question progression (source: `internal/discovery/domain/`)
- `DiscoveryCompletedEvent` — carries `SessionID()`, `Persona()`, `Answers()` (source: `internal/discovery/domain/`)
- `buildModel()` — maps Q1-Q10 answers to domain model elements (source: `internal/discovery/application/artifact_generation_handler.go:195-231`)
- `ArtifactGenerationHandler.BuildPreview()` — takes `DiscoveryCompletedEvent` as input (source: line 63-66)

Doc import has no session, no persona, no Q&A answers, and cannot produce a `DiscoveryCompletedEvent` (as stated in the spike ticket). Forcing it into discovery would require:

1. A synthetic `DiscoveryCompletedEvent` with fake session data — violates the event's invariants
2. An alternative code path in `ArtifactGenerationHandler` that bypasses its entire `buildModel()` pipeline — the handler would have two unrelated responsibilities
3. New domain concepts (DocParser, DocInference) that have nothing to do with guided questions

This violates SRP and bounded context autonomy. Discovery's ubiquitous language is about guides, questions, phases, personas, registers, and playbacks — not parsing and extraction.

---

## Port Interface Draft

```go
// Package application contains the DocImport bounded context's application layer.
package application

import (
    "context"

    "github.com/alty-cli/alty/internal/shared/domain/ddd"
)

// DocImporter parses existing documentation files and constructs a DomainModel.
// Implementations handle specific document formats (markdown DDD.md, ARCHITECTURE.md).
//
// The returned DomainModel is NOT finalized — the caller decides whether to
// call Finalize() (which enforces invariants and emits DomainModelGenerated).
// This allows the caller to enrich or correct the model before validation.
type DocImporter interface {
    // Import reads documentation from the given directory and returns a
    // partially-populated DomainModel. The model may have warnings for
    // sections that could not be parsed.
    //
    // docDir is typically "docs/" — the directory containing DDD.md and
    // ARCHITECTURE.md.
    Import(ctx context.Context, docDir string) (*ddd.DomainModel, error)
}

// DocImportHandler orchestrates the doc import workflow.
type DocImportHandler struct {
    importer  DocImporter
    publisher EventPublisher // from shared application ports
}
```

**Key design decisions:**

1. **Returns unfinalized DomainModel.** The `DomainModel.Finalize()` method enforces invariants (every UL term in a story, every BC classified, every Core BC has an aggregate). Imported docs may not satisfy all invariants. The handler should present warnings and let the user decide whether to proceed with a partial model or fix the docs first.

2. **Single Import method.** The port takes a directory, not individual files. This keeps the interface stable if we add support for more file types later (e.g., OpenAPI specs, proto files).

3. **No session coupling.** The handler does not create a `DiscoverySession` or emit `DiscoveryCompletedEvent`. It works directly with `*ddd.DomainModel`.

## Adapter Placement

```
internal/docimport/
  domain/           # DocImport-specific domain types (ImportResult, ParseWarning)
  application/      # DocImportHandler, DocImporter port
    ports.go        # DocImporter interface
    handler.go      # DocImportHandler
  infrastructure/   # Concrete parsers
    markdown_doc_importer.go   # Implements DocImporter for markdown
```

The `MarkdownDocImporter` adapter lives in `internal/docimport/infrastructure/`. It:
- Reads `DDD.md` and `ARCHITECTURE.md` from the given directory
- Extracts bounded contexts from heading patterns (e.g., `### N. Context Name (Classification)`)
- Extracts UL terms from glossary tables
- Extracts aggregate designs from event storming sections
- Constructs `*ddd.DomainModel` using the existing aggregate's command methods (`AddBoundedContext`, `AddTerm`, `DesignAggregate`, etc.)

This mirrors the existing pattern in `internal/fitness/infrastructure/bounded_context_map_parser.go` (source: lines 34-69) — an infrastructure adapter that parses YAML files and returns domain types.

## Impact on Existing Contexts

| Context | Change Required | Reason |
|---------|----------------|--------|
| **discovery** | None | DocImport does not touch discovery. The existing `ArtifactGenerationHandler.BuildPreview(DiscoveryCompletedEvent)` path remains unchanged. |
| **fitness** | None | `FitnessGenerationHandler.BuildPreview(*ddd.DomainModel, ...)` already accepts `*ddd.DomainModel` directly (source: `fitness_generation_handler.go:51-56`). No change needed. |
| **ticket** | None | `TicketGenerationHandler.BuildPreview(*ddd.DomainModel, ...)` already accepts `*ddd.DomainModel` directly (source: `ticket_generation_handler.go:39-42`). No change needed. |
| **composition** | Add wiring | New `DocImportHandler` field in `App` struct, new adapter instantiation in `NewApp()`. ~10 lines. |
| **shared kernel** | None | `*ddd.DomainModel` and all VOs already exist and are sufficient. |
| **DDD.md** | Update glossary | Change DocImport and DocParser from "TBD" to "DocImport" context. |

**Key insight:** The integration point is `*ddd.DomainModel`, which already lives in the shared kernel. Both fitness and ticket handlers accept it directly — they do not care whether it came from discovery or doc import. This is the DDD shared kernel pattern working as intended.

## Event Flow Comparison

**Current (discovery path):**
```
DiscoverySession.Complete() → DiscoveryCompletedEvent
  → ArtifactGenerationHandler.BuildPreview(event)
    → buildModel(event) → *ddd.DomainModel
      → model.Finalize() → DomainModelGenerated event
        → FitnessGenerationHandler, TicketGenerationHandler consume DomainModel
```

**Proposed (doc import path):**
```
DocImportHandler.Import(docDir)
  → MarkdownDocImporter.Import(ctx, docDir) → *ddd.DomainModel (unfinalized)
    → user reviews/corrects model
      → model.Finalize() → DomainModelGenerated event
        → FitnessGenerationHandler, TicketGenerationHandler consume DomainModel
```

Both paths converge at `*ddd.DomainModel`. The `DomainModelGenerated` event is emitted by `Finalize()` regardless of the source. Downstream consumers (fitness, ticket, tool translation) are source-agnostic.

## Follow-Up Ticket Updates

### alty-cli-623.2 (Implement alty import command)

**Current DDD Alignment section says:**
- Bounded Context: Discovery

**Should be updated to:**
- Bounded Context: DocImport
- Layer: Application (DocImportHandler) + Infrastructure (MarkdownDocImporter)
- Port interface: DocImporter (in `internal/docimport/application/ports.go`)
- Adapter: MarkdownDocImporter (in `internal/docimport/infrastructure/`)
- Output type: `*ddd.DomainModel` (unfinalized)
- No DiscoveryCompletedEvent — works directly with DomainModel
- Design section's parsing strategy is correct but should reference the new BC path

### alty-cli-623.3 (Add --from-docs flag to generate commands)

**Current DDD Alignment section says:**
- Bounded Context: Ticket, Fitness, Bootstrap

**Should be updated to:**
- The `--from-docs` flag triggers DocImportHandler, which produces a `*ddd.DomainModel`
- The model is then passed to existing handlers (FitnessGenerationHandler, TicketGenerationHandler) unchanged
- No changes needed in ticket, fitness, or bootstrap contexts themselves
- The CLI layer (`cmd/alty/`) handles the flag and calls DocImportHandler instead of requiring a DiscoveryCompletedEvent
- Consider: `--from-docs` may be better named `--import` to align with the `alty import` command

## Risks

1. **Parsing fidelity.** Markdown docs written by humans will have inconsistent structure. The MarkdownDocImporter must be tolerant of variations while still extracting useful domain model elements. Mitigation: return warnings for unparseable sections rather than failing.

2. **Invariant gaps.** Imported docs may not satisfy DomainModel's Finalize() invariants (e.g., missing subdomain classifications, UL terms not appearing in stories). Mitigation: the handler should support partial models and present the user with a list of what needs to be added.

3. **Scope creep.** "Parse any markdown" is unbounded. Mitigation: start with a strict parser for alty-generated docs (which have known heading structure), then expand to freeform docs in a later ticket.
