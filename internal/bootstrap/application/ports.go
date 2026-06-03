// Package application defines ports for the Bootstrap bounded context.
package application

import (
	"context"

	"github.com/alto-cli/alto/internal/bootstrap/domain"
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

// GitCommitter stages and commits generated files after bootstrap execution.
// Defined as a narrow ISP interface — only the methods bootstrap needs.
type GitCommitter interface {
	// HasGit checks whether the directory is inside a git repository.
	HasGit(ctx context.Context, projectDir string) (bool, error)
	// StageFiles stages specific file paths for commit.
	StageFiles(ctx context.Context, projectDir string, paths []string) error
	// Commit creates a commit with the given message.
	Commit(ctx context.Context, projectDir string, message string) error
}

// ScaffoldWriter writes the embedded alto-scaffold/ scaffold tree to a target
// project directory, substituting the five template parameters carried by
// ScaffoldParams via text/template data binding. Adapters create the
// `<targetDir>/alto-scaffold/` subtree internally — the caller passes the project
// root, not the alto-scaffold path.
type ScaffoldWriter interface {
	// WriteScaffold renders the embedded scaffold into <targetDir>/alto-scaffold/.
	// When force is false the call fails (wrapped ErrAlreadyExists) if any
	// destination file would be overwritten; when true the existing files
	// are reported via [OVERWRITE] lines before being truncated.
	WriteScaffold(ctx context.Context, targetDir string, params domain.ScaffoldParams, force bool) error
}

// BeadsHookWriter renders the beads `.beads/hooks/post-close` hook (plus
// a Windows .bat shim) into a target project so `bd close` automatically
// fires `alto ticket-ripple` regardless of which session / tool issued
// the close. The hook body is a fixed template constant — primaryTool
// is forwarded for future per-tool overrides but is unused by the
// current adapter.
type BeadsHookWriter interface {
	// WriteBeadsPostCloseHook writes <targetDir>/.beads/hooks/post-close
	// (POSIX, 0o755) and <targetDir>/.beads/hooks/post-close.bat
	// (Windows shim, 0o644). When force is false the call fails with a
	// wrapped ErrAlreadyExists if either hook exists with non-matching
	// content; identical existing content is a no-op. When force is true
	// the existing files are announced via [OVERWRITE] lines on stderr
	// and then truncated and rewritten.
	WriteBeadsPostCloseHook(ctx context.Context, targetDir string, primaryTool string, force bool) error
}

// WorkflowAssetGenerator translates a freshly written alto-scaffold/commands/ tree
// into a tool-native view directory for the given primary tool (e.g.
// "opencode" -> `<targetDir>/.opencode/commands/...`). Defined in the
// Bootstrap context as a thin port so the handler can remain ignorant of
// the ToolTranslation bounded context — composition wires an adapter that
// bridges to the tooltranslation handler. The primaryTool argument is a
// string (not a SupportedTool enum) so the bootstrap layer has no
// compile-time dependency on the tooltranslation domain.
type WorkflowAssetGenerator interface {
	// GenerateForTool reads workflow assets from sourceDir (typically
	// `<projectRoot>/alto-scaffold/commands`) and renders them under projectRoot
	// in the tool-native layout for primaryTool.
	GenerateForTool(ctx context.Context, primaryTool string, sourceDir string, projectRoot string) error
}
