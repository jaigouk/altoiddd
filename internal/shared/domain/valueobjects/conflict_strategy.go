package valueobjects

// ConflictStrategy defines how file conflicts are resolved when alto output
// would overwrite an existing user file.
type ConflictStrategy string

// Conflict strategy constants.
const (
	// ConflictStrategyRename renames the alto output file to avoid overwriting.
	ConflictStrategyRename ConflictStrategy = "rename"
	// ConflictStrategySkip skips writing the file entirely.
	ConflictStrategySkip ConflictStrategy = "skip"
)

// AllConflictStrategies returns all valid conflict strategy values.
func AllConflictStrategies() []ConflictStrategy {
	return []ConflictStrategy{ConflictStrategyRename, ConflictStrategySkip}
}
