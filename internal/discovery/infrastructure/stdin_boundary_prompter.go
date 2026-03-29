package infrastructure

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/alto-cli/alto/internal/discovery/application"
	discoverydomain "github.com/alto-cli/alto/internal/discovery/domain"
)

// Compile-time interface satisfaction check.
var _ application.BoundaryPrompter = (*StdinBoundaryPrompter)(nil)

// StdinBoundaryPrompter implements BoundaryPrompter using plain stdin/stdout.
type StdinBoundaryPrompter struct {
	scanner *bufio.Scanner
	writer  io.Writer
}

// NewStdinBoundaryPrompter creates a new StdinBoundaryPrompter.
func NewStdinBoundaryPrompter(r io.Reader, w io.Writer) *StdinBoundaryPrompter {
	return &StdinBoundaryPrompter{
		scanner: bufio.NewScanner(r),
		writer:  w,
	}
}

// scanOrCancel reads the next line, returning context.Canceled on EOF.
func (p *StdinBoundaryPrompter) scanOrCancel() (string, error) {
	if !p.scanner.Scan() {
		if err := p.scanner.Err(); err != nil {
			return "", fmt.Errorf("reading input: %w", err)
		}
		return "", context.Canceled
	}
	return strings.TrimSpace(p.scanner.Text()), nil
}

// DisplayBoundaryProposals presents sketches and reads acceptance.
func (p *StdinBoundaryPrompter) DisplayBoundaryProposals(
	_ context.Context,
	proposals []discoverydomain.BoundedContextSketch,
) ([]string, error) {
	_, _ = fmt.Fprintln(p.writer, "Boundary Detection Results:")
	for _, sketch := range proposals {
		_, _ = fmt.Fprintf(p.writer, "  - %s (%s, confidence: %.2f)\n",
			sketch.Name(), sketch.Classification(), sketch.Confidence())
	}
	_, _ = fmt.Fprint(p.writer, "Accept proposals (space-separated names, 'all', or 'none'): ")

	line, err := p.scanOrCancel()
	if err != nil {
		return nil, err
	}

	switch line {
	case "all":
		names := make([]string, 0, len(proposals))
		for _, s := range proposals {
			names = append(names, s.Name())
		}
		return names, nil
	case "none", "":
		return []string{}, nil
	default:
		requested := strings.Fields(line)
		valid := make(map[string]bool, len(proposals))
		for _, s := range proposals {
			valid[s.Name()] = true
		}
		accepted := make([]string, 0, len(requested))
		for _, name := range requested {
			if valid[name] {
				accepted = append(accepted, name)
			}
		}
		return accepted, nil
	}
}

// AskMissingContext asks if the user sees a missing bounded context.
func (p *StdinBoundaryPrompter) AskMissingContext(_ context.Context) (string, error) {
	_, _ = fmt.Fprint(p.writer, "Do you see a missing bounded context? (enter name, or press Enter to skip): ")

	line, err := p.scanOrCancel()
	if err != nil {
		return "", err
	}

	return line, nil
}
