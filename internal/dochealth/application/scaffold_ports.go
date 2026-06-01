// Package application — scaffold-health ports.
//
// These ports sit alongside the existing DocHealth + DocReview ports
// (ports.go) but address a distinct concern: validating `alto-scaffold/**/*.md`
// scaffold assets (frontmatter schema, leak rules, overlay pairing) as
// opposed to docs/ freshness.
//
// Single `ValidationRule` interface — no `ListValidationRule` second port
// and no runtime type assertion in the handler. Rules that don't need the
// full corpus simply ignore the second parameter (signature honesty over
// a runtime fallback no-op).
package application

import (
	"context"

	dochealthdomain "github.com/alto-cli/alto/internal/dochealth/domain"
)

// ValidationRule checks one ScaffoldAsset against the full corpus and
// returns zero or more violations.
//
// `corpus` is the complete slice of assets discovered by the walker; rules
// that don't need cross-asset context (e.g. FrontmatterSchemaRule) ignore
// it. Rules that DO need it (e.g. OrphanOverlayRule) use it to look up
// siblings.
//
// Implementations MUST be deterministic — same (asset, corpus) input
// produces the same violation slice. They MUST NOT perform I/O.
type ValidationRule interface {
	// Name returns a stable identifier — used in violation reports and
	// suppression configuration (future).
	Name() string

	// Check returns the violations produced by this rule on this asset.
	// Empty slice (or nil) means "clean".
	Check(asset dochealthdomain.ScaffoldAsset, corpus []dochealthdomain.ScaffoldAsset) []dochealthdomain.ScaffoldViolation
}

// ScaffoldWalker enumerates `alto-scaffold/**/*.md` files under altoDir, parses
// each into a ScaffoldAsset, and returns the slice. The walker is the
// only piece that performs filesystem I/O for the scaffold-health flow.
type ScaffoldWalker interface {
	Walk(ctx context.Context, altoDir string) ([]dochealthdomain.ScaffoldAsset, error)
}
