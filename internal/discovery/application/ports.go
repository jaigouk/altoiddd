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

// --- Session Repository Port ---

// SessionRepository persists and retrieves discovery sessions.
type SessionRepository interface {
	// Save persists a discovery session.
	Save(ctx context.Context, session *discoverydomain.DiscoverySession) error

	// Load retrieves a discovery session by ID.
	Load(ctx context.Context, sessionID string) (*discoverydomain.DiscoverySession, error)

	// Exists checks whether a persisted session exists.
	Exists(ctx context.Context, sessionID string) (bool, error)
}

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

// --- Prompter Port ---

// Prompter handles interactive CLI prompts for discovery flow.
type Prompter interface {
	// SelectPersona displays persona choices and returns the selected choice ("1"-"4").
	SelectPersona(ctx context.Context) (string, error)

	// AskQuestion displays a question and returns the user's answer.
	// Returns empty string if the user wants to skip.
	AskQuestion(ctx context.Context, question string) (string, error)

	// AskSkipReason prompts for a reason when skipping a question.
	AskSkipReason(ctx context.Context) (string, error)

	// ConfirmPlayback displays a summary and asks for confirmation.
	// Returns true if confirmed, false if user wants to review/edit.
	ConfirmPlayback(ctx context.Context, summary string) (bool, error)
}

// --- Doc Reader Port ---

// DocReader reads documentation files from a directory.
type DocReader interface {
	ReadDocs(ctx context.Context, docsDir string) (map[string]string, error)
}

// --- LLM Doc Reader Port ---

// LLMDocReader infers a domain model from document contents using an LLM.
type LLMDocReader interface {
	// InferModel takes document contents (filename->content map) and returns
	// an InferenceResult with a structured DomainModel.
	InferModel(ctx context.Context, docs map[string]string) (*discoverydomain.InferenceResult, error)
}

// --- Regex Importer Port ---

// RegexImporter imports a domain model from a docs directory using regex parsing.
// Used as fallback when LLM is unavailable.
type RegexImporter interface {
	Import(ctx context.Context, docDir string) (*ddd.DomainModel, error)
}

// --- Tool Detection Port ---

// ToolDetection detects installed AI coding tools and scans for configuration conflicts.
type ToolDetection interface {
	// Detect detects installed AI coding tools in the project directory.
	Detect(ctx context.Context, projectDir string) ([]string, error)

	// ScanConflicts scans for global settings conflicts between detected tools.
	ScanConflicts(ctx context.Context, projectDir string) ([]discoverydomain.SettingsConflict, error)
}
