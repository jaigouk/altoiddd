# Spike: DiscoverySession Refactoring for Conversational Flow

**Ticket:** alto-cli-3yp
**Date:** 2026-03-14
**Researcher:** researcher-session
**Status:** Complete

## Research Questions

1. Should `ConversationState` replace or extend `DiscoverySession`?
2. How do we preserve existing phase ordering, playback, and MVP invariants?
3. Can the aggregate support both fixed-question and adaptive-question modes via strategy pattern?
4. What is the minimal domain model change that enables conversational flow without breaking existing tests?

---

## 1. Invariant Map

Full analysis of every DiscoverySession invariant and its applicability to each mode.

**Source:** `internal/discovery/domain/discovery_session.go` (all line references to this file unless noted)

| # | Invariant | Fixed-Question Mode | Conversational Mode | Design Decision |
|---|-----------|-------------------|-------------------|-----------------|
| I1 | Cannot answer before persona detection (L239-241) | YES -- gates state machine entry | YES -- persona informs register and question tone | **Keep as-is.** Universal invariant. |
| I2 | Phase order enforced: Actors->Story->Events->Boundaries (L440-471, L30) | YES -- hardcoded in `enforcePhaseOrder()` via `questionPhases` var | NO -- LLM chooses question order based on gaps | **Mode-gated.** `enforcePhaseOrder()` skipped when mode is conversational. The LLM decides ordering; the aggregate trusts it. |
| I3 | Playback after every 3 answers (L275-277, L43) | YES -- `playbackInterval = 3` const | RETHINK -- playback still valuable but cadence should be configurable | **Configurable.** Extract `playbackInterval` to a `FlowConfig` value object. Fixed mode: 3. Conversational mode: configurable (default 5, or LLM-triggered). |
| I4 | Cannot answer/skip during PlaybackPending (L242-244, L283-285) | YES -- hard gate | YES -- playback confirmation loop still applies | **Keep as-is.** Both modes need the user to confirm understanding before proceeding. |
| I5 | Skip requires non-empty reason (L293-295) | YES | YES -- even in conversational mode, skipping domain questions needs rationale | **Keep as-is.** Universal invariant. |
| I6 | Duplicate answer rejected (L252-257) | YES -- keyed on questionID | ADAPT -- conversational mode uses dynamic question IDs, not Q1-Q10 | **Keep invariant, adapt key space.** Fixed mode: Q1-Q10. Conversational mode: generated IDs (e.g., `conv_actors_1`, `conv_story_2`). The duplicate-check logic is unchanged. |
| I7 | Empty answer rejected (L248-249) | YES | YES | **Keep as-is.** Universal invariant. |
| I8 | Complete requires MVP questions answered (L320-335) | YES -- `MVPQuestionIDs()` = {Q1, Q3, Q4, Q9, Q10} | RETHINK -- conversational mode doesn't use fixed question IDs | **Replace with coverage check.** Introduce `ModelCompletenessCheck` interface. Fixed mode: checks MVP question IDs. Conversational mode: checks that domain model has actors, primary story, invariants, bounded contexts, and classifications (same semantic coverage, different verification). |
| I9 | Mode set-once in CREATED state (L209-218) | YES | YES | **Keep as-is.** Universal invariant. Add `ModeConversational` to the enum. |
| I10 | TechStack only in CREATED/PERSONA_DETECTED (L199-206) | YES | YES | **Keep as-is.** Universal invariant. |
| I11 | ClassifyBoundedContext only in COMPLETED/ROUND_1_COMPLETE (L406-411) | YES | YES -- classification happens after discovery regardless of mode | **Keep as-is.** Universal invariant. |
| I12 | Event emitted only at terminal COMPLETED (L429-438) | YES | YES | **Keep as-is.** Universal invariant. |

**Summary:** 9 of 12 invariants apply unchanged. 3 need adaptation (I2: phase order, I3: playback cadence, I8: MVP check). None need removal.

---

## 2. Proposed Refactoring Approach

### Options Evaluated

#### Option A: Strategy Pattern (RECOMMENDED)

Introduce a `DiscoveryFlow` interface that encapsulates mode-specific behavior. The aggregate delegates to the active flow strategy.

```go
// DiscoveryFlow encapsulates mode-specific question ordering and completeness rules.
type DiscoveryFlow interface {
    // ValidateQuestionOrder checks if answering this question is allowed given current state.
    // Fixed mode: enforces phase ordering. Conversational mode: always allows.
    ValidateQuestionOrder(question QuestionRef, answered []Answer, skipped map[string]bool) error

    // IsPlaybackDue returns true if a playback should be triggered.
    // Fixed mode: every 3 answers. Conversational mode: configurable cadence.
    IsPlaybackDue(answersSinceLastPlayback int) bool

    // CheckCompleteness verifies the session has sufficient coverage to complete.
    // Fixed mode: checks MVPQuestionIDs. Conversational mode: checks model coverage.
    CheckCompleteness(answers []Answer, skipped map[string]bool) error

    // PlaybackInterval returns the number of answers between playbacks.
    PlaybackInterval() int
}
```

**New types needed:**
- `DiscoveryFlow` interface (in `domain/`)
- `FixedQuestionFlow` struct (in `domain/`) -- implements current behavior exactly
- `ConversationalFlow` struct (in `domain/`) -- implements adaptive behavior
- `QuestionRef` value object -- abstracts over fixed Question and adaptive question IDs
- `ModeConversational` constant added to `DiscoveryMode` enum
- `ModelCompletenessCheck` -- semantic coverage check for conversational mode

**Changes to DiscoverySession:**
- Add `flow DiscoveryFlow` field
- `SetMode()` now also sets the flow strategy
- `AnswerQuestion()` calls `s.flow.ValidateQuestionOrder()` instead of `s.enforcePhaseOrder()`
- `AnswerQuestion()` calls `s.flow.IsPlaybackDue()` instead of checking `playbackInterval` const
- `Complete()` calls `s.flow.CheckCompleteness()` instead of checking `MVPQuestionIDs()`
- `AnswerQuestion()` accepts any question ID (not just Q1-Q10) when in conversational mode

**Trade-offs:**
| Pro | Con |
|-----|-----|
| Zero changes to existing fixed-mode behavior | 6 new types/interfaces |
| Strategy is a domain concept (OCP) | Flow interface may accumulate methods over time |
| Easy to add future modes (e.g., `ModeHybrid`) | Indirection cost for simple operations |
| Existing tests pass without modification | New tests needed for conversational flow |
| Clean ISP -- each concern is one method | Strategy must be set before answering (new invariant) |

#### Option B: Mode Flag with Conditionals

Add `if s.Mode() == ModeConversational` branches throughout the aggregate.

**Trade-offs:**
| Pro | Con |
|-----|-----|
| Fewer new types (just the mode constant) | Violates OCP -- every new mode adds more branches |
| Lower initial complexity | SRP violation -- aggregate knows about all modes |
| No new interfaces | Harder to test modes in isolation |
| | Phase order, playback, and completion logic all need conditionals |

**Verdict:** Rejected. The aggregate already has mode-conditional logic for Deep mode (L337-343, L351-363). Adding a third mode via conditionals would make the aggregate unwieldy. The existing `ModeDeep` conditionals are a code smell that the strategy pattern would also clean up.

#### Option C: Separate Aggregate (`ConversationSession`)

Create a new aggregate root alongside `DiscoverySession`.

**Trade-offs:**
| Pro | Con |
|-----|-----|
| Total separation of concerns | Duplicates shared invariants (I1, I4-I7, I9-I12) |
| No risk to existing tests | Two aggregates emitting `DiscoveryCompletedEvent` -- confusing |
| Each aggregate is simpler | Handler, port, MCP tools all need parallel implementations |
| | ~60% code duplication in invariant enforcement |
| | Downstream consumers (artifact generation, ticket pipeline) need to handle both types |

**Verdict:** Rejected. 9 of 12 invariants are shared. Code duplication would be severe and the downstream event contract (`DiscoveryCompletedEvent`) would need to unify them anyway.

### Recommendation: Option A (Strategy Pattern)

The strategy pattern is the correct DDD solution because:
1. The "flow" is a genuine domain concept -- it represents how discovery is conducted
2. It respects OCP -- adding modes means adding strategies, not modifying the aggregate
3. It preserves all 12 invariants with minimal adaptation
4. It enables refactoring the existing Deep mode conditionals into the strategy (future cleanup)
5. Zero existing test breakage (see Section 4)

---

## 3. Interface Draft -- Conversational Path

### New Domain Types

```go
// --- discovery_values.go additions ---

const ModeConversational DiscoveryMode = "conversational"

// QuestionRef is a mode-agnostic reference to a question.
// Fixed mode: ID is "Q1"-"Q10". Conversational mode: generated IDs like "conv_actors_1".
type QuestionRef struct {
    id    string
    phase QuestionPhase  // still tracked for model coverage analysis
}

func NewQuestionRef(id string, phase QuestionPhase) QuestionRef {
    return QuestionRef{id: id, phase: phase}
}

func (r QuestionRef) ID() string          { return r.id }
func (r QuestionRef) Phase() QuestionPhase { return r.phase }
```

```go
// --- discovery_flow.go (new file) ---

// DiscoveryFlow encapsulates mode-specific question flow behavior.
type DiscoveryFlow interface {
    ValidateQuestionOrder(ref QuestionRef, answered []Answer, skipped map[string]bool) error
    IsPlaybackDue(answersSinceLastPlayback int) bool
    CheckCompleteness(answers []Answer, skipped map[string]bool) error
    PlaybackInterval() int
}

// FixedQuestionFlow implements the existing 10-question sequential flow.
type FixedQuestionFlow struct{}

func NewFixedQuestionFlow() *FixedQuestionFlow { return &FixedQuestionFlow{} }

func (f *FixedQuestionFlow) ValidateQuestionOrder(ref QuestionRef, answered []Answer, skipped map[string]bool) error {
    // Exact copy of current enforcePhaseOrder logic
    // Uses questionPhases var and QuestionCatalog()
}

func (f *FixedQuestionFlow) IsPlaybackDue(count int) bool {
    return count >= 3
}

func (f *FixedQuestionFlow) CheckCompleteness(answers []Answer, skipped map[string]bool) error {
    // Exact copy of current MVP question ID check
    answeredIDs := make(map[string]bool)
    for _, a := range answers {
        answeredIDs[a.QuestionID()] = true
    }
    mvpIDs := MVPQuestionIDs()
    var missing []string
    for id := range mvpIDs {
        if !answeredIDs[id] {
            missing = append(missing, id)
        }
    }
    if len(missing) > 0 {
        return fmt.Errorf("MVP questions not answered: %v: %w", missing, domainerrors.ErrInvariantViolation)
    }
    return nil
}

func (f *FixedQuestionFlow) PlaybackInterval() int { return 3 }

// ConversationalFlow implements adaptive LLM-driven discovery.
type ConversationalFlow struct {
    playbackInterval int
}

func NewConversationalFlow(playbackInterval int) *ConversationalFlow {
    if playbackInterval <= 0 {
        playbackInterval = 5 // default for conversational
    }
    return &ConversationalFlow{playbackInterval: playbackInterval}
}

func (f *ConversationalFlow) ValidateQuestionOrder(_ QuestionRef, _ []Answer, _ map[string]bool) error {
    return nil // LLM controls ordering
}

func (f *ConversationalFlow) IsPlaybackDue(count int) bool {
    return count >= f.playbackInterval
}

func (f *ConversationalFlow) CheckCompleteness(answers []Answer, _ map[string]bool) error {
    // Semantic coverage check: verify all phases have at least one answer
    phases := map[QuestionPhase]bool{
        PhaseActors:     false,
        PhaseStory:      false,
        PhaseEvents:     false,
        PhaseBoundaries: false,
    }
    for _, a := range answers {
        // QuestionRef carries phase info -- need to look it up
        // This requires answers to carry phase metadata (see QuestionRef)
    }
    // ... check all required phases covered
    return nil
}

func (f *ConversationalFlow) PlaybackInterval() int { return f.playbackInterval }
```

### Modified DiscoverySession API

```go
// AnswerQuestion now accepts any question ID (not just catalog IDs) when in conversational mode.
// The questionID must still be unique within the session (I6 preserved).
func (s *DiscoverySession) AnswerQuestion(questionID, response string) error {
    // ... existing state checks (I1, I4, I7 unchanged) ...

    // Mode-specific: validate question is known
    if s.Mode() == ModeConversational {
        // Accept any non-empty questionID -- the LLM generates them
    } else {
        qByID := QuestionByID()
        question, ok := qByID[questionID]
        if !ok {
            return fmt.Errorf("unknown question '%s'", questionID)
        }
        ref := NewQuestionRef(question.ID(), question.Phase())
        if err := s.flow.ValidateQuestionOrder(ref, s.answers, s.skipped); err != nil {
            return err
        }
    }

    // ... rest unchanged (record answer, check playback) ...
    if s.flow.IsPlaybackDue(s.answersSinceLastPlayback) {
        s.status = StatusPlaybackPending
    }
    return nil
}
```

### New Application Layer Port

```go
// --- ports.go additions ---

// ConversationalDiscovery extends Discovery with conversational flow methods.
// This is a separate interface (ISP) so existing consumers don't need to change.
type ConversationalDiscovery interface {
    // RegisterConversationalQuestion registers a dynamically generated question.
    // The question ID is generated by the LLM, and the phase indicates what domain
    // aspect this question targets (for model completeness checking).
    RegisterConversationalQuestion(sessionID string, questionID string, phase string, text string) error

    // GetModelGaps returns which domain model aspects are not yet covered.
    // Returns a list of phases/aspects that still need answers.
    GetModelGaps(sessionID string) ([]ModelGap, error)

    // SuggestNextQuestion asks the flow to suggest what to ask next.
    // In conversational mode, this triggers gap analysis and question generation.
    SuggestNextQuestion(sessionID string) (*SuggestedQuestion, error)
}
```

### New MCP Tools (Conversational)

```
guide_start              -- unchanged, but accepts mode="conversational"
guide_detect_persona     -- unchanged
guide_suggest_question   -- NEW: returns LLM-generated question based on gaps
guide_answer             -- unchanged API, but accepts dynamic question IDs
guide_confirm_playback   -- unchanged
guide_model_gaps         -- NEW: returns current model coverage gaps
guide_complete           -- unchanged API, but uses semantic completeness check
guide_status             -- unchanged, add gap coverage to output
```

---

## 4. Test Impact Assessment

### Test Inventory

| Test File | Test Count | Impact |
|-----------|-----------|--------|
| `discovery/domain/discovery_session_test.go` | 105 | **0 breakage** -- all tests use fixed-question flow (default) |
| `discovery/domain/question_test.go` | 11 | **0 breakage** -- question catalog unchanged |
| `discovery/domain/discovery_values_test.go` | 25 | **1 update** -- add `ModeConversational` to enum tests |
| `discovery/domain/classification_question_test.go` | 14 | **0 breakage** |
| `discovery/domain/discovery_events_test.go` | 6 | **0 breakage** |
| `discovery/application/discovery_handler_test.go` | 9 | **0 breakage** -- handler delegates to aggregate |
| `discovery/application/artifact_generation_handler_test.go` | 15 | **0 breakage** |
| `discovery/infrastructure/cli_discovery_adapter_test.go` | 13 | **0 breakage** -- adapter tests fixed-question CLI flow |
| `discovery/infrastructure/stdin_prompter_test.go` | 17 | **0 breakage** |
| `discovery/infrastructure/markdown_artifact_renderer_test.go` | 19 | **0 breakage** |
| `discovery/infrastructure/filesystem_tool_scanner_test.go` | 18 | **0 breakage** |
| `mcp/tools_discovery_test.go` | 31 | **0 breakage** -- all tests use default (express/fixed) flow |
| `integration/handler_flow_test.go` | 25 | **0 breakage** |
| `integration/cli_discovery_test.go` | 8 | **0 breakage** |
| `integration/classification_flow_test.go` | 1 | **0 breakage** |
| Other discovery domain tests | 45 | **0 breakage** |
| **Total** | **362** | **1 trivial update** |

### Why Zero Breakage

The key design decision is: **when no flow strategy is explicitly set, the aggregate defaults to `FixedQuestionFlow`**. This mirrors the existing pattern where `Mode()` defaults to `ModeExpress` when not set (L120-125):

```go
func (s *DiscoverySession) flow() DiscoveryFlow {
    if s.flowStrategy == nil {
        return &FixedQuestionFlow{} // default preserves all existing behavior
    }
    return s.flowStrategy
}
```

Every existing test creates sessions without setting conversational mode, so they all get `FixedQuestionFlow` automatically. The `enforcePhaseOrder` logic moves from the aggregate into `FixedQuestionFlow` but produces identical results.

### New Tests Required

| Category | Estimated Tests |
|----------|----------------|
| `ConversationalFlow` unit tests | ~20 (validation, playback cadence, completeness) |
| `FixedQuestionFlow` unit tests | ~10 (extracted from aggregate, verify identical behavior) |
| `DiscoverySession` + conversational mode integration | ~15 (state machine with dynamic questions) |
| `ConversationalDiscovery` handler tests | ~10 |
| MCP conversational tool tests | ~15 |
| **Total new tests** | **~70** |

---

## 5. Coupling Hotspot Resolutions

| # | Hotspot | Current Location | Resolution |
|---|---------|-----------------|------------|
| 1 | `questionPhases` var (hardcoded phase ordering) | `discovery_session.go:30` | Move into `FixedQuestionFlow`. The var becomes an implementation detail of the fixed strategy. `ConversationalFlow` does not reference it. |
| 2 | `QuestionCatalog()` singleton (CLI adapter iterates directly) | `question.go:89-93`, `cli_discovery_adapter.go:72` | **Keep for fixed mode.** CLI adapter already only runs in fixed-question mode. For conversational mode, the adapter will use `ConversationalDiscovery.SuggestNextQuestion()` instead. No change to existing code needed. |
| 3 | `playbackInterval` const (hardcoded to 3) | `discovery_session.go:43` | Move into `DiscoveryFlow.PlaybackInterval()`. Fixed flow returns 3. Conversational flow returns configurable value (default 5). The const becomes private to `FixedQuestionFlow`. |
| 4 | `MVPQuestionIDs` (hardcoded Q1,Q3,Q4,Q9,Q10) | `question.go:105-116` | Move into `FixedQuestionFlow.CheckCompleteness()`. Conversational flow uses semantic phase coverage check instead. The function remains exported for backward compatibility but is only used by `FixedQuestionFlow`. |
| 5 | Discovery port interface (methods assume fixed-question flow) | `application/ports.go:16-34` | **Keep unchanged.** The existing `Discovery` interface works for both modes -- `AnswerQuestion(sessionID, questionID, answer)` is mode-agnostic. Add `ConversationalDiscovery` as a separate interface (ISP) for conversational-specific methods (`SuggestNextQuestion`, `GetModelGaps`). |

---

## 6. Ubiquitous Language Note

### Problem

Ticket f0g.4 uses `GapAnalysis` in its UL section. This term already exists in the Rescue bounded context:

- `docs/DDD.md` L206: "Gap Analysis -- A scan of an existing project compared against a fully-seeded project to identify what's missing. Rescue context only"
- `internal/rescue/domain/gap_analysis.go`: `GapAnalysis` aggregate root already exists
- `docs/DDD.md` L544: `GapAnalysisCompleted` domain event already exists

Using `GapAnalysis` in Discovery would create cross-context UL ambiguity.

### Recommendation

**Use `ModelGap` (already documented in DDD.md L206)** for Discovery context:

> "Discovery uses `ModelGap` for missing model elements" -- `docs/DDD.md:206`

Proposed UL for Discovery's conversational flow:

| Term | Definition | Replaces |
|------|-----------|----------|
| `ModelGap` | A missing element in the domain model (e.g., no actors defined, no bounded contexts identified) | `GapAnalysis` (in Discovery context) |
| `ModelCompleteness` | The degree to which the domain model covers all required aspects (actors, stories, events, boundaries) | N/A (new concept) |
| `AdaptiveQuestion` | A question generated by the LLM based on current model gaps, adapted to project state | N/A (new concept) |
| `ConversationState` | **(AVOID as separate entity)** -- The DiscoverySession aggregate itself tracks conversation state via its flow strategy. No separate entity needed. | `ConversationState` (from f0g.4) |
| `DiscoveryFlow` | The strategy that governs question ordering, playback cadence, and completeness rules | N/A (new concept) |

**Rationale for avoiding `ConversationState` as a separate entity:** The DiscoverySession aggregate already manages all state transitions. Creating a parallel `ConversationState` entity would violate the "one aggregate per transaction" rule and create split-brain risk. Instead, the conversational behavior is encoded in the `ConversationalFlow` strategy, and the aggregate remains the single source of truth.

---

## 7. Follow-Up Ticket Updates for f0g.4

### Required Changes to alto-cli-f0g.4

**UL Section:**
- Replace `GapAnalysis` with `ModelGap` / `ModelCompleteness`
- Remove `ConversationState` as a separate entity -- use `DiscoverySession` with `ConversationalFlow` strategy
- Keep `AdaptiveQuestion` -- rename to `AdaptiveQuestion` (consistent with `DiscoveryFlow` UL)

**Design Section -- Replace "Steps" with:**

1. Add `ModeConversational` to `DiscoveryMode` enum
2. Create `DiscoveryFlow` interface in `domain/`
3. Extract existing logic into `FixedQuestionFlow` (pure refactor, zero behavior change)
4. Implement `ConversationalFlow` with configurable playback interval and semantic completeness
5. Add `QuestionRef` value object for mode-agnostic question references
6. Modify `DiscoverySession.AnswerQuestion()` to delegate to flow strategy
7. Modify `DiscoverySession.Complete()` to delegate completeness check to flow strategy
8. Add `ConversationalDiscovery` port interface (ISP -- separate from existing `Discovery`)
9. Implement `ConversationalDiscoveryHandler` in application layer
10. Add `guide_suggest_question` and `guide_model_gaps` MCP tools
11. Update `CLIDiscoveryAdapter` with `--conversational` flag (or make default based on LLM availability)

**TDD RED Phase -- Replace with:**

```go
// Phase 1: Extract strategy (pure refactor)
TestFixedQuestionFlow_EnforcesPhaseOrder
TestFixedQuestionFlow_PlaybackEvery3Answers
TestFixedQuestionFlow_ChecksMVPQuestions
TestDiscoverySession_DefaultFlowIsFixed  // ensures backward compat

// Phase 2: Conversational flow
TestConversationalFlow_AllowsAnyQuestionOrder
TestConversationalFlow_PlaybackEvery5Answers
TestConversationalFlow_SemanticCompletenessCheck
TestDiscoverySession_ConversationalMode_AcceptsDynamicQuestionIDs
TestDiscoverySession_ConversationalMode_PreservesUniversalInvariants

// Phase 3: Model gap analysis
TestModelGap_EmptySession_AllPhasesGapped
TestModelGap_PartialAnswers_ReportsRemainingGaps
TestModelCompleteness_AllPhasesHaveAnswers_IsComplete
```

**Acceptance Criteria -- Replace with:**

- [ ] `DiscoveryFlow` interface extracted; `FixedQuestionFlow` passes all 105 existing session tests
- [ ] `ConversationalFlow` implemented with configurable playback interval
- [ ] `DiscoverySession` delegates to flow strategy; zero existing test breakage
- [ ] `ModeConversational` added to enum; mode is set-once (I9 preserved)
- [ ] `ModelGap` and `ModelCompleteness` types replace `GapAnalysis` in Discovery UL
- [ ] `ConversationalDiscovery` port interface defined (ISP, does not modify `Discovery`)
- [ ] MCP tools `guide_suggest_question` and `guide_model_gaps` implemented
- [ ] All 12 invariants documented as preserved or adapted (per invariant map)
- [ ] golangci-lint + go test -race pass with 0 issues

---

## Summary of Recommendation

### Approach

**Strategy Pattern** -- Introduce `DiscoveryFlow` interface with `FixedQuestionFlow` and `ConversationalFlow` implementations. The `DiscoverySession` aggregate delegates mode-specific behavior (question ordering, playback cadence, completeness check) to its flow strategy.

### Key Numbers

| Metric | Value |
|--------|-------|
| Existing tests affected | 1 (trivial enum update) |
| Existing tests broken | 0 |
| New types/interfaces | 6 (`DiscoveryFlow`, `FixedQuestionFlow`, `ConversationalFlow`, `QuestionRef`, `ModelGap`, `ModelCompleteness`) |
| New port interfaces | 1 (`ConversationalDiscovery`) |
| New MCP tools | 2 (`guide_suggest_question`, `guide_model_gaps`) |
| Estimated new tests | ~70 |
| Invariants preserved unchanged | 9 of 12 |
| Invariants adapted (not removed) | 3 of 12 |

### Risk

The biggest risk is the **`Answer` value object's coupling to `QuestionCatalog()`**. Currently, `AnswerQuestion()` (L260-263) validates question IDs against the hardcoded catalog. In conversational mode, question IDs are LLM-generated. The refactoring must ensure that:
- Fixed mode continues to validate against the catalog (existing behavior)
- Conversational mode accepts any non-empty question ID
- The `Answer` VO itself does not change (it just stores `questionID` + `responseText`)

This is a clean seam -- the validation is in the aggregate, not the VO -- but it requires careful conditional logic to avoid breaking the fixed-mode validation.

### Implementation Order (3 phases)

1. **Pure refactor** (zero behavior change): Extract `FixedQuestionFlow` from existing aggregate logic. All 362 tests pass unchanged.
2. **Add conversational mode**: Implement `ConversationalFlow`, `ModeConversational`, `QuestionRef`. New tests only.
3. **Wire MCP tools**: Add `guide_suggest_question`, `guide_model_gaps`. This is where the LLM integration lives.

Phase 1 can be merged independently as a safe refactor. Phases 2-3 deliver the conversational feature.
