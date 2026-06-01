// Package domain — Tool Translation sentinel errors for workflow asset
// translation (`alto-scaffold/commands/*.md` → tool-native command files).
//
// These sentinels live in the domain layer with stdlib `errors` as their
// only dependency, per the bounded-context layer rules in arch-go.yml.
package domain

import "errors"

// ErrInvocationProtectionNotSupported is returned by an adapter when a source
// asset declares `disable_model_invocation: true` but the target tool has no
// equivalent gating field. The adapter SKIPS the asset and returns this
// sentinel aggregated alongside any others — the handler treats it as a
// known-protected-skip signal, NOT a fatal error.
var ErrInvocationProtectionNotSupported = errors.New("invocation protection not supported by target tool")

// ErrMissingTemplate is returned when a source body references a template via
// `${CLAUDE_SKILL_DIR}/../templates/<file>.md` but the file is absent at
// translation time. The adapter emits a placeholder HTML comment in the
// output and aggregates this error.
var ErrMissingTemplate = errors.New("missing template reference")

// ErrInvalidAssetName is returned when the `name` frontmatter field fails the
// canonical `^[a-z][a-z0-9-]*$` validation. The adapter never sanitises — it
// rejects with this sentinel so the author can fix the source.
var ErrInvalidAssetName = errors.New("invalid asset name")

// ErrPathTraversal is returned when a computed output path escapes the
// declared output root (post `filepath.Clean` + prefix check). Defends
// against `name: ../evil` and similar inputs.
var ErrPathTraversal = errors.New("path traversal detected")

// ErrInvalidFrontmatter is returned when the source `alto-scaffold/commands/<name>.md`
// frontmatter is malformed YAML or is missing a required field
// (per the 8-field canonical schema).
var ErrInvalidFrontmatter = errors.New("invalid frontmatter")

// ErrOrphanOverlay is returned when a `<name>.project.md` overlay exists in
// the source directory but its primary `<name>.md` does not. Matches the
// fitness grep introduced by alty-cli-766.2.
var ErrOrphanOverlay = errors.New("orphan overlay without primary asset")
