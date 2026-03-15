// Package application defines ports for the Bootstrap bounded context.
package application

import (
	"context"

	"github.com/alty-cli/alty/internal/bootstrap/domain"
)

// Bootstrap defines the interface for project bootstrap operations.
// Adapters implement this to handle the preview-confirm-execute flow
// for creating a new project seed from a README idea.
type Bootstrap interface {
	// Preview returns a human-readable preview of planned bootstrap actions.
	Preview(ctx context.Context, projectDir string) (string, error)

	// Confirm confirms a previewed bootstrap session.
	Confirm(ctx context.Context, sessionID string) (string, error)

	// Execute executes a confirmed bootstrap session.
	Execute(ctx context.Context, sessionID string) (string, error)
}

// ProjectDetector detects the state of an existing project directory.
// Used by init to auto-detect whether to run the new-project or rescue path.
type ProjectDetector interface {
	Detect(projectDir string) (domain.ProjectDetectionResult, error)
}
