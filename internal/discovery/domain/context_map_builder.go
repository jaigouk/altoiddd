package domain

import (
	"fmt"
	"strings"
)

// BuildContextMap infers ContextRelationships from shared work objects across
// sketch pairs and returns a complete ContextMap. This is a pure function
// following the ComputeBoundaryScore pattern.
func BuildContextMap(
	projectName string,
	sketches []BoundedContextSketch,
) (*ContextMap, error) {
	var relationships []ContextRelationship

	for i := 0; i < len(sketches); i++ {
		for j := i + 1; j < len(sketches); j++ {
			if sketches[i].Name() == sketches[j].Name() {
				continue
			}

			shared := intersectWorkObjects(sketches[i].WorkObjects(), sketches[j].WorkObjects())
			if len(shared) == 0 {
				continue
			}

			upstream, downstream := sketches[i].Name(), sketches[j].Name()
			if upstream > downstream {
				upstream, downstream = downstream, upstream
			}

			description := fmt.Sprintf(
				"Shared work objects (alphabetical ordering, no true directionality): %s",
				strings.Join(shared, ", "),
			)

			rel, err := NewContextRelationship(upstream, downstream, RelationshipTypeSharedKernel, shared, description)
			if err != nil {
				return nil, fmt.Errorf("creating relationship between %q and %q: %w", upstream, downstream, err)
			}

			relationships = append(relationships, rel)
		}
	}

	cm, err := NewContextMap(projectName, sketches, relationships)
	if err != nil {
		return nil, fmt.Errorf("creating context map: %w", err)
	}

	return cm, nil
}

// intersectWorkObjects returns work objects present in both slices,
// using case-insensitive comparison. Items are returned in their original
// case from the first slice.
func intersectWorkObjects(a, b []string) []string {
	bLower := make(map[string]struct{}, len(b))
	for _, item := range b {
		bLower[strings.ToLower(item)] = struct{}{}
	}

	var shared []string

	for _, item := range a {
		if _, ok := bLower[strings.ToLower(item)]; ok {
			shared = append(shared, item)
		}
	}

	return shared
}
