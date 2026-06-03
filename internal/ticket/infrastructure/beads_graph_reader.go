package infrastructure

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/alto-cli/alto/internal/ticket/application"
)

// BeadsGraphReaderTimeout is the per-invocation deadline for every `bd`
// call this adapter issues — matches the convention enforced by
// BeadsLabelWriter at beads_label_writer.go:10.
const BeadsGraphReaderTimeout = 5 * time.Second

// BDCommandRunner is the seam the adapter uses to invoke `bd`. Production
// code uses execBDCommand which wraps exec.CommandContext; tests inject a
// fake that returns captured JSON fixtures. The seam is intentionally
// narrower than application.CommandRunner — that port enforces an
// allowlist that would reject `bd show --json` — so the adapter owns its
// own minimal runner type.
type BDCommandRunner func(ctx context.Context, args ...string) ([]byte, error)

// BeadsGraphReader wraps `bd show --json` and `bd children --json` to
// project the dependency-graph slice the RippleHandler needs.
type BeadsGraphReader struct {
	run     BDCommandRunner
	timeout time.Duration
}

// NewBeadsGraphReader constructs a reader bound to the real `bd` binary
// with the 5-second per-call timeout convention.
func NewBeadsGraphReader() *BeadsGraphReader {
	return &BeadsGraphReader{
		run:     execBDCommand,
		timeout: BeadsGraphReaderTimeout,
	}
}

// NewBeadsGraphReaderWithRunner constructs a reader with an injected
// runner for testing.
func NewBeadsGraphReaderWithRunner(run BDCommandRunner) *BeadsGraphReader {
	return &BeadsGraphReader{
		run:     run,
		timeout: BeadsGraphReaderTimeout,
	}
}

// Timeout returns the configured per-invocation timeout (for diagnostic
// access in tests).
func (r *BeadsGraphReader) Timeout() time.Duration { return r.timeout }

// bdDependencyEntry mirrors the JSON shape emitted by `bd show --json`
// inside the `dependencies` and `dependents` arrays. Beads embeds the
// linked ticket's full record plus a `dependency_type` discriminator.
type bdDependencyEntry struct {
	ID             string `json:"id"`
	Status         string `json:"status"`
	DependencyType string `json:"dependency_type"`
}

// bdShowRecord projects the fields of `bd show --json` that the adapter
// needs. The full record is much wider; missing fields decode to zero.
type bdShowRecord struct {
	ID           string              `json:"id"`
	Status       string              `json:"status"`
	CloseReason  string              `json:"close_reason"`
	Dependencies []bdDependencyEntry `json:"dependencies"`
	Dependents   []bdDependencyEntry `json:"dependents"`
}

// bdChildRecord projects fields of `bd children --json` entries.
type bdChildRecord struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

// ReadParent extracts the parent-child dependency from `bd show --json`.
// Returns "" when no parent edge exists.
func (r *BeadsGraphReader) ReadParent(ctx context.Context, ticketID string) (string, error) {
	rec, err := r.showRecord(ctx, ticketID)
	if err != nil {
		return "", err
	}
	for _, dep := range rec.Dependencies {
		if dep.DependencyType == "parent-child" {
			return dep.ID, nil
		}
	}
	return "", nil
}

// ReadSiblings wraps `bd children <parentID> --json` and excludes
// selfID from the result. Returns the raw list (open + closed); the
// handler filters.
func (r *BeadsGraphReader) ReadSiblings(ctx context.Context, parentID, selfID string) ([]application.BeadsGraphTicket, error) {
	if parentID == "" {
		return nil, nil
	}
	timed, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	out, err := r.run(timed, "children", parentID, "--json")
	if err != nil {
		// Beads emits "Issue 'X' has no children" or "not found" on stderr
		// with a non-zero exit — both are valid "no siblings" outcomes for
		// the handler. Distinguishing transport failure from "no children"
		// is intentionally lenient: empty slice + nil error.
		stderr := commandStderr(err)
		if strings.Contains(stderr, "has no children") || strings.Contains(stderr, "not found") {
			return nil, nil
		}
		return nil, fmt.Errorf("running bd children %s: %w", parentID, wrapBDError(err))
	}

	var children []bdChildRecord
	if jerr := json.Unmarshal(out, &children); jerr != nil {
		return nil, fmt.Errorf("parsing bd children json for %s: %w", parentID, jerr)
	}
	tickets := make([]application.BeadsGraphTicket, 0, len(children))
	for _, c := range children {
		if c.ID == selfID {
			continue
		}
		tickets = append(tickets, application.BeadsGraphTicket{ID: c.ID, Status: c.Status})
	}
	return tickets, nil
}

// ReadDependents returns tickets in `dependents[]` with
// dependency_type=="blocks". Closed ones are included; handler filters.
func (r *BeadsGraphReader) ReadDependents(ctx context.Context, ticketID string) ([]application.BeadsGraphTicket, error) {
	rec, err := r.showRecord(ctx, ticketID)
	if err != nil {
		return nil, err
	}
	out := make([]application.BeadsGraphTicket, 0, len(rec.Dependents))
	for _, d := range rec.Dependents {
		if d.DependencyType != "blocks" {
			continue
		}
		out = append(out, application.BeadsGraphTicket{ID: d.ID, Status: d.Status})
	}
	return out, nil
}

// ReadRelated returns tickets joined by bidirectional `related` edges
// (either `dependencies[]` or `dependents[]`). Closed ones are
// included; handler filters and dedupes.
func (r *BeadsGraphReader) ReadRelated(ctx context.Context, ticketID string) ([]application.BeadsGraphTicket, error) {
	rec, err := r.showRecord(ctx, ticketID)
	if err != nil {
		return nil, err
	}
	var out []application.BeadsGraphTicket
	for _, d := range rec.Dependencies {
		if d.DependencyType == "related" {
			out = append(out, application.BeadsGraphTicket{ID: d.ID, Status: d.Status})
		}
	}
	for _, d := range rec.Dependents {
		if d.DependencyType == "related" {
			out = append(out, application.BeadsGraphTicket{ID: d.ID, Status: d.Status})
		}
	}
	return out, nil
}

// ReadCloseContext returns the close_reason of the ticket, or "" when
// the field is absent or empty.
func (r *BeadsGraphReader) ReadCloseContext(ctx context.Context, ticketID string) (string, error) {
	rec, err := r.showRecord(ctx, ticketID)
	if err != nil {
		return "", err
	}
	return rec.CloseReason, nil
}

// showRecord runs `bd show <id> --json` with the configured timeout and
// parses the first record from the returned array.
func (r *BeadsGraphReader) showRecord(ctx context.Context, ticketID string) (bdShowRecord, error) {
	if ticketID == "" {
		return bdShowRecord{}, fmt.Errorf("ticket ID cannot be empty")
	}
	timed, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	out, err := r.run(timed, "show", ticketID, "--json")
	if err != nil {
		return bdShowRecord{}, fmt.Errorf("running bd show %s: %w", ticketID, wrapBDError(err))
	}
	var records []bdShowRecord
	if jerr := json.Unmarshal(out, &records); jerr != nil {
		return bdShowRecord{}, fmt.Errorf("parsing bd show json for %s: %w", ticketID, jerr)
	}
	if len(records) == 0 {
		return bdShowRecord{}, fmt.Errorf("bd show %s returned no records", ticketID)
	}
	return records[0], nil
}

// execBDCommand is the default runner: invokes `bd` via exec.CommandContext.
func execBDCommand(ctx context.Context, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "bd", args...)
	out, err := cmd.Output()
	if err != nil {
		return nil, err //nolint:wrapcheck // wrapped at the caller boundary via wrapBDError
	}
	return out, nil
}

// wrapBDError surfaces the timeout vs transport distinction in the
// adapter's wrapping layer so downstream callers can pattern-match.
func wrapBDError(err error) error {
	if ctxErr := contextErrFromExec(err); ctxErr != nil {
		return ctxErr
	}
	return err
}

// contextErrFromExec teases the deadline-exceeded sentinel out of an
// exec.ExitError when applicable.
func contextErrFromExec(err error) error {
	if err == nil {
		return nil
	}
	if strings.Contains(err.Error(), context.DeadlineExceeded.Error()) {
		return context.DeadlineExceeded
	}
	return nil
}

// commandStderr extracts captured stderr from an *exec.ExitError so the
// adapter can pattern-match beads's "has no children" and "not found"
// soft failures without exposing exec internals to callers.
func commandStderr(err error) string {
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return string(exitErr.Stderr)
	}
	// Test-injected runners may emit a plain error.
	return err.Error()
}
