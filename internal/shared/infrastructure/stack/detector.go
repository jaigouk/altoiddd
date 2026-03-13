// Package stack provides stack detection utilities.
package stack

import (
	"os"
	"path/filepath"

	vo "github.com/alty-cli/alty/internal/shared/domain/valueobjects"
)

// DetectProfile detects the project stack from manifest files.
// Returns GoModProfile for Go projects, PythonUvProfile for Python, GenericProfile otherwise.
// If projectDir is empty, uses the current working directory.
func DetectProfile(projectDir string) vo.StackProfile {
	if projectDir == "" {
		var err error
		projectDir, err = os.Getwd()
		if err != nil {
			return vo.GenericProfile{}
		}
	}

	// Go takes precedence
	if _, err := os.Stat(filepath.Join(projectDir, "go.mod")); err == nil {
		return vo.GoModProfile{}
	}

	// Then Python
	if _, err := os.Stat(filepath.Join(projectDir, "pyproject.toml")); err == nil {
		return vo.PythonUvProfile{}
	}

	// Generic fallback
	return vo.GenericProfile{}
}
