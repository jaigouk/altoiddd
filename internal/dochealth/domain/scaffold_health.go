// Package domain — scaffold health value objects.
//
// This file adds the alto-scaffold/ scaffold validation domain model alongside the
// existing docs/ DocHealth model. Both live in the DocHealth bounded context
// but address distinct concerns: docs/ tracks freshness + broken links;
// scaffold tracks frontmatter schema + leak rules + overlay pairing.
//
// Ubiquitous-language source: docs/DDD.md (ScaffoldAsset, ValidationRule,
// ScaffoldViolation, ViolationSeverity, ScaffoldHealthReport, GenericAsset,
// OverlayAsset).
package domain

import (
	"fmt"
	"maps"
	"regexp"
	"strings"
	"time"

	domainerrors "github.com/alto-cli/alto/internal/shared/domain/errors"
)

// ViolationSeverity classifies a ScaffoldViolation. ERROR severities block CI;
// WARNING severities are reported but non-blocking (for future extension rules).
type ViolationSeverity string

// Severity constants.
const (
	// SeverityError blocks: handler returns a non-nil error when any ERROR fires.
	SeverityError ViolationSeverity = "ERROR"

	// SeverityWarning is reported but does not block. Reserved for the
	// fast-follow ticket's WARNING rules (BodySize, UnknownTools, etc.).
	SeverityWarning ViolationSeverity = "WARNING"
)

// IsValid reports whether s is a recognised severity constant.
func (s ViolationSeverity) IsValid() bool {
	switch s {
	case SeverityError, SeverityWarning:
		return true
	default:
		return false
	}
}

// ScaffoldViolation is an immutable record of one rule failure on one asset.
type ScaffoldViolation struct {
	file     string
	rule     string
	message  string
	severity ViolationSeverity
	line     int
}

// NewScaffoldViolation constructs a validated ScaffoldViolation.
// file, rule and message must be non-empty; severity must be a known constant;
// line may be 0 to indicate "no specific line".
func NewScaffoldViolation(file, rule, message string, severity ViolationSeverity, line int) (ScaffoldViolation, error) {
	if strings.TrimSpace(file) == "" {
		return ScaffoldViolation{}, fmt.Errorf("file required: %w", domainerrors.ErrInvariantViolation)
	}
	if strings.TrimSpace(rule) == "" {
		return ScaffoldViolation{}, fmt.Errorf("rule required: %w", domainerrors.ErrInvariantViolation)
	}
	if strings.TrimSpace(message) == "" {
		return ScaffoldViolation{}, fmt.Errorf("message required: %w", domainerrors.ErrInvariantViolation)
	}
	if !severity.IsValid() {
		return ScaffoldViolation{}, fmt.Errorf("invalid severity %q: %w", severity, domainerrors.ErrInvariantViolation)
	}
	if line < 0 {
		return ScaffoldViolation{}, fmt.Errorf("line must be >= 0, got %d: %w", line, domainerrors.ErrInvariantViolation)
	}
	return ScaffoldViolation{
		file:     file,
		rule:     rule,
		message:  message,
		severity: severity,
		line:     line,
	}, nil
}

// File returns the asset path that produced the violation.
func (v ScaffoldViolation) File() string { return v.file }

// Rule returns the name of the ValidationRule that produced the violation.
func (v ScaffoldViolation) Rule() string { return v.rule }

// Message returns the human-readable description.
func (v ScaffoldViolation) Message() string { return v.message }

// Severity returns ERROR or WARNING.
func (v ScaffoldViolation) Severity() ViolationSeverity { return v.severity }

// Line returns the line number (1-based) or 0 if not applicable.
func (v ScaffoldViolation) Line() int { return v.line }

// ScaffoldAsset is a single alto-scaffold/**/*.md file already parsed into
// frontmatter + body. Walker constructs these; rules read them.
type ScaffoldAsset struct {
	path          string
	frontmatter   map[string]any
	body          string
	bodyLineCount int
	isOverlay     bool
	modTime       time.Time
}

// NewScaffoldAsset constructs a ScaffoldAsset. path must be non-empty; the
// frontmatter map is defensively copied (shallow). modTime defaults to the
// zero value (rules that consume mtime — e.g. LifecycleStalenessRule —
// interpret zero as "skip"). Use NewScaffoldAssetWithModTime when the
// caller has a real timestamp.
func NewScaffoldAsset(path string, frontmatter map[string]any, body string, bodyLineCount int, isOverlay bool) (ScaffoldAsset, error) {
	return NewScaffoldAssetWithModTime(path, frontmatter, body, bodyLineCount, isOverlay, time.Time{})
}

// NewScaffoldAssetWithModTime constructs a ScaffoldAsset with an explicit
// modification time captured by the walker. Used by LifecycleStalenessRule.
func NewScaffoldAssetWithModTime(path string, frontmatter map[string]any, body string, bodyLineCount int, isOverlay bool, modTime time.Time) (ScaffoldAsset, error) {
	if strings.TrimSpace(path) == "" {
		return ScaffoldAsset{}, fmt.Errorf("path required: %w", domainerrors.ErrInvariantViolation)
	}
	if bodyLineCount < 0 {
		return ScaffoldAsset{}, fmt.Errorf("bodyLineCount must be >= 0, got %d: %w", bodyLineCount, domainerrors.ErrInvariantViolation)
	}
	fm := make(map[string]any, len(frontmatter))
	maps.Copy(fm, frontmatter)
	return ScaffoldAsset{
		path:          path,
		frontmatter:   fm,
		body:          body,
		bodyLineCount: bodyLineCount,
		isOverlay:     isOverlay,
		modTime:       modTime,
	}, nil
}

// Path returns the asset path (filepath.ToSlash-normalised, repo-relative).
func (a ScaffoldAsset) Path() string { return a.path }

// Frontmatter returns a defensive copy of the parsed frontmatter map.
func (a ScaffoldAsset) Frontmatter() map[string]any {
	out := make(map[string]any, len(a.frontmatter))
	maps.Copy(out, a.frontmatter)
	return out
}

// FrontmatterValue returns the raw frontmatter value for key (or nil + false).
func (a ScaffoldAsset) FrontmatterValue(key string) (any, bool) {
	v, ok := a.frontmatter[key]
	return v, ok
}

// Body returns the markdown body (post-frontmatter).
func (a ScaffoldAsset) Body() string { return a.body }

// BodyLineCount returns the number of lines in Body (for future BodySizeRule).
func (a ScaffoldAsset) BodyLineCount() int { return a.bodyLineCount }

// IsOverlay reports whether this asset is a `.project.md` overlay.
// Overlays carry alto-internal content by design and are exempt from
// NoInternalLeaksRule (and FrontmatterSchemaRule — overlays have no
// frontmatter; the parent GENERIC sibling carries the schema).
func (a ScaffoldAsset) IsOverlay() bool { return a.isOverlay }

// ModTime returns the file modification time captured by the walker. Zero
// value means "not captured" — staleness rules treat zero as skip.
func (a ScaffoldAsset) ModTime() time.Time { return a.modTime }

// ScaffoldHealthReport aggregates ScaffoldViolations produced by all rules.
type ScaffoldHealthReport struct {
	violations []ScaffoldViolation
}

// NewScaffoldHealthReport constructs a report from a slice of violations.
// The slice is defensively copied.
func NewScaffoldHealthReport(violations []ScaffoldViolation) ScaffoldHealthReport {
	v := make([]ScaffoldViolation, len(violations))
	copy(v, violations)
	return ScaffoldHealthReport{violations: v}
}

// Violations returns a defensive copy of all violations.
func (r ScaffoldHealthReport) Violations() []ScaffoldViolation {
	out := make([]ScaffoldViolation, len(r.violations))
	copy(out, r.violations)
	return out
}

// ErrorCount returns the number of ERROR-severity violations.
func (r ScaffoldHealthReport) ErrorCount() int {
	n := 0
	for _, v := range r.violations {
		if v.severity == SeverityError {
			n++
		}
	}
	return n
}

// WarningCount returns the number of WARNING-severity violations.
func (r ScaffoldHealthReport) WarningCount() int {
	n := 0
	for _, v := range r.violations {
		if v.severity == SeverityWarning {
			n++
		}
	}
	return n
}

// HasErrors reports whether the report contains any ERROR-severity violations.
// HasErrors is the CLI's exit-code signal: non-zero exit when true.
func (r ScaffoldHealthReport) HasErrors() bool { return r.ErrorCount() > 0 }

// TotalCount returns the total number of violations.
func (r ScaffoldHealthReport) TotalCount() int { return len(r.violations) }

// SecretPattern wraps a compiled regexp + display name. Used by
// SecretsGrepRule to detect credential leaks in scaffold bodies. The name
// surfaces in violation messages so operators can identify which pattern
// fired.
type SecretPattern struct {
	name    string
	pattern *regexp.Regexp
}

// NewSecretPattern compiles pattern and constructs a validated
// SecretPattern. name must be non-empty; pattern must compile under RE2
// (regexp.Compile failure → ErrInvariantViolation).
func NewSecretPattern(name, pattern string) (SecretPattern, error) {
	if strings.TrimSpace(name) == "" {
		return SecretPattern{}, fmt.Errorf("name required: %w", domainerrors.ErrInvariantViolation)
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return SecretPattern{}, fmt.Errorf("compiling pattern %q: %w", pattern, domainerrors.ErrInvariantViolation)
	}
	return SecretPattern{name: name, pattern: re}, nil
}

// Name returns the human-readable identifier.
func (p SecretPattern) Name() string { return p.name }

// Pattern returns the compiled regexp.
func (p SecretPattern) Pattern() *regexp.Regexp { return p.pattern }

// ScaffoldParams bundles configuration passed to DefaultScaffoldRules.
// Constructed once at composition root, then handed to rule factories
// that need parameters (LifecycleStalenessRule, SecretsGrepRule).
type ScaffoldParams struct {
	defaultStalenessDays int
	secretPatterns       []SecretPattern
}

// NewScaffoldParams constructs validated ScaffoldParams.
// defaultStalenessDays must be >= 1; secretPatterns may be nil/empty
// (empty means "rule applies binding-floor defaults").
func NewScaffoldParams(defaultStalenessDays int, secretPatterns []SecretPattern) (ScaffoldParams, error) {
	if defaultStalenessDays < 1 {
		return ScaffoldParams{}, fmt.Errorf("defaultStalenessDays must be >= 1, got %d: %w", defaultStalenessDays, domainerrors.ErrInvariantViolation)
	}
	sp := make([]SecretPattern, len(secretPatterns))
	copy(sp, secretPatterns)
	return ScaffoldParams{
		defaultStalenessDays: defaultStalenessDays,
		secretPatterns:       sp,
	}, nil
}

// DefaultStalenessDays returns the configured staleness threshold (days).
func (p ScaffoldParams) DefaultStalenessDays() int { return p.defaultStalenessDays }

// SecretPatterns returns a defensive copy of the configured secret
// patterns. Empty slice signals "use binding-floor defaults".
func (p ScaffoldParams) SecretPatterns() []SecretPattern {
	out := make([]SecretPattern, len(p.secretPatterns))
	copy(out, p.secretPatterns)
	return out
}
