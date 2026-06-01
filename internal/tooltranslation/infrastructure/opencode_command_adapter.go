// Package infrastructure — OpenCode command adapter.
//
// Renders `alto-scaffold/commands/*.md` workflow assets into `.opencode/commands/*.md`
// per the binding spike L307-319. Implements the
// application.WorkflowAssetGeneration port; the sibling
// application.ConfigGeneration port (DomainModel-based) is untouched.
//
// Adapter responsibilities:
//   - Enumerate primary assets under sourceDir (overlay siblings consumed in
//     merge, not emitted standalone).
//   - Skip + aggregate ErrInvocationProtectionNotSupported when a source
//     declares `disable_model_invocation: true` (binding AC: non-fatal).
//   - Strip Claude-Code bash substitution blocks, inline template refs.
//   - Emit per-command files atomically with `filepath.Clean` + prefix
//     containment defending the output path.
package infrastructure

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	ttapp "github.com/alto-cli/alto/internal/tooltranslation/application"
	ttdomain "github.com/alto-cli/alto/internal/tooltranslation/domain"
)

// openCodeCommandsSubdir is the canonical output layout under projectRoot.
// Verified against the pinned OpenCode schema fixture
// `testdata/schemas/opencode-commands.2026-05-30.md`.
const openCodeCommandsSubdir = ".opencode/commands"

// templatesSubdir is the canonical template root, sibling to commands/.
// Derived from sourceDir at translation time (so test setups that put
// commands/ + templates/ side-by-side work without configuration).
const templatesSubdir = "templates"

// OpenCodeCommandAdapter renders workflow assets into OpenCode's
// `.opencode/commands/<name>.md` format. The adapter is stateless beyond
// its (immutable) configuration — safe to share across goroutines.
type OpenCodeCommandAdapter struct{}

// NewOpenCodeCommandAdapter constructs an adapter. No configuration is
// required today; future per-tool tweaks (e.g. tools-restriction mapping)
// land via constructor options without breaking callers.
func NewOpenCodeCommandAdapter() *OpenCodeCommandAdapter {
	return &OpenCodeCommandAdapter{}
}

// Compile-time interface check — guards against silent port drift.
var _ ttapp.WorkflowAssetGeneration = (*OpenCodeCommandAdapter)(nil)

// GenerateFromAssets renders every primary `<name>.md` under sourceDir
// into `<projectRoot>/.opencode/commands/<name>.md`. Overlay siblings
// (`<name>.project.md`) are merged into the body before emission.
//
// Non-fatal per-asset errors are aggregated and returned via
// errors.Join — the caller (handler) can choose whether to surface them as
// CI failures or warnings. Fatal errors (sourceDir unreadable, output path
// traversal) short-circuit immediately.
func (a *OpenCodeCommandAdapter) GenerateFromAssets(_ context.Context, sourceDir, projectRoot string) error {
	names, orphanErrs, err := listPrimaryAssetNames(sourceDir)
	if err != nil {
		return fmt.Errorf("listing assets in %s: %w", sourceDir, err)
	}

	outputRoot := filepath.Join(projectRoot, openCodeCommandsSubdir)
	templatesDir := filepath.Join(filepath.Dir(sourceDir), templatesSubdir)

	var aggregated []error
	aggregated = append(aggregated, orphanErrs...)

	for _, name := range names {
		if err := a.renderOne(name, sourceDir, templatesDir, outputRoot, &aggregated); err != nil {
			// Per-asset fatal errors (e.g. ErrPathTraversal) are aggregated
			// alongside non-fatal ones; the binding ACs only require fatal
			// short-circuit for sourceDir-level failures.
			aggregated = append(aggregated, fmt.Errorf("asset %s: %w", name, err))
		}
	}
	return aggregateErrors(aggregated)
}

// renderOne loads a single asset, applies all transformations, and emits
// the output file. Non-fatal errors (missing templates, invocation
// protection) are appended to *aggregated; fatal errors (traversal,
// frontmatter, write failure) are returned.
func (a *OpenCodeCommandAdapter) renderOne(
	name, sourceDir, templatesDir, outputRoot string,
	aggregated *[]error,
) error {
	src, err := loadAssetSource(sourceDir, name)
	if err != nil {
		return err
	}

	// Binding AC: skip + aggregate, never silently drop and never fatal.
	if src.frontmatter.DisableModelInvocation {
		*aggregated = append(*aggregated, fmt.Errorf("asset %s: %w", name, ttdomain.ErrInvocationProtectionNotSupported))
		return nil
	}

	body := stripBashBlocks(src.body)
	rewritten, tmplErrs := inlineTemplateRefs(body, templatesDir)
	*aggregated = append(*aggregated, tmplErrs...)

	out, err := safeJoin(outputRoot, name+".md")
	if err != nil {
		return err
	}

	rendered := renderOpenCodeOutput(src.frontmatter, rewritten)
	if err := writeAtomic(out, []byte(rendered)); err != nil {
		return fmt.Errorf("writing %s: %w", out, err)
	}
	return nil
}

// renderOpenCodeOutput formats the OpenCode-flavoured frontmatter
// (`description` + optional `agent`) followed by the transformed body.
//
// Per binding AC: `agent:` is emitted ONLY when the source declared it
// (even if empty). Absent in source → absent in output. This avoids the
// "ghost `agent: ""`" anti-pattern that downstream tools may interpret as
// an explicit assignment.
func renderOpenCodeOutput(fm workflowAssetFrontmatter, body string) string {
	var b strings.Builder
	b.WriteString("---\n")
	b.WriteString("description: ")
	b.WriteString(yamlScalar(fm.Description))
	b.WriteString("\n")
	if fm.HasAgent {
		b.WriteString("agent: ")
		b.WriteString(yamlScalar(fm.Agent))
		b.WriteString("\n")
	}
	b.WriteString("---\n")
	if !strings.HasPrefix(body, "\n") {
		b.WriteString("\n")
	}
	b.WriteString(body)
	if !strings.HasSuffix(body, "\n") {
		b.WriteString("\n")
	}
	return b.String()
}

// yamlScalar emits a single-line scalar with minimal YAML safety. Strings
// containing `:` `#` `"` or leading whitespace are double-quoted. Newlines
// are NOT supported — frontmatter values in this schema are single-line
// by design (per the 8-field canonical schema).
func yamlScalar(v string) string {
	if v == "" {
		return `""`
	}
	if strings.ContainsAny(v, ":#\"\n") || strings.HasPrefix(v, " ") {
		escaped := strings.ReplaceAll(v, `\`, `\\`)
		escaped = strings.ReplaceAll(escaped, `"`, `\"`)
		return `"` + escaped + `"`
	}
	return v
}
