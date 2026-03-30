package domain

import (
	"strings"

	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
)

// ResolveProfile maps a TechStack to the corresponding StackProfile.
// Uses case-insensitive matching. Returns PythonUvProfile for "python",
// GoModProfile for "go", GenericProfile otherwise.
func ResolveProfile(techStack *vo.TechStack) vo.StackProfile {
	if techStack == nil {
		return vo.GenericProfile{}
	}

	switch strings.ToLower(techStack.Language()) {
	case "python":
		return vo.PythonUvProfile{}
	case "go":
		return vo.GoModProfile{}
	default:
		return vo.GenericProfile{}
	}
}
