// Package application defines ports for the Discovery bounded context.
package application

import (
	"context"

	discoverydomain "github.com/alty-cli/alty/internal/discovery/domain"
	"github.com/alty-cli/alty/internal/shared/domain/ddd"
)

// --- Discovery Port ---

// Discovery manages the conversational flow of the 10-question DDD framework
// with persona detection, register selection, and playback confirmation loops.
// Methods omit context.Context because this runs synchronously in a CLI process.
type Discovery interface {
	// StartSession starts a new guided discovery session from README content.
	StartSession(readmeContent string) (*discoverydomain.DiscoverySession, error)

	// DetectPersona detects the user persona based on their self-identification choice.
	DetectPersona(sessionID string, choice string) (*discoverydomain.DiscoverySession, error)

	// AnswerQuestion submits an answer to a discovery question.
	AnswerQuestion(sessionID string, questionID string, answer string) (*discoverydomain.DiscoverySession, error)

	// SkipQuestion skips a question with an explicit reason.
	SkipQuestion(sessionID string, questionID string, reason string) (*discoverydomain.DiscoverySession, error)

	// ConfirmPlayback confirms or rejects the playback summary.
	ConfirmPlayback(sessionID string, confirmed bool) (*discoverydomain.DiscoverySession, error)

	// Complete completes the discovery session and produces domain artifacts.
	Complete(sessionID string) (*discoverydomain.DiscoverySession, error)
}

// Compile-time interface compliance check.
var _ Discovery = (*DiscoveryHandler)(nil)

// --- Artifact Renderer Port ---

// ArtifactRenderer renders a DomainModel into markdown documents (PRD, DDD.md, ARCHITECTURE.md).
type ArtifactRenderer interface {
	// RenderPRD renders the PRD markdown from a domain model.
	RenderPRD(ctx context.Context, model *ddd.DomainModel) (string, error)

	// RenderDDD renders the DDD.md markdown from a domain model.
	RenderDDD(ctx context.Context, model *ddd.DomainModel) (string, error)

	// RenderArchitecture renders the ARCHITECTURE.md markdown from a domain model.
	RenderArchitecture(ctx context.Context, model *ddd.DomainModel) (string, error)
}

// --- Tool Detection Port ---

// ToolDetection detects installed AI coding tools and scans for configuration conflicts.
type ToolDetection interface {
	// Detect detects installed AI coding tools in the project directory.
	Detect(ctx context.Context, projectDir string) ([]string, error)

	// ScanConflicts scans for global settings conflicts between detected tools.
	ScanConflicts(ctx context.Context, projectDir string) ([]discoverydomain.SettingsConflict, error)
}
