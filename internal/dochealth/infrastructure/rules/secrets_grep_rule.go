package rules

import (
	"fmt"
	"regexp"

	dochealthdomain "github.com/alto-cli/alto/internal/dochealth/domain"
)

// defaultSecretPatterns is the binding-floor regex set per the spike
// (L824). When SecretsGrepRule is constructed with an empty/nil pattern
// slice, these defaults apply.
//
// RE2 — no backtracking. Case-insensitive keyword set wrapped in `(?i)`.
var defaultSecretPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)\b(password|secret|api[_-]?key|token|bearer|private_key|client_secret|jwt|credentials|aws_access_key)\b`),
	regexp.MustCompile(`AKIA[0-9A-Z]{16}`),
	regexp.MustCompile(`gh[pousr]_[A-Za-z0-9]{36,}`),
}

// SecretsGrepRule warns when a scaffold body matches any configured
// secret-detection regex. Scans BODY ONLY — frontmatter is excluded to
// avoid false positives on descriptive fields (e.g.
// `description: "Manages API keys"`).
//
// Operator override: when constructed with a non-empty []SecretPattern,
// the supplied patterns REPLACE the defaults (override, not append) —
// per Contract 7 of the LOCKED Phase 1 broadcast.
type SecretsGrepRule struct {
	patterns []namedPattern
}

type namedPattern struct {
	name    string
	pattern *regexp.Regexp
}

// NewSecretsGrepRule constructs the rule. Nil/empty patterns slice means
// "apply binding-floor defaults".
func NewSecretsGrepRule(patterns []dochealthdomain.SecretPattern) *SecretsGrepRule {
	if len(patterns) == 0 {
		out := make([]namedPattern, 0, len(defaultSecretPatterns))
		names := []string{"keyword", "aws_access_key", "github_pat"}
		for i, p := range defaultSecretPatterns {
			out = append(out, namedPattern{name: names[i], pattern: p})
		}
		return &SecretsGrepRule{patterns: out}
	}
	out := make([]namedPattern, 0, len(patterns))
	for _, p := range patterns {
		out = append(out, namedPattern{name: p.Name(), pattern: p.Pattern()})
	}
	return &SecretsGrepRule{patterns: out}
}

// Name returns the stable rule identifier.
func (r *SecretsGrepRule) Name() string { return "secrets_grep" }

// Check emits one WARNING per matching pattern (deduped per pattern, not
// per match — keeps the report readable when an asset is shot through
// with the same kind of secret).
func (r *SecretsGrepRule) Check(asset dochealthdomain.ScaffoldAsset, _ []dochealthdomain.ScaffoldAsset) []dochealthdomain.ScaffoldViolation {
	body := asset.Body()
	if body == "" {
		return nil
	}
	var out []dochealthdomain.ScaffoldViolation
	for _, np := range r.patterns {
		if np.pattern == nil {
			continue
		}
		if match := np.pattern.FindString(body); match != "" {
			out = append(out, makeWarning(
				pathDisplay(asset.Path()),
				r.Name(),
				fmt.Sprintf("possible secret matched %s pattern: %q", np.name, match),
			))
		}
	}
	return out
}
