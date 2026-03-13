package valueobjects

import "strings"

// StackProfile is the strategy interface providing stack-specific knowledge.
type StackProfile interface {
	StackID() string
	FileGlob() string
	ProjectManifest() string
	SourceLayout() []string
	QualityGateCommands() map[QualityGate][]string
	QualityGateDisplay() string
	FitnessAvailable() bool
	ToRootPackage(projectName string) string
}

// PythonUvProfile provides full Python+uv pipeline values.
type PythonUvProfile struct{}

// StackID returns the stack identifier.
func (p PythonUvProfile) StackID() string { return "python-uv" }

// FileGlob returns the file glob pattern.
func (p PythonUvProfile) FileGlob() string { return "**/*.py" }

// ProjectManifest returns the project manifest filename.
func (p PythonUvProfile) ProjectManifest() string { return "pyproject.toml" }

// SourceLayout returns the DDD source layout directories.
func (p PythonUvProfile) SourceLayout() []string {
	return []string{"src/domain/", "src/application/", "src/infrastructure/"}
}

// QualityGateCommands returns commands for each quality gate.
func (p PythonUvProfile) QualityGateCommands() map[QualityGate][]string {
	return map[QualityGate][]string{
		QualityGateLint:    {"uv", "run", "ruff", "check", "."},
		QualityGateTypes:   {"uv", "run", "mypy", "."},
		QualityGateTests:   {"uv", "run", "pytest"},
		QualityGateFitness: {"uv", "run", "pytest", "tests/architecture/"},
	}
}

// QualityGateDisplay returns the markdown display for quality gates.
func (p PythonUvProfile) QualityGateDisplay() string {
	return "## Quality Gates\n" +
		"\n" +
		"```bash\n" +
		"uv run ruff check .              # Lint\n" +
		"uv run mypy .                    # Type check\n" +
		"uv run pytest                    # Tests\n" +
		"```\n"
}

// FitnessAvailable returns whether fitness tests are available.
func (p PythonUvProfile) FitnessAvailable() bool { return true }

// ToRootPackage converts a project name to a root package name.
func (p PythonUvProfile) ToRootPackage(projectName string) string {
	return strings.ReplaceAll(projectName, "-", "_")
}

// GenericProfile is a fallback profile for unknown stacks.
type GenericProfile struct{}

// StackID returns the stack identifier.
func (g GenericProfile) StackID() string { return "generic" }

// FileGlob returns the file glob pattern.
func (g GenericProfile) FileGlob() string { return "*" }

// ProjectManifest returns the project manifest filename.
func (g GenericProfile) ProjectManifest() string { return "" }

// SourceLayout returns the source layout directories.
func (g GenericProfile) SourceLayout() []string { return nil }

// QualityGateCommands returns commands for each quality gate.
func (g GenericProfile) QualityGateCommands() map[QualityGate][]string { return nil }

// QualityGateDisplay returns the markdown display for quality gates.
func (g GenericProfile) QualityGateDisplay() string { return "" }

// FitnessAvailable returns whether fitness tests are available.
func (g GenericProfile) FitnessAvailable() bool { return false }

// ToRootPackage converts a project name to a root package name.
func (g GenericProfile) ToRootPackage(projectName string) string { return projectName }

// GoModProfile provides full Go modules pipeline values.
type GoModProfile struct{}

// StackID returns the stack identifier.
func (g GoModProfile) StackID() string { return "go-mod" }

// FileGlob returns the file glob pattern.
func (g GoModProfile) FileGlob() string { return "**/*.go" }

// ProjectManifest returns the project manifest filename.
func (g GoModProfile) ProjectManifest() string { return "go.mod" }

// SourceLayout returns the DDD source layout directories.
func (g GoModProfile) SourceLayout() []string {
	return []string{"internal/*/domain/", "internal/*/application/", "internal/*/infrastructure/"}
}

// QualityGateCommands returns commands for each quality gate.
func (g GoModProfile) QualityGateCommands() map[QualityGate][]string {
	return map[QualityGate][]string{
		QualityGateLint:    {"golangci-lint", "run", "./..."},
		QualityGateTypes:   {"go", "vet", "./..."},
		QualityGateTests:   {"go", "test", "-race", "./..."},
		QualityGateFitness: {"arch-go"}, // MIT-licensed architecture testing tool
	}
}

// QualityGateDisplay returns the markdown display for quality gates.
func (g GoModProfile) QualityGateDisplay() string {
	return "## Quality Gates\n" +
		"\n" +
		"```bash\n" +
		"golangci-lint run ./...          # Lint\n" +
		"go vet ./...                     # Type check\n" +
		"go test -race ./...              # Tests\n" +
		"arch-go                          # Architecture fitness\n" +
		"```\n"
}

// FitnessAvailable returns whether fitness tests are available.
func (g GoModProfile) FitnessAvailable() bool { return true }

// ToRootPackage converts a project name to a root package name.
func (g GoModProfile) ToRootPackage(projectName string) string { return projectName }

// Compile-time interface checks.
var (
	_ StackProfile = PythonUvProfile{}
	_ StackProfile = GenericProfile{}
	_ StackProfile = GoModProfile{}
)
