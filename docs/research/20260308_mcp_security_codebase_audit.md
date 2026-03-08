# Security Audit: MCP Server Pre-Implementation Codebase Review

**Date:** 2026-03-08
**Auditor:** QA Engineer (security-focused review)
**Scope:** Existing Go codebase (`internal/`, `cmd/`) ahead of MCP server implementation (Epic 3, alty-0m9)
**Method:** Static code analysis of all Go source files, dependency scanning via `govulncheck`

---

## Executive Summary

**Overall Risk Level: Medium**

The codebase demonstrates good security hygiene in several areas: session IDs use `crypto/rand` (UUID v4), git branch names are validated with a strict regex, subprocess commands use argument arrays (not shell strings), and the `KnowledgePath` value object blocks `..` path traversal. However, there are **10 findings** that would become exploitable attack surfaces once the MCP server exposes these handlers over JSON-RPC. The most critical issues involve path traversal through `FilesystemFileWriter`, missing session cleanup, and an unsanitized ticket ID passed to `exec.Command`.

**Summary by Severity:**

| Severity | Count |
|----------|-------|
| High     | 3     |
| Medium   | 5     |
| Low      | 2     |

---

## Findings

### [F1] HIGH: Path Traversal in FilesystemFileWriter -- No Path Sanitization

**Affected Files:**
- `/Users/jaigoukkim/Alty/alty-cli/internal/shared/infrastructure/persistence/filesystem_file_writer.go`
- `/Users/jaigoukkim/Alty/alty-cli/internal/tooltranslation/application/persona_handler.go` (line 79)
- `/Users/jaigoukkim/Alty/alty-cli/internal/tooltranslation/application/config_generation_handler.go` (line 94)
- `/Users/jaigoukkim/Alty/alty-cli/internal/rescue/application/rescue_handler.go` (line 161)
- `/Users/jaigoukkim/Alty/alty-cli/internal/ticket/application/ticket_generation_handler.go` (line 78)
- `/Users/jaigoukkim/Alty/alty-cli/internal/discovery/application/artifact_generation_handler.go` (line 98)

**Description:**
`FilesystemFileWriter.WriteFile()` accepts an arbitrary path and writes content to it with `os.MkdirAll` + `os.WriteFile`. There is no validation that the resolved path stays within a project directory boundary. The `filepath.Join(outputDir, subPath)` pattern used by all callers is vulnerable to traversal when `subPath` contains `..` segments or absolute paths.

For example, in `config_generation_handler.go`:
```go
target := filepath.Join(outputDir, section.FilePath())
```
If `section.FilePath()` is controlled by an attacker through MCP tool input (e.g., a crafted DomainModel with a malicious bounded context name that flows into file path generation), the write could escape the project directory.

Currently, only `KnowledgePath` (in `internal/knowledge/domain/knowledge_entry.go`) validates against `..`:
```go
if strings.Contains(raw, "..") {
    return KnowledgePath{}, fmt.Errorf("knowledge path must not contain path traversal (..): %w", ...)
}
```

No other file path input has this protection.

**MCP Exploitation Vector:** An MCP client sends a `generate_config` or `write_persona` tool call with a crafted `outputDir` like `/tmp/project/../../../etc/`. The server would write files outside the project boundary.

**Recommended Fix:**
1. Create a `SafeProjectPath(projectRoot, subPath string) (string, error)` utility in the shared domain that:
   - Calls `filepath.Clean()` on the joined path
   - Resolves symlinks with `filepath.EvalSymlinks()`
   - Verifies the resolved path has the project root as a prefix
   - Rejects absolute sub-paths
2. Use `SafeProjectPath` in `FilesystemFileWriter.WriteFile()` or at every callsite
3. Add the project root as a field on `FilesystemFileWriter` so it can enforce boundaries

---

### [F2] HIGH: Unsanitized ticketID Passed to exec.Command

**Affected File:**
- `/Users/jaigoukkim/Alty/alty-cli/internal/ticket/infrastructure/beads_ticket_reader.go` (line 189)

**Description:**
The `readFlagsFromBDComments` method passes `ticketID` directly to `exec.CommandContext`:
```go
cmd := exec.CommandContext(ctx, "bd", "comments", ticketID)
```

While Go's `exec.Command` does not invoke a shell (so classic shell injection via `;` or `$()` is not possible), the `ticketID` is not validated against any format. An attacker-controlled ticket ID could:
- Pass flags to the `bd` command (e.g., `--help`, `--output=/some/file`) if `bd` interprets arguments starting with `-`
- Cause unexpected behavior with excessively long strings
- Exploit any argument parsing vulnerabilities in the `bd` binary

The `getFlaggedIDs` method at line 157 has a similar pattern but uses a hardcoded query string (`"label=review_needed"`), which is safe.

**MCP Exploitation Vector:** An MCP client calls `ticket_health` with a ticket ID like `--output=/tmp/exfil` that gets passed to the `bd` subprocess.

**Recommended Fix:**
1. Validate `ticketID` matches the expected format (e.g., alphanumeric + hyphens, similar to `branchNamePattern` in git_ops_adapter.go)
2. Use `--` separator before positional arguments: `exec.CommandContext(ctx, "bd", "comments", "--", ticketID)`
3. Apply the same validation pattern to all ticket ID inputs at the domain boundary

---

### [F3] HIGH: Session Memory Leak -- No Periodic Cleanup Goroutine

**Affected Files:**
- `/Users/jaigoukkim/Alty/alty-cli/internal/discovery/application/discovery_handler.go` (line 13)
- `/Users/jaigoukkim/Alty/alty-cli/internal/bootstrap/application/bootstrap_handler.go` (line 42)
- `/Users/jaigoukkim/Alty/alty-cli/internal/shared/infrastructure/persistence/session_store.go`

**Description:**
Three separate session stores exist in the codebase:
1. `DiscoveryHandler.sessions` -- plain `map[string]*DiscoverySession` with mutex, **no TTL, no cleanup**
2. `BootstrapHandler.sessions` -- plain `map[string]*BootstrapSession` with mutex, **no TTL, no cleanup**
3. `SessionStore` -- has TTL and `CleanupExpired()`, but is **never instantiated** in `composition/app.go`

The `SessionStore` with TTL-based expiration was built (`internal/shared/infrastructure/persistence/session_store.go`) but neither `DiscoveryHandler` nor `BootstrapHandler` uses it. Both handlers maintain their own plain maps that grow unboundedly.

In CLI mode this is harmless (process exits). In MCP server mode (long-running process), every `StartSession` or `Preview` call adds an entry that is never removed. A malicious or buggy MCP client could exhaust server memory by creating thousands of sessions.

**MCP Exploitation Vector:** An MCP client repeatedly calls `start_discovery_session` without completing or cancelling sessions. Each abandoned session holds the full README content, all answers, and all playback data in memory indefinitely.

**Recommended Fix:**
1. Migrate `DiscoveryHandler` and `BootstrapHandler` to use `SessionStore` instead of plain maps
2. Start a background goroutine in `NewApp()` or the MCP server that calls `CleanupExpired()` periodically (e.g., every 5 minutes)
3. Add a max active sessions limit (e.g., 100) with a clear error when exceeded
4. Sessions in terminal states (completed, cancelled) should be removed immediately

---

### [F4] MEDIUM: TOCTOU Race in BootstrapHandler -- Session Mutation Outside Lock

**Affected Files:**
- `/Users/jaigoukkim/Alty/alty-cli/internal/bootstrap/application/bootstrap_handler.go` (lines 90-92, 136-138)
- `/Users/jaigoukkim/Alty/alty-cli/internal/discovery/application/discovery_handler.go` (lines 26-28, 93-95)

**Description:**
Both handlers use a "lock-get-unlock-mutate" pattern:

```go
// bootstrap_handler.go:135-142
func (h *BootstrapHandler) getSession(sessionID string) (*domain.BootstrapSession, error) {
    h.mu.Lock()
    session, ok := h.sessions[sessionID]
    h.mu.Unlock()               // <-- unlocked here
    if !ok {
        return nil, fmt.Errorf(...)
    }
    return session, nil          // <-- returned pointer, no lock held
}
```

The session pointer is returned while the mutex is unlocked. Two concurrent MCP requests operating on the same session could:
1. Both call `getSession()` and get the same pointer
2. Both call state-transition methods on the `BootstrapSession`
3. Race on the session's internal state fields (status, preview, etc.)

The domain aggregates (`BootstrapSession`, `DiscoverySession`) are not thread-safe -- they are pure domain objects that assume single-threaded access.

**MCP Exploitation Vector:** Two concurrent MCP tool calls reference the same session ID. One calls `Confirm()` while another calls `Cancel()`, causing the session to enter an inconsistent state. With Go's race detector (`-race`), this would be flagged as a data race.

**Recommended Fix:**
1. Either hold the lock for the entire operation (not just the map lookup), or
2. Copy the session data under the lock and operate on the copy, or
3. Use a per-session mutex (recommended for MCP where concurrent access to the same session is expected), or
4. Document that sessions are single-writer and enforce this at the MCP transport layer (serialize requests per session)

---

### [F5] MEDIUM: Error Messages Leak Internal File Paths

**Affected Files (representative samples):**
- `/Users/jaigoukkim/Alty/alty-cli/internal/shared/infrastructure/persistence/filesystem_file_writer.go` (lines 32, 35)
- `/Users/jaigoukkim/Alty/alty-cli/internal/knowledge/infrastructure/file_knowledge_reader.go` (line 42)
- `/Users/jaigoukkim/Alty/alty-cli/internal/fitness/infrastructure/subprocess_gate_runner.go` (line 87-94)

**Description:**
Many error messages include full filesystem paths:
```go
// filesystem_file_writer.go
return fmt.Errorf("creating directory %s: %w", dir, err)
return fmt.Errorf("writing file %s: %w", path, err)

// file_knowledge_reader.go
return ..., fmt.Errorf("knowledge entry not found: %s (looked at %s): %w", path.Raw(), filePath, ...)
```

The subprocess gate runner returns raw command output on failure:
```go
return vo.NewGateResult(gate, false, string(output), durationMS), nil
```

In CLI mode these paths help debugging. In MCP mode, they would be returned in JSON-RPC responses, leaking:
- The server's absolute filesystem layout
- Presence/absence of specific files
- Raw subprocess stderr output (which may contain environment variables, paths, or stack traces)

**MCP Exploitation Vector:** An MCP client sends requests with invalid paths to enumerate the server's directory structure through error messages.

**Recommended Fix:**
1. Create an error sanitization layer at the MCP transport boundary that strips absolute paths from error messages
2. For infrastructure errors, wrap with a generic message: "file operation failed" rather than including the full path
3. Log the detailed error server-side but return a sanitized version to the client
4. For subprocess output, truncate and sanitize before including in responses

---

### [F6] MEDIUM: filepath.Walk Follows Symlinks -- Potential Information Disclosure

**Affected File:**
- `/Users/jaigoukkim/Alty/alty-cli/internal/dochealth/infrastructure/filesystem_doc_scanner.go` (line 118)

**Description:**
`ScanUnregistered` uses `filepath.Walk()` to traverse the docs directory:
```go
err := filepath.Walk(docsDir, func(path string, info os.FileInfo, err error) error {
```

Go's `filepath.Walk` follows symlinks. If an attacker can create a symlink inside the project's `docs/` directory (e.g., `docs/link -> /etc/`), the scanner would:
1. Follow the symlink and traverse outside the project
2. Read and process files from arbitrary locations
3. Include their content in doc health reports returned via MCP

Additionally, `os.Stat` (used pervasively) follows symlinks by default. No code in the codebase uses `os.Lstat` to detect symlinks.

**MCP Exploitation Vector:** An attacker with write access to the project directory creates `docs/secret -> /etc/shadow`, then triggers doc-health via MCP. The scanner reads the symlink target and reports its content.

**Recommended Fix:**
1. Use `filepath.WalkDir` with a symlink check: skip entries where `d.Type()&fs.ModeSymlink != 0`
2. Or resolve symlinks and verify the resolved path stays within the project root
3. Consider using `os.Lstat` instead of `os.Stat` in security-sensitive checks

---

### [F7] MEDIUM: No Input Size Limits on File Reads

**Affected Files:**
- `/Users/jaigoukkim/Alty/alty-cli/internal/dochealth/infrastructure/filesystem_doc_scanner.go` (lines 37, 79, 165) -- `os.ReadFile`
- `/Users/jaigoukkim/Alty/alty-cli/internal/knowledge/infrastructure/file_knowledge_reader.go` (lines 119, 140) -- `os.ReadFile`
- `/Users/jaigoukkim/Alty/alty-cli/internal/research/infrastructure/spike_follow_up_adapter.go` (line 125) -- `os.ReadFile`
- `/Users/jaigoukkim/Alty/alty-cli/internal/research/infrastructure/markdown_spike_parser.go` (line 53) -- `os.ReadFile`
- `/Users/jaigoukkim/Alty/alty-cli/internal/knowledge/infrastructure/knowledge_drift_detector.go` (line 262) -- `os.ReadFile`
- `/Users/jaigoukkim/Alty/alty-cli/internal/ticket/infrastructure/beads_ticket_reader.go` (lines 83, 116) -- `bufio.NewScanner` (default 64KB buffer, but no line count limit)

**Description:**
All file reads use `os.ReadFile()` which loads the entire file into memory with no size limit. The `bufio.Scanner` in `BeadsTicketReader` uses the default 64KB line buffer but processes an unbounded number of lines.

In MCP mode, if a malicious project contains a multi-gigabyte `docs/PRD.md` or `issues.jsonl`, scanning it would exhaust server memory.

**MCP Exploitation Vector:** An attacker creates a 10GB `.beads/issues.jsonl` file, then triggers ticket health scanning via MCP.

**Recommended Fix:**
1. Add file size checks before `os.ReadFile` calls: `info, _ := os.Stat(path); if info.Size() > maxFileSize { return error }`
2. Use `io.LimitReader` for streaming reads
3. Set reasonable limits (e.g., 10MB for markdown files, 50MB for JSONL)

---

### [F8] MEDIUM: No Rate Limiting or Request Throttling Infrastructure

**Affected Files:**
- All handler files in `internal/*/application/`

**Description:**
No handler implements any form of rate limiting, request throttling, or concurrent request limiting. The MCP server will expose all handlers as tool endpoints. Without rate limiting:
- A client could flood `SubprocessGateRunner.Run()` with quality gate requests, spawning hundreds of concurrent subprocesses
- A client could create thousands of discovery sessions (see F3)
- File I/O operations could saturate disk throughput

**MCP Exploitation Vector:** An MCP client sends 1000 concurrent `check --gate=tests` requests, spawning 1000 `uv run pytest` subprocesses simultaneously, exhausting system resources.

**Recommended Fix:**
1. Implement a semaphore/token bucket at the MCP transport layer limiting concurrent tool invocations
2. Limit concurrent subprocess executions (e.g., max 3 simultaneous quality gate runs)
3. Add per-client rate limiting (e.g., max 10 requests/second)

---

### [F9] LOW: Session ID Exposed in Error Messages

**Affected Files:**
- `/Users/jaigoukkim/Alty/alty-cli/internal/discovery/application/discovery_handler.go` (line 97)
- `/Users/jaigoukkim/Alty/alty-cli/internal/bootstrap/application/bootstrap_handler.go` (line 140)
- `/Users/jaigoukkim/Alty/alty-cli/internal/shared/infrastructure/persistence/session_store.go` (lines 58, 63)

**Description:**
Error messages include the session ID:
```go
return nil, fmt.Errorf("no active discovery session with id '%s'", sessionID)
return nil, fmt.Errorf("session '%s' has expired: %w", sessionID, domainerrors.ErrNotFound)
```

While session IDs are UUID v4 (unpredictable, from `crypto/rand`), confirming whether a specific session ID exists or has expired leaks information. An attacker could distinguish between "never existed" and "expired" states.

The session ID generation itself is sound:
```go
// internal/shared/domain/identity/id.go
func NewID() string {
    var b [16]byte
    _, _ = rand.Read(b[:])  // crypto/rand -- good
    ...
}
```

**MCP Exploitation Vector:** An attacker brute-forces session IDs (impractical given UUID v4 entropy) or uses a leaked session ID to probe session state.

**Recommended Fix:**
1. Return the same generic error for both "not found" and "expired": `"session not found"`
2. Keep the distinction in server-side logs

---

### [F10] LOW: SubprocessGateRunner Commands Are Not Validated

**Affected File:**
- `/Users/jaigoukkim/Alty/alty-cli/internal/fitness/infrastructure/subprocess_gate_runner.go` (lines 42-45, 74)

**Description:**
`NewSubprocessGateRunner` accepts a `StackProfile` and uses its `QualityGateCommands()` to determine what subprocesses to run:
```go
func NewSubprocessGateRunner(projectDir string, profile vo.StackProfile) *SubprocessGateRunner {
    ...
    return &SubprocessGateRunner{
        projectDir: projectDir,
        commands:   profile.QualityGateCommands(),  // commands come from profile
    }
}
```

Currently, profiles are hardcoded (`PythonUvProfile`, `GenericProfile`). However, if a future stack profile is loaded from user-provided configuration (e.g., `.alty/config.toml`), an attacker could inject arbitrary commands:
```toml
[quality_gates]
lint = ["rm", "-rf", "/"]
```

**MCP Exploitation Vector:** Not directly exploitable today, but becomes a risk if stack profiles are ever loaded from project configuration files accessible to MCP clients.

**Recommended Fix:**
1. Validate that gate commands are from an allow-list of known executables
2. Document that `StackProfile` implementations must never load commands from user input
3. When implementing config-file-based profiles, validate against an allow-list

---

## Positive Findings (Good Patterns to Preserve)

These patterns are correctly implemented and should be maintained during MCP server development:

| Pattern | Location | Assessment |
|---------|----------|------------|
| **UUID v4 session IDs** from `crypto/rand` | `internal/shared/domain/identity/id.go` | Unpredictable, no session fixation risk |
| **Branch name validation** with strict regex | `internal/rescue/infrastructure/git_ops_adapter.go:14` | `^[a-zA-Z0-9/_-]+$` blocks injection |
| **No shell invocation** in subprocess execution | All `exec.CommandContext` calls | Arguments passed as arrays, not shell strings |
| **Context timeouts** on subprocess execution | `subprocess_gate_runner.go:71`, `beads_ticket_reader.go:154,186` | 10s-300s timeouts prevent hangs |
| **Path traversal check** in KnowledgePath | `internal/knowledge/domain/knowledge_entry.go:61` | Blocks `..` in knowledge paths |
| **Defensive copies** on all slice/map returns | All domain aggregate getters | Prevents caller mutation of internal state |
| **Compile-time interface checks** | All infrastructure adapters | `var _ Port = (*Adapter)(nil)` pattern |
| **Watermill event bus** with typed deserialization | `internal/shared/infrastructure/eventbus/subscriber.go` | Type-safe event dispatch |
| **No known dependency vulnerabilities** | `go.mod` dependencies | `govulncheck ./...` returned zero findings |

---

## Dependency Vulnerability Scan

```
$ govulncheck ./...
No vulnerabilities found.
```

All direct and indirect dependencies are clean as of 2026-03-08. Dependencies are minimal and well-maintained:
- `github.com/spf13/cobra v1.10.2` (CLI)
- `github.com/ThreeDotsLabs/watermill v1.5.1` (event bus)
- `github.com/BurntSushi/toml v1.6.0` (TOML parsing)
- `github.com/stretchr/testify v1.11.1` (testing)

---

## Recommendations for MCP Server Implementation

### Priority 1 (Must-Have Before First MCP Release)

1. **Implement `SafeProjectPath` utility** -- shared kernel function that validates all file paths stay within the project root. Wire it into `FilesystemFileWriter` or add it as middleware. Blocks F1.

2. **Migrate session management** -- Replace plain maps in `DiscoveryHandler` and `BootstrapHandler` with `SessionStore`. Add periodic cleanup goroutine and max session limits. Blocks F3.

3. **Validate ticket IDs** at the domain boundary with a format regex (e.g., `^[a-z0-9-]+$`). Blocks F2.

4. **Add an error sanitization layer** at the MCP JSON-RPC boundary. Map infrastructure errors to generic messages for clients. Log full details server-side. Blocks F5.

### Priority 2 (Should-Have for Production)

5. **Add request rate limiting** at the MCP transport layer. Use a semaphore for concurrent subprocess execution. Blocks F8.

6. **Add file size limits** before all `os.ReadFile` calls. Blocks F7.

7. **Replace `filepath.Walk` with symlink-safe traversal**. Blocks F6.

8. **Fix TOCTOU in session handlers** -- either hold the lock during the full operation or add per-session mutexes. Blocks F4.

### Priority 3 (Nice-to-Have)

9. **Unify session ID error messages** to not distinguish "not found" from "expired". Blocks F9.

10. **Document command allow-list policy** for StackProfile implementations. Blocks F10.

---

## Testing Recommendations

The following tests should be added before the MCP server ships:

```go
// Path traversal tests
func TestSafeProjectPath_RejectsDotDot(t *testing.T)
func TestSafeProjectPath_RejectsAbsolutePath(t *testing.T)
func TestSafeProjectPath_RejectsSymlinkEscape(t *testing.T)
func TestSafeProjectPath_AcceptsValidSubpath(t *testing.T)

// Session management tests
func TestDiscoveryHandler_SessionCleanup_RemovesExpired(t *testing.T)
func TestBootstrapHandler_SessionCleanup_RemovesExpired(t *testing.T)
func TestSessionStore_MaxSessions_RejectsOverLimit(t *testing.T)

// Input validation tests
func TestBeadsTicketReader_RejectsInvalidTicketID(t *testing.T)
func TestBeadsTicketReader_TicketIDWithDashes_Accepted(t *testing.T)

// Concurrency tests
func TestDiscoveryHandler_ConcurrentAccess_NoRace(t *testing.T)
func TestBootstrapHandler_ConcurrentSessionCreation_NoRace(t *testing.T)

// Error sanitization tests
func TestErrorSanitizer_StripsAbsolutePaths(t *testing.T)
func TestErrorSanitizer_PreservesDomainErrors(t *testing.T)
```

All existing tests continue to pass (verified via codebase analysis of test patterns).
