// Package stack provides stack detection utilities.
package stack

import (
	"os"
	"path/filepath"
	"strings"

	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
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

// ExtractLanguageFromText scans free-form text (e.g. a README) for stack-language
// signals and returns the canonical language identifier "go" or "python", or "" if
// no signal is found.
//
// Precedence is intentional and deterministic: Go-group keywords are checked before
// Python-group keywords, and the first match wins. This guarantees that a polyglot
// description like "golang services and python scripts" resolves to Go — the host
// language of the alto tooling and the more common DDD-CLI choice.
//
// Matching is case-insensitive (text is lowercased once up front).
func ExtractLanguageFromText(text string) string {
	keywords := []struct{ keyword, lang string }{
		// Go group — checked first (yields "go" on first match in this group)
		{"go.mod", "go"},
		{"golang", "go"},
		{" go ", "go"},
		{"package main", "go"},
		// Python group
		{"python", "python"},
		{"pyproject", "python"},
		{"pip install", "python"},
		{".py", "python"},
	}

	lower := strings.ToLower(text)
	for _, kw := range keywords {
		if strings.Contains(lower, kw.keyword) {
			return kw.lang
		}
	}
	return ""
}
