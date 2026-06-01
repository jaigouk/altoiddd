package domain

import "errors"

// ErrUnsafeTemplateParameter indicates a scaffold-template parameter contains
// characters that could compromise shell-quoting or path-traversal safety.
// Returned only from value-object constructors in the bootstrap domain layer;
// callers wrap with the higher-level ErrInvariantViolation as appropriate.
var ErrUnsafeTemplateParameter = errors.New("unsafe template parameter")
