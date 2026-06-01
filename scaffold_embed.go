// Package alto hosts the //go:embed declaration for the alto-scaffold/ scaffold
// tree. It exists because go:embed patterns are package-directory-relative
// and may not traverse upward with ".." — and alto-scaffold/ lives at the module
// root, so the embed directive MUST be in a Go file sitting next to the
// alto-scaffold/ directory. The consuming adapter is
// internal/bootstrap/infrastructure/embed_scaffold_writer.go, which imports
// the FS into a domain-typed EmbedScaffoldWriter.
//
// This package is intentionally not under internal/ so that the //go:embed
// directive can reach the alto-scaffold/ subtree. It contains NO business logic —
// only the embed.FS variable. All extraction, templating, and validation
// logic stays in internal/bootstrap/infrastructure/.
//
// NOTE on package name: the TL decision authorised a "scaffoldfs" host
// package; however, //go:embed paths are relative to the source file's
// directory, so a subdirectory (e.g. scaffoldfs/) cannot reach
// ../alto-scaffold/ either. The file therefore lives at the module root with
// package name `alto`; the exported symbol `ScaffoldFS` remains the sole
// public surface and matches the spirit of the locked contract.
package alto

import "embed"

// ScaffoldFS is the embedded alto-scaffold/ scaffold tree shipped with the alto
// binary. It is the sole public surface of this build-time resource
// package; consumers obtain a domain-typed wrapper via
// internal/bootstrap/infrastructure.NewEmbedScaffoldWriter.
//
//go:embed alto-scaffold/CONTEXT.md alto-scaffold/README.md alto-scaffold/commands alto-scaffold/agents alto-scaffold/templates alto-scaffold/scripts all:alto-scaffold/skills
var ScaffoldFS embed.FS
