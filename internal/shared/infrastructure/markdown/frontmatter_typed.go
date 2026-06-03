package markdown

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// ParseGeneric unmarshals raw YAML frontmatter into a generic
// map[string]any. Wraps yaml.v3 errors with consistent context so callers
// can distinguish unmarshal failures from extraction failures.
func ParseGeneric(raw string) (map[string]any, error) {
	out := map[string]any{}
	if err := yaml.Unmarshal([]byte(raw), &out); err != nil {
		return nil, fmt.Errorf("unmarshalling frontmatter: %w", err)
	}
	return out, nil
}

// ParseTyped unmarshals raw YAML frontmatter into the caller's typed
// struct. Same error-wrapping contract as ParseGeneric.
func ParseTyped[T any](raw string) (T, error) {
	var out T
	if err := yaml.Unmarshal([]byte(raw), &out); err != nil {
		return out, fmt.Errorf("unmarshalling frontmatter: %w", err)
	}
	return out, nil
}
