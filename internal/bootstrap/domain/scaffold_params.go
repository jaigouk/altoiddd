package domain

import (
	"fmt"
	"strings"

	domainerrors "github.com/alto-cli/alto/internal/shared/domain/errors"
)

// ScaffoldParams is the immutable value object carrying the five parameters
// substituted into the embedded alto-scaffold/ scaffold at `alto init --with-scaffold`
// time. Fields are exported PascalCase because text/template addresses them
// via {{.ProjectName}} etc.; the constructor enforces all invariants so the
// adapter never receives an unsafe value.
type ScaffoldParams struct {
	ProjectName     string
	TicketPrefix    string
	IssueTracker    string
	BoundedContexts []string
	PrimaryTool     string
	// IncludeHooks toggles the post-close beads hook scaffold. Defaults
	// to true at constructor entry so new projects get the auto-ripple
	// wiring without any extra flag. Operators opt out via --no-hooks.
	IncludeHooks bool
}

// knownIssueTrackers enumerates the supported --issue-tracker values.
var knownIssueTrackers = map[string]struct{}{
	"beads":  {},
	"github": {},
	"linear": {},
}

// knownPrimaryTools enumerates the supported --primary-tool values per the
// epic §Scope Update. Cursor and Roo Code are intentionally excluded; their
// values are rejected as unknown, not "not yet implemented".
var knownPrimaryTools = map[string]struct{}{
	"claude":   {},
	"opencode": {},
}

// NewScaffoldParams validates the five scaffold-template parameters and
// returns an immutable ScaffoldParams. All validation errors wrap
// domainerrors.ErrInvariantViolation; unsafe characters in string fields
// additionally cite ErrUnsafeTemplateParameter for diagnostic clarity.
func NewScaffoldParams(
	projectName string,
	ticketPrefix string,
	issueTracker string,
	boundedContexts []string,
	primaryTool string,
) (ScaffoldParams, error) {
	if err := validateProjectName(projectName); err != nil {
		return ScaffoldParams{}, err
	}
	if err := validateTicketPrefix(ticketPrefix); err != nil {
		return ScaffoldParams{}, err
	}
	if _, ok := knownIssueTrackers[issueTracker]; !ok {
		return ScaffoldParams{}, fmt.Errorf("unknown issue tracker %q: %w", issueTracker, domainerrors.ErrInvariantViolation)
	}
	for _, ctx := range boundedContexts {
		if err := validateBoundedContextName(ctx); err != nil {
			return ScaffoldParams{}, err
		}
	}
	if _, ok := knownPrimaryTools[primaryTool]; !ok {
		return ScaffoldParams{}, fmt.Errorf("unknown primary tool %q: %w", primaryTool, domainerrors.ErrInvariantViolation)
	}

	contextsCopy := append([]string(nil), boundedContexts...)
	return ScaffoldParams{
		ProjectName:     projectName,
		TicketPrefix:    ticketPrefix,
		IssueTracker:    issueTracker,
		BoundedContexts: contextsCopy,
		PrimaryTool:     primaryTool,
		IncludeHooks:    true,
	}, nil
}

// WithIncludeHooks returns a copy of params with the IncludeHooks toggle
// set to the supplied value. Used by the CLI to thread the --no-hooks
// flag without mutating the value object.
func (p ScaffoldParams) WithIncludeHooks(include bool) ScaffoldParams {
	cp := p
	cp.IncludeHooks = include
	return cp
}

// validateProjectName rejects empty values, path-traversal characters,
// embedded NUL, and shell metacharacters. `{` and `}` are intentionally NOT
// in the deny list — template-injection is mitigated by text/template DATA
// binding at render time, not by broader character denial.
func validateProjectName(name string) error {
	if name == "" {
		return fmt.Errorf("project name must not be empty: %w", domainerrors.ErrInvariantViolation)
	}
	if strings.ContainsAny(name, `/\`) {
		return fmt.Errorf("project name %q contains path separator: %w: %w", name, ErrUnsafeTemplateParameter, domainerrors.ErrInvariantViolation)
	}
	if strings.Contains(name, "..") {
		return fmt.Errorf("project name %q contains path traversal: %w: %w", name, ErrUnsafeTemplateParameter, domainerrors.ErrInvariantViolation)
	}
	if strings.ContainsRune(name, '\x00') {
		return fmt.Errorf("project name contains NUL: %w: %w", ErrUnsafeTemplateParameter, domainerrors.ErrInvariantViolation)
	}
	if containsShellMetachar(name) {
		return fmt.Errorf("project name %q contains shell metacharacter: %w: %w", name, ErrUnsafeTemplateParameter, domainerrors.ErrInvariantViolation)
	}
	return nil
}

// validateTicketPrefix enforces the regex ^[a-zA-Z][a-zA-Z0-9-]*-$ inline
// (no regexp dependency in the domain layer per arch-go strictness).
func validateTicketPrefix(prefix string) error {
	if prefix == "" {
		return fmt.Errorf("ticket prefix must not be empty: %w", domainerrors.ErrInvariantViolation)
	}
	if containsShellMetachar(prefix) || strings.ContainsAny(prefix, `/\`) || strings.Contains(prefix, "..") {
		return fmt.Errorf("ticket prefix %q contains unsafe character: %w: %w", prefix, ErrUnsafeTemplateParameter, domainerrors.ErrInvariantViolation)
	}
	if !strings.HasSuffix(prefix, "-") {
		return fmt.Errorf("ticket prefix %q must end with '-': %w", prefix, domainerrors.ErrInvariantViolation)
	}
	first := prefix[0]
	if !isAlpha(first) {
		return fmt.Errorf("ticket prefix %q must begin with a letter: %w", prefix, domainerrors.ErrInvariantViolation)
	}
	for i := 1; i < len(prefix); i++ {
		c := prefix[i]
		if !isAlpha(c) && !isDigit(c) && c != '-' {
			return fmt.Errorf("ticket prefix %q contains invalid character %q: %w", prefix, c, domainerrors.ErrInvariantViolation)
		}
	}
	return nil
}

// validateBoundedContextName enforces ^[A-Z][a-zA-Z0-9]*$ (PascalCase).
func validateBoundedContextName(name string) error {
	if name == "" {
		return fmt.Errorf("bounded-context name must not be empty: %w", domainerrors.ErrInvariantViolation)
	}
	if first := name[0]; first < 'A' || first > 'Z' {
		return fmt.Errorf("bounded-context name %q must begin with an uppercase letter: %w", name, domainerrors.ErrInvariantViolation)
	}
	for i := 1; i < len(name); i++ {
		c := name[i]
		if !isAlpha(c) && !isDigit(c) {
			return fmt.Errorf("bounded-context name %q contains invalid character %q: %w", name, c, domainerrors.ErrInvariantViolation)
		}
	}
	return nil
}

// containsShellMetachar reports whether s contains any character that
// requires quoting in a POSIX shell.
//
// Known limitation: this deny list is for POST-TEMPLATE-RENDER SAFETY — it
// prevents shell metachars from reaching template-binding contexts where
// they could be re-evaluated. It is NOT a shell-eval-safe sanitiser; if
// future callers interpolate ProjectName into shell commands (e.g.
// `git init "{{.ProjectName}}"`), they MUST add their own escaping. Glob
// metachars (`*`, `?`, `[`, `]`, `~`, `!`) are intentionally NOT in the
// deny list because text/template renders them as literal text. `{` /
// `}` are likewise omitted — text/template DATA binding handles
// attacker-controlled `{{.Evil}}` strings as literal output, never as
// re-evaluated template expressions.
func containsShellMetachar(s string) bool {
	const metas = "$`;|&<>()\n\r"
	return strings.ContainsAny(s, metas)
}

func isAlpha(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z')
}

func isDigit(b byte) bool {
	return b >= '0' && b <= '9'
}
