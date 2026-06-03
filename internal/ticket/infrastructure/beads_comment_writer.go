package infrastructure

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"time"
)

// CommentWriterTimeout is the per-invocation deadline for `bd comment add`.
const CommentWriterTimeout = 5 * time.Second

// BDCommentRunner is the runner seam for `bd comment add`. Production
// code uses execBDCommentCommand which forwards the body on stdin;
// tests inject a fake that captures the body. The signature differs
// from BDCommandRunner because comment bodies are streamed via stdin to
// avoid shell-escaping arbitrary markdown.
type BDCommentRunner func(ctx context.Context, ticketID, body string) error

// BeadsCommentWriter posts comments to beads tickets via `bd comment add`.
type BeadsCommentWriter struct {
	run     BDCommentRunner
	timeout time.Duration
}

// NewBeadsCommentWriter constructs a writer wired to the real `bd` binary.
func NewBeadsCommentWriter() *BeadsCommentWriter {
	return &BeadsCommentWriter{
		run:     execBDCommentCommand,
		timeout: CommentWriterTimeout,
	}
}

// NewBeadsCommentWriterWithRunner constructs a writer with an injected
// runner for testing.
func NewBeadsCommentWriterWithRunner(run BDCommentRunner) *BeadsCommentWriter {
	return &BeadsCommentWriter{
		run:     run,
		timeout: CommentWriterTimeout,
	}
}

// Timeout returns the per-invocation timeout (diagnostic accessor).
func (w *BeadsCommentWriter) Timeout() time.Duration { return w.timeout }

// AddComment posts comment as a new comment on ticketID.
func (w *BeadsCommentWriter) AddComment(ctx context.Context, ticketID, comment string) error {
	if ticketID == "" {
		return fmt.Errorf("ticket ID cannot be empty")
	}
	if comment == "" {
		return fmt.Errorf("comment cannot be empty")
	}
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context error: %w", err)
	}

	timed, cancel := context.WithTimeout(ctx, w.timeout)
	defer cancel()

	if err := w.run(timed, ticketID, comment); err != nil {
		return fmt.Errorf("running bd comment add %s: %w", ticketID, err)
	}
	return nil
}

// execBDCommentCommand streams the comment body to `bd comment add <id>`
// via stdin so arbitrary markdown (including backticks and quotes) cannot
// be reinterpreted by the shell.
func execBDCommentCommand(ctx context.Context, ticketID, body string) error {
	cmd := exec.CommandContext(ctx, "bd", "comment", "add", ticketID)
	cmd.Stdin = bytes.NewBufferString(body)
	if err := cmd.Run(); err != nil {
		return err //nolint:wrapcheck // wrapped by caller AddComment
	}
	return nil
}
