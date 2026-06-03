// Package infrastructure — workflow-asset frontmatter helper used by the
// OpenCode command adapter (and any future Cursor/Roo Code adapters).
//
// Responsibilities (single-file scope by design):
//   - Parse the canonical 8-field frontmatter schema produced by
//     alty-cli-766.2 from `alto-scaffold/commands/<name>.md`.
//   - Detect overlay siblings (`<name>.project.md`) and append their body
//     (newline-separated) per the binding spike L307-319.
//   - Strip `!“cmd```` inline and ```!“ fenced bash blocks, replacing
//     each with an HTML comment that preserves the original command verbatim.
//   - Inline `${CLAUDE_SKILL_DIR}/../templates/<file>.md` references so the
//     emitted asset is self-contained.
//   - Validate `name` against `^[a-z][a-z0-9-]*$`.
//   - Defend the output path with `filepath.Clean` + prefix containment.
//
// arch-go forbids cross-context import, so the parser duplicates a small
// subset of `internal/challenge/infrastructure/yaml_frontmatter_parser.go`.
// A shared-kernel refactor stub ticket will hoist both into
// `internal/shared/infrastructure/markdown/` once 766.5 and 766.7 close.
package infrastructure

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/alto-cli/alto/internal/shared/infrastructure/markdown"
	ttdomain "github.com/alto-cli/alto/internal/tooltranslation/domain"
)

// workflowAssetFrontmatter mirrors the canonical 8-field schema from the
// alty-cli-766.2 spike (lines 224-244). `Agent` and
// `DisableModelInvocation` are optional extensions.
type workflowAssetFrontmatter struct {
	Name                   string `yaml:"name"`
	Description            string `yaml:"description"`
	Kind                   string `yaml:"kind"`
	Phase                  string `yaml:"phase"`
	WhenToUse              string `yaml:"when_to_use"`
	ToolsRequired          any    `yaml:"tools"`
	BashSubstitutionPolicy string `yaml:"bash_substitution_policy"`
	License                string `yaml:"license"`
	Agent                  string `yaml:"agent,omitempty"`
	DisableModelInvocation bool   `yaml:"disable_model_invocation,omitempty"`
	HasAgent               bool   `yaml:"-"` // tracked separately because `Agent: ""` cannot be distinguished from absent.
}

// workflowAssetSource is the merged primary + overlay view of one asset.
type workflowAssetSource struct {
	name        string
	primaryPath string
	frontmatter workflowAssetFrontmatter
	body        string
}

// nameRegex is the canonical asset-name validator. Anchored on both ends;
// linear (RE2) — no ReDoS surface.
var nameRegex = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)

// bashInlineRegex matches “ !`<cmd>` “ — Claude Code's inline bash
// substitution syntax.
var bashInlineRegex = regexp.MustCompile("!`([^`]+)`")

// bashFencedRegex matches a ```!\n…``` fenced block. `(?s)` enables
// dot-matches-newline for the body capture.
var bashFencedRegex = regexp.MustCompile("(?s)```!\\n(.*?)```")

// templateRefRegex matches `${CLAUDE_SKILL_DIR}/../templates/<file>.md`.
// The file portion is restricted to a single path segment for defence
// against `../../../../etc/passwd` style attacks.
var templateRefRegex = regexp.MustCompile(`\$\{CLAUDE_SKILL_DIR\}/\.\./templates/([a-zA-Z0-9_.-]+\.md)`)

// loadAssetSource reads `<sourceDir>/<name>.md`, parses frontmatter, applies
// the matching `<name>.project.md` overlay (if any), and returns a merged
// workflowAssetSource. Frontmatter validation is performed here so callers
// can rely on `frontmatter.Name` being non-empty + canonical.
func loadAssetSource(sourceDir, name string) (workflowAssetSource, error) {
	primaryPath := filepath.Join(sourceDir, name+".md")
	primaryBytes, err := os.ReadFile(primaryPath) //nolint:gosec // sourceDir is operator-supplied; name is regex-validated below.
	if err != nil {
		return workflowAssetSource{}, fmt.Errorf("reading primary %s: %w", primaryPath, err)
	}

	fm, body, err := parseFrontmatter(string(primaryBytes))
	if err != nil {
		return workflowAssetSource{}, fmt.Errorf("parsing %s: %w", primaryPath, err)
	}

	if !nameRegex.MatchString(fm.Name) {
		return workflowAssetSource{}, fmt.Errorf("name %q does not match %s: %w",
			fm.Name, nameRegex.String(), ttdomain.ErrInvalidAssetName)
	}

	// Overlay merge — body APPENDED to primary, newline-separated.
	overlayPath := filepath.Join(sourceDir, name+".project.md")
	if overlayBytes, oerr := os.ReadFile(overlayPath); oerr == nil { //nolint:gosec // path derived from validated name.
		_, overlayBody, perr := parseFrontmatter(string(overlayBytes))
		if perr != nil {
			return workflowAssetSource{}, fmt.Errorf("parsing overlay %s: %w", overlayPath, perr)
		}
		body = strings.TrimRight(body, "\n") + "\n\n" + strings.TrimLeft(overlayBody, "\n")
	}

	return workflowAssetSource{
		name:        fm.Name,
		primaryPath: primaryPath,
		frontmatter: fm,
		body:        body,
	}, nil
}

// parseFrontmatter splits `---\n…\n---\n<body>`. Empty/no-frontmatter input
// returns ErrInvalidFrontmatter — workflow assets always carry frontmatter.
//
// The split + generic unmarshal pass run through the shared markdown kernel
// (alty-cli-1r0). The TYPED second pass STAYS local — workflowAssetFrontmatter
// is context-specific and the HasAgent presence-detection trick (compare
// generic map keys against typed struct) is a tooltranslation invariant.
func parseFrontmatter(content string) (workflowAssetFrontmatter, string, error) {
	rawFM, body, hasFrontmatter, err := markdown.ExtractFrontmatter(content)
	if err != nil {
		return workflowAssetFrontmatter{}, "", fmt.Errorf("extracting frontmatter: %w", ttdomain.ErrInvalidFrontmatter)
	}
	if !hasFrontmatter {
		return workflowAssetFrontmatter{}, "", fmt.Errorf("missing frontmatter delimiter: %w", ttdomain.ErrInvalidFrontmatter)
	}

	// Marshal into a generic map first so HasAgent can be detected even when
	// the value is the empty string.
	generic, err := markdown.ParseGeneric(rawFM)
	if err != nil {
		return workflowAssetFrontmatter{}, "", fmt.Errorf("yaml unmarshal: %w", ttdomain.ErrInvalidFrontmatter)
	}

	var fm workflowAssetFrontmatter
	if err := yaml.Unmarshal([]byte(rawFM), &fm); err != nil {
		return workflowAssetFrontmatter{}, "", fmt.Errorf("yaml unmarshal (typed): %w", ttdomain.ErrInvalidFrontmatter)
	}
	if _, ok := generic["agent"]; ok {
		fm.HasAgent = true
	}
	return fm, body, nil
}

// stripBashBlocks replaces every inline “ !`<cmd>` “ and fenced ```! …```
// block with an HTML comment that preserves the original command verbatim.
// The replacement is deterministic — no escaping is applied to the captured
// command, since it's already inside `<!-- … -->`.
func stripBashBlocks(body string) string {
	body = bashFencedRegex.ReplaceAllStringFunc(body, func(match string) string {
		cmd := bashFencedRegex.FindStringSubmatch(match)[1]
		return fmt.Sprintf("<!-- Tool Translation stripped fenced bash block: ```!\n%s``` — port manually if running outside Claude Code -->", cmd)
	})
	body = bashInlineRegex.ReplaceAllStringFunc(body, func(match string) string {
		cmd := bashInlineRegex.FindStringSubmatch(match)[1]
		return fmt.Sprintf("<!-- Tool Translation stripped: !`%s` — port manually if running outside Claude Code -->", cmd)
	})
	return body
}

// inlineTemplateRefs replaces `${CLAUDE_SKILL_DIR}/../templates/<file>.md`
// with the literal contents of `<sourceDirRoot>/templates/<file>.md`.
//
// Per the spike L307-319 the OpenCode adapter INLINES referenced templates
// — assuming the consumer tool can resolve `@<relpath>` is unsafe because
// `alto-scaffold/templates/` may not exist in the target. Missing templates produce
// an `ErrMissingTemplate` aggregated alongside the rendered body.
func inlineTemplateRefs(body, templatesDir string) (string, []error) {
	var errs []error
	rewritten := templateRefRegex.ReplaceAllStringFunc(body, func(match string) string {
		file := templateRefRegex.FindStringSubmatch(match)[1]
		// Defence-in-depth: file segment regex already excludes `/` and `..`,
		// but Clean + prefix-check guards against future regex relaxation.
		candidate := filepath.Clean(filepath.Join(templatesDir, file))
		cleanedRoot := filepath.Clean(templatesDir)
		if !strings.HasPrefix(candidate, cleanedRoot+string(os.PathSeparator)) && candidate != cleanedRoot {
			errs = append(errs, fmt.Errorf("template %q escapes templates root: %w", file, ttdomain.ErrPathTraversal))
			return fmt.Sprintf("<!-- Tool Translation: rejected template %s (path traversal) -->", file)
		}
		contents, err := os.ReadFile(candidate) //nolint:gosec // candidate is contained by Clean + prefix check above.
		if err != nil {
			errs = append(errs, fmt.Errorf("template %s: %w", file, ttdomain.ErrMissingTemplate))
			return fmt.Sprintf("<!-- Tool Translation: missing template %s -->", file)
		}
		return string(contents)
	})
	return rewritten, errs
}

// safeJoin returns filepath.Clean(filepath.Join(root, name)) only when the
// result is contained by root. Otherwise ErrPathTraversal is returned.
func safeJoin(root, name string) (string, error) {
	cleanedRoot := filepath.Clean(root)
	candidate := filepath.Clean(filepath.Join(cleanedRoot, name))
	if candidate == cleanedRoot {
		return "", fmt.Errorf("name %q resolves to root: %w", name, ttdomain.ErrPathTraversal)
	}
	if !strings.HasPrefix(candidate, cleanedRoot+string(os.PathSeparator)) {
		return "", fmt.Errorf("name %q escapes root %s: %w", name, cleanedRoot, ttdomain.ErrPathTraversal)
	}
	return candidate, nil
}

// writeAtomic writes content to target using a sibling tempfile + os.Rename
// so partial writes are never observable. Mode 0o644 matches the existing
// FilesystemFileWriter contract used by ConfigGeneration.
func writeAtomic(target string, content []byte) error {
	dir := filepath.Dir(target)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", dir, err)
	}
	tmp, err := os.CreateTemp(dir, ".workflow-asset-*.tmp")
	if err != nil {
		return fmt.Errorf("create tempfile in %s: %w", dir, err)
	}
	tmpPath := tmp.Name()
	cleanup := func() { _ = os.Remove(tmpPath) }

	if _, err := tmp.Write(content); err != nil {
		_ = tmp.Close()
		cleanup()
		return fmt.Errorf("write %s: %w", tmpPath, err)
	}
	if err := tmp.Chmod(0o644); err != nil {
		_ = tmp.Close()
		cleanup()
		return fmt.Errorf("chmod %s: %w", tmpPath, err)
	}
	if err := tmp.Close(); err != nil {
		cleanup()
		return fmt.Errorf("close %s: %w", tmpPath, err)
	}
	if err := os.Rename(tmpPath, target); err != nil {
		cleanup()
		return fmt.Errorf("rename %s -> %s: %w", tmpPath, target, err)
	}
	return nil
}

// listPrimaryAssetNames returns every `<name>.md` under sourceDir whose
// matching `<name>.project.md` does not exist as the only entry (overlays
// are silently consumed during the merge step, not emitted standalone).
//
// Orphan overlays (`foo.project.md` with no `foo.md` sibling) are detected
// here and surfaced via ErrOrphanOverlay so the handler can aggregate.
func listPrimaryAssetNames(sourceDir string) ([]string, []error, error) {
	entries, err := os.ReadDir(sourceDir)
	if err != nil {
		return nil, nil, fmt.Errorf("read source dir %s: %w", sourceDir, err)
	}
	primaries := map[string]struct{}{}
	overlays := map[string]struct{}{}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		n := e.Name()
		switch {
		case strings.HasSuffix(n, ".project.md"):
			overlays[strings.TrimSuffix(n, ".project.md")] = struct{}{}
		case strings.HasSuffix(n, ".md"):
			primaries[strings.TrimSuffix(n, ".md")] = struct{}{}
		}
	}
	var errs []error
	for o := range overlays {
		if _, ok := primaries[o]; !ok {
			errs = append(errs, fmt.Errorf("overlay %s.project.md has no primary: %w", o, ttdomain.ErrOrphanOverlay))
		}
	}
	names := make([]string, 0, len(primaries))
	for p := range primaries {
		names = append(names, p)
	}
	return names, errs, nil
}

// aggregateErrors combines a slice of errors using errors.Join. Returns nil
// for empty input so callers can detect the "all clean" case.
func aggregateErrors(errs []error) error {
	if len(errs) == 0 {
		return nil
	}
	return errors.Join(errs...)
}
