# Research: CLI Prompt Library for Interactive Discovery Flow

**Date:** 2026-03-11
**Status:** Final

## Summary

Evaluated three options for implementing alto's interactive CLI prompts (`alto guide`): charmbracelet/huh, charmbracelet/bubbles (raw Bubble Tea), and raw stdin with bufio.Scanner. **Recommendation: charmbracelet/huh v2** -- it provides the exact prompt primitives alto needs (Select, Input, Text, Confirm) with built-in accessibility mode, minimal integration surface, and wraps cleanly behind a port interface.

## Research Question

Which CLI prompt library best fits alto's interactive discovery flow requirements?

- Persona selection (choose 1-4)
- Question display with multi-line text input
- Free text answer input
- Skip with reason prompt
- Playback confirmation (y/n/edit)
- Ctrl+C graceful exit
- Cross-platform: Linux, macOS, Windows
- Wrappable behind a port interface (DDD boundary requirement)

## Options Considered

### Option 1: charmbracelet/huh v2

| Attribute | Detail | Source |
|-----------|--------|--------|
| **Version** | v2.0.3 (stable) | [GitHub releases](https://github.com/charmbracelet/huh/releases) |
| **Release date** | 2026-03-10 | [GitHub releases](https://github.com/charmbracelet/huh/releases) |
| **License** | MIT | [pkg.go.dev](https://pkg.go.dev/github.com/charmbracelet/huh/v2) |
| **Go version** | Requires Go 1.23+ | [go.mod](https://github.com/charmbracelet/huh/blob/main/go.mod) |
| **CGO** | None -- pure Go | No CGO imports in dependency tree |
| **Direct deps** | 13 external packages | [pkg.go.dev imports](https://pkg.go.dev/github.com/charmbracelet/huh/v2?tab=imports) |
| **Transitive deps** | ~27 total (charmbracelet/x, lipgloss, bubbletea, bubbles) | pkg.go.dev imports tab |
| **Cross-platform** | Linux, macOS, Windows, js/wasm | [bubbletea platform files](https://github.com/charmbracelet/bubbletea) |
| **Accessibility** | Built-in `WithAccessible(true)` mode for screen readers | [Context7 docs](/charmbracelet/huh), [Issue #611](https://github.com/charmbracelet/huh/issues/611) |
| **Ctrl+C** | Returns `huh.ErrUserAborted` -- sentinel error | [Context7 docs](/charmbracelet/huh) |

**API mapping to alto use cases:**

| Use Case | huh Component | Code Pattern |
|----------|---------------|-------------|
| Persona selection (1-4) | `huh.NewSelect[string]()` | `.Title("Which describes you?").Options(...)` |
| Question display + answer | `huh.NewText()` | `.Title(question).Lines(6).Value(&answer)` |
| Free text (single line) | `huh.NewInput()` | `.Title("...").Value(&answer)` |
| Skip reason | `huh.NewInput()` | `.Title("Reason for skipping?").Value(&reason)` |
| Playback confirm | `huh.NewConfirm()` | `.Title("Is this correct?").Value(&confirmed)` |
| Edit on rejection | `huh.NewText()` | `.Title("Corrections:").Value(&corrections)` |

**Key strengths:**
- Form-level abstraction: group fields, auto-navigate between groups
- Built-in validation via `.Validate(func(string) error)`
- Declarative API: bind to Go variables with `.Value(&v)`
- Generic types: `Select[T]` means we can bind to domain types directly
- Accessible mode drops TUI in favor of simple stdin prompts -- critical for CI/testing

### Option 2: charmbracelet/bubbles (raw Bubble Tea)

| Attribute | Detail | Source |
|-----------|--------|--------|
| **Version** | v2.0.2 (bubbletea), v2.x (bubbles) | [pkg.go.dev](https://pkg.go.dev/charm.land/bubbletea/v2) |
| **Release date** | 2026-03-06 (bubbletea v2.0.2) | [pkg.go.dev](https://pkg.go.dev/charm.land/bubbletea/v2) |
| **License** | MIT | [GitHub](https://github.com/charmbracelet/bubbletea) |
| **Go version** | Go 1.23+ | go.mod |
| **CGO** | None -- pure Go | No CGO imports |
| **Direct deps** | 26 imports (bubbletea alone) | [pkg.go.dev](https://pkg.go.dev/charm.land/bubbletea/v2) |
| **Cross-platform** | Linux, macOS, Windows, js/wasm | Platform-specific files in repo |
| **Accessibility** | None built-in -- must implement manually | No accessibility package |
| **Ctrl+C** | Manual: check `tea.KeyPressMsg` for "ctrl+c" and return `tea.Quit` | [Context7 docs](/charmbracelet/bubbletea) |

**API mapping to alto use cases:**

Each prompt requires implementing the full Model-View-Update (MVU) pattern:
- `Init() tea.Cmd`
- `Update(msg tea.Msg) (tea.Model, tea.Cmd)`
- `View() tea.View`

For persona selection alone, you'd need: a model struct, key handling for arrow keys + enter, rendering logic for highlighted/selected items, and a quit handler. Roughly 60-80 lines per prompt type vs 5-10 lines with huh.

**Key strengths:**
- Maximum flexibility for custom UI (progress bars, animations, live updates)
- Same dependency tree as huh (huh is built on bubbles/bubbletea)

**Key weaknesses:**
- 5-10x more code per prompt compared to huh
- No form grouping -- must wire state machine manually
- No built-in validation feedback
- No accessibility mode -- would need custom implementation
- MVU architecture is overkill for sequential form prompts

### Option 3: Raw stdin (bufio.Scanner + fmt.Print)

| Attribute | Detail | Source |
|-----------|--------|--------|
| **Version** | stdlib (Go 1.26) | N/A |
| **License** | BSD-3-Clause (Go stdlib) | N/A |
| **Go version** | Any | stdlib |
| **CGO** | None | stdlib |
| **Dependencies** | Zero | N/A |
| **Cross-platform** | Full Go platform support | stdlib |
| **Accessibility** | Inherently accessible (plain text I/O) | N/A |
| **Ctrl+C** | Must handle `os.Signal` manually | N/A |

**API mapping to alto use cases:**

| Use Case | Implementation |
|----------|---------------|
| Persona selection | Print numbered list, read line, validate "1"-"4" |
| Multi-line text | Read lines until sentinel (e.g., blank line or Ctrl+D) |
| Free text | `scanner.Scan()`, trim, validate non-empty |
| Skip reason | Same as free text |
| Playback confirm | Read "y"/"n"/"edit", validate |

**Key strengths:**
- Zero dependencies
- Trivial to test (inject `io.Reader`/`io.Writer`)
- Inherently accessible (no TUI escape sequences)
- Maximally portable

**Key weaknesses:**
- No arrow-key navigation for selection (type number only)
- No inline validation feedback (validate after submit)
- Multi-line input UX is poor (no editing previous lines)
- No styled output (bold questions, colored prompts) without manual ANSI
- Must build every UX affordance from scratch
- Ctrl+C handling requires signal goroutine + context cancellation wiring
- **Product Owner / Domain Expert persona expects a polished experience** (PRD Scenario 4-5)

## Comparison Matrix

| Criteria | huh v2 | bubbles (raw) | Raw stdin |
|----------|--------|---------------|-----------|
| **Lines of code per prompt** | 5-10 | 60-80 | 20-40 |
| **Accessibility** | Built-in toggle | Manual build | Inherent |
| **Cross-platform** | Yes | Yes | Yes |
| **Ctrl+C handling** | Sentinel error | Manual | Manual |
| **Validation feedback** | Inline | Manual | Post-submit |
| **Multi-line editing** | Yes (textarea) | Yes (textarea) | Poor |
| **Arrow-key selection** | Yes | Yes (manual) | No |
| **Dependency count** | ~27 transitive | ~26 transitive | 0 |
| **Testability** | `WithAccessible(true)` -> stdin/stdout | Mock tea.Program | Inject Reader/Writer |
| **Port wrappability** | Easy (call huh behind interface) | Easy (same) | Easy |
| **UX polish** | High (themes, animation) | High (custom) | Low |
| **Maintenance burden** | Low (Charm maintains) | Medium (we maintain glue) | High (we maintain everything) |
| **License** | MIT | MIT | BSD-3 (stdlib) |
| **CGO** | No | No | No |

## Architecture Integration

All three options can be wrapped behind a port interface. The existing `Discovery` port in `internal/discovery/application/ports.go` manages the domain session state machine. The CLI prompt layer would be a separate concern -- a `Prompter` port in the CLI or composition layer:

```go
// Prompter defines the UI contract for interactive discovery prompts.
// Infrastructure: huh adapter, stdin adapter, or test fake.
type Prompter interface {
    SelectPersona(ctx context.Context, choices []string) (string, error)
    AskQuestion(ctx context.Context, question string) (string, error)
    AskSkipReason(ctx context.Context) (string, error)
    ConfirmPlayback(ctx context.Context, summary string) (bool, string, error)
}
```

This port keeps the domain clean regardless of which library implements the prompts. The huh adapter would be ~50 lines total. A test fake would be ~20 lines (return canned answers).

## Decision Drivers (from PRD)

1. **Non-technical users** (PRD Personas: Product Owner, Domain Expert) -- needs polished UX, not raw terminal prompts. Eliminates raw stdin as primary option.
2. **Cross-platform** (PRD: "works with Claude Code, Cursor, Roo Code, OpenCode") -- all three options satisfy this. No differentiator.
3. **Accessibility** -- huh has first-class support. Raw stdin is inherently accessible. Bubbles requires custom work.
4. **Dependency budget** -- alto already depends on Cobra (which is lightweight). Adding ~27 transitive deps from the Charm ecosystem is a real cost, but these are all pure Go, well-maintained, MIT-licensed packages from a reputable organization (Charm, 28k+ GitHub stars on bubbletea).
5. **Maintainability** -- huh's declarative API means less custom code to maintain. Bubbles' MVU pattern means more custom code. Raw stdin means the most custom code.

## Recommendation

**Use charmbracelet/huh v2** (`github.com/charmbracelet/huh/v2` at v2.0.3).

**Rationale:**

1. **Direct API match**: Every alto prompt type (Select, Input, Text, Confirm) maps 1:1 to a huh component. No glue code needed.
2. **Accessibility toggle**: `WithAccessible(true)` gives free screen reader support and is also useful for testing (no TUI, just stdin/stdout).
3. **Ctrl+C is a sentinel error**: `huh.ErrUserAborted` integrates cleanly with Go error handling -- no signal goroutine needed.
4. **Pure Go, MIT, actively maintained**: v2.0.3 released 2026-03-10. No CGO. Compatible with Go 1.23+ (alto uses 1.26).
5. **Port-friendly**: Wraps trivially behind a `Prompter` interface, keeping domain layer clean.

**Tradeoff acknowledged**: ~27 transitive dependencies is non-trivial. However, these are all from the Charm ecosystem (well-maintained, pure Go, MIT), and the alternative is writing and maintaining ~500+ lines of custom prompt code (bubbles) or ~300+ lines with poor UX (raw stdin).

**Fallback strategy**: If huh proves problematic (e.g., terminal compatibility issue in a specific environment), the `Prompter` port interface means we can swap to a raw stdin adapter without changing any domain or application code. Consider shipping both adapters: huh as default, stdin as `--accessible` or `--no-tui` flag.

## References

- [charmbracelet/huh GitHub](https://github.com/charmbracelet/huh) -- Source repository
- [charmbracelet/huh releases](https://github.com/charmbracelet/huh/releases) -- v2.0.3 released 2026-03-10
- [huh v2 pkg.go.dev](https://pkg.go.dev/github.com/charmbracelet/huh/v2) -- API docs, import list
- [charmbracelet/bubbletea GitHub](https://github.com/charmbracelet/bubbletea) -- Underlying TUI framework
- [bubbletea v2 pkg.go.dev](https://pkg.go.dev/charm.land/bubbletea/v2) -- v2.0.2, MIT, Go 1.23+
- [huh accessibility issue #611](https://github.com/charmbracelet/huh/issues/611) -- Screen reader improvements
- [huh accessibility PR #620](https://github.com/charmbracelet/huh/pull/620) -- Prompt improvements for accessible mode
- [alto Discovery ports](file:///home/kusanagi/Alto/alto-cli/internal/discovery/application/ports.go) -- Existing port interfaces
- [alto PRD](file:///home/kusanagi/Alto/alto-cli/docs/PRD.md) -- Persona and UX requirements

## Follow-up Tasks

- [ ] Task 1: Define `Prompter` port interface in `internal/discovery/application/ports.go` (or CLI-level composition)
- [ ] Task 2: Implement huh v2 adapter for `Prompter` in infrastructure layer
- [ ] Task 3: Implement raw stdin adapter for `Prompter` (fallback / `--no-tui` flag)
- [ ] Task 4: Add `--accessible` / `--no-tui` CLI flag to `alto guide` command
- [ ] Task 5: Wire huh v2 dependency: `go get github.com/charmbracelet/huh/v2@v2.0.3`
