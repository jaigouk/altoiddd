package rules

import (
	"bytes"
	"fmt"
	"text/template"

	dochealthdomain "github.com/alto-cli/alto/internal/dochealth/domain"
)

// sentinelTemplateParams is the deterministic, in-memory ScaffoldParams shape
// used by TemplateRenderabilityRule to evaluate every embedded template at CI
// time. Field set MIRRORS the bootstrap domain ScaffoldParams (ProjectName,
// TicketPrefix, IssueTracker, BoundedContexts, PrimaryTool, IncludeHooks) so
// that any embed referencing a real field renders cleanly and any embed
// referencing an UNKNOWN field surfaces an ERROR violation here — before ship.
//
// We define a LOCAL struct rather than importing
// internal/bootstrap/domain.ScaffoldParams. Two reasons:
//
//  1. Zero coupling — the rule lives in the Doc Health bounded context and
//     should not reach into Bootstrap to read a runtime VO.
//  2. The sentinel must never contain real data; values below are synthetic
//     by construction and reviewable in one place.
//
// Privacy: every value is a safe synthetic literal. No home paths, no
// usernames, no hostnames, no emails, no secrets.
type sentinelTemplateParams struct {
	ProjectName     string
	TicketPrefix    string
	IssueTracker    string
	BoundedContexts []string
	PrimaryTool     string
	IncludeHooks    bool
}

// newSentinelTemplateParams returns a deterministic sentinel value. Constructed
// once per Check call — the struct is tiny and the rule remains pure.
func newSentinelTemplateParams() sentinelTemplateParams {
	return sentinelTemplateParams{
		ProjectName:     "demo",
		TicketPrefix:    "demo-",
		IssueTracker:    "beads",
		BoundedContexts: []string{"Core"},
		PrimaryTool:     "claude",
		IncludeHooks:    true,
	}
}

// TemplateRenderabilityRule asserts that every embedded scaffold asset parses
// as a text/template AND executes cleanly against a sentinel ScaffoldParams
// under Option("missingkey=error"). It mirrors the EmbedScaffoldWriter
// renderTemplate chain (internal/bootstrap/infrastructure/embed_scaffold_writer.go)
// so author typos like {{.UnknownField}} surface at CI time instead of
// blowing up `alto init --with-scaffold` for the operator.
//
// Severity ERROR — author-typo DoS at runtime is exactly the failure mode
// this rule prevents; downgrading to WARNING would defeat the purpose.
//
// The rule is pure: deterministic, no I/O, no clock reads.
type TemplateRenderabilityRule struct{}

// NewTemplateRenderabilityRule constructs the rule.
func NewTemplateRenderabilityRule() *TemplateRenderabilityRule {
	return &TemplateRenderabilityRule{}
}

// Name returns the stable rule identifier.
func (r *TemplateRenderabilityRule) Name() string { return "template_renderability" }

// Check parses asset.Body() as a text/template and executes it against a
// sentinel ScaffoldParams. Parse-time AND execute-time failures BOTH produce a
// single severity-ERROR violation, distinguishable by message prefix.
//
// The chain MIRRORS embed_scaffold_writer.renderTemplate:
//
//	template.New(path).Option("missingkey=error").Parse(body).Execute(buf, params)
//
// The rule reimplements the chain rather than calling renderTemplate directly
// to keep zero coupling between the Doc Health bounded context and the
// Bootstrap infrastructure layer (and because renderTemplate is intentionally
// unexported in the bootstrap package).
func (r *TemplateRenderabilityRule) Check(
	asset dochealthdomain.ScaffoldAsset,
	_ []dochealthdomain.ScaffoldAsset,
) []dochealthdomain.ScaffoldViolation {
	body := asset.Body()
	if body == "" {
		return nil
	}
	tmpl, err := template.New(asset.Path()).Option("missingkey=error").Parse(body)
	if err != nil {
		return []dochealthdomain.ScaffoldViolation{
			makeViolation(
				pathDisplay(asset.Path()),
				r.Name(),
				fmt.Sprintf("template parse error: %s", err.Error()),
			),
		}
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, newSentinelTemplateParams()); err != nil {
		return []dochealthdomain.ScaffoldViolation{
			makeViolation(
				pathDisplay(asset.Path()),
				r.Name(),
				fmt.Sprintf("template execute error: %s", err.Error()),
			),
		}
	}
	return nil
}
