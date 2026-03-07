package domain

import vo "github.com/alty-cli/alty/internal/shared/domain/valueobjects"

// ResolveProfile maps a TechStack to the corresponding StackProfile.
// Returns PythonUvProfile for "python", GenericProfile otherwise.
func ResolveProfile(techStack *vo.TechStack) vo.StackProfile {
	if techStack != nil && techStack.Language() == "python" {
		return vo.PythonUvProfile{}
	}
	return vo.GenericProfile{}
}
