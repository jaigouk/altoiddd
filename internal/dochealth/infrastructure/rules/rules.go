// Package rules — ValidationRule implementations for ScaffoldHealthHandler.
//
// Each rule lives in its own file (ISP — testable + suppressable
// individually). Rules are stateless and deterministic; the handler
// composes them via DefaultScaffoldRules in
// internal/dochealth/infrastructure/default_scaffold_rules.go.
//
// Ubiquitous-language: every rule's Name() returns a stable string used
// in violation reports and (future) per-rule suppression configuration.
package rules

import (
	"path/filepath"
	"strings"

	dochealthapp "github.com/alto-cli/alto/internal/dochealth/application"
	dochealthdomain "github.com/alto-cli/alto/internal/dochealth/domain"
)

// makeViolation panics if construction fails — every input here is
// controlled by the rule itself, never by attacker input, so a panic
// signals a programmer error (the rule built an invalid violation).
func makeViolation(file, ruleName, message string) dochealthdomain.ScaffoldViolation {
	v, err := dochealthdomain.NewScaffoldViolation(file, ruleName, message, dochealthdomain.SeverityError, 0)
	if err != nil {
		panic("rule built invalid violation: " + err.Error())
	}
	return v
}

// pathDisplay returns a repo-relative slash-normalised path for display.
func pathDisplay(p string) string { return filepath.ToSlash(p) }

// Compile-time interface checks for every concrete rule.
var (
	_ dochealthapp.ValidationRule = (*FrontmatterSchemaRule)(nil)
	_ dochealthapp.ValidationRule = (*PhaseEnumRule)(nil)
	_ dochealthapp.ValidationRule = (*NoInternalLeaksRule)(nil)
	_ dochealthapp.ValidationRule = (*OrphanOverlayRule)(nil)
)

// stringValue extracts a non-empty string from a frontmatter map. Returns
// (value, true) when key exists, value is a string AND non-empty. Empty
// strings are treated as missing — frontmatter rules want "present and
// non-empty" semantics across the board.
func stringValue(fm map[string]any, key string) (string, bool) {
	raw, ok := fm[key]
	if !ok {
		return "", false
	}
	s, ok := raw.(string)
	if !ok {
		return "", false
	}
	if strings.TrimSpace(s) == "" {
		return "", false
	}
	return s, true
}
