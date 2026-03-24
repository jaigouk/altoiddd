package domain

import (
	"fmt"

	domainerrors "github.com/alto-cli/alto/internal/shared/domain/errors"
)

// RelationshipType classifies the strategic relationship between two bounded contexts.
type RelationshipType string

// RelationshipType constants.
const (
	RelationshipTypeSharedKernel        RelationshipType = "shared_kernel"
	RelationshipTypeCustomerSupplier    RelationshipType = "customer_supplier"
	RelationshipTypeConformist          RelationshipType = "conformist"
	RelationshipTypeAnticorruptionLayer RelationshipType = "anticorruption_layer"
	RelationshipTypeOpenHostService     RelationshipType = "open_host_service"
	RelationshipTypePublishedLanguage   RelationshipType = "published_language"
	RelationshipTypePartnership         RelationshipType = "partnership"
	RelationshipTypeSeparateWays        RelationshipType = "separate_ways"
)

var validRelationshipTypes = map[RelationshipType]struct{}{
	RelationshipTypeSharedKernel:        {},
	RelationshipTypeCustomerSupplier:    {},
	RelationshipTypeConformist:          {},
	RelationshipTypeAnticorruptionLayer: {},
	RelationshipTypeOpenHostService:     {},
	RelationshipTypePublishedLanguage:   {},
	RelationshipTypePartnership:         {},
	RelationshipTypeSeparateWays:        {},
}

// NewRelationshipType creates a RelationshipType from a string, returning an error if invalid.
func NewRelationshipType(s string) (RelationshipType, error) {
	rt := RelationshipType(s)
	if err := rt.Validate(); err != nil {
		return "", err
	}

	return rt, nil
}

// AllRelationshipTypes returns all valid RelationshipType values.
func AllRelationshipTypes() []RelationshipType {
	return []RelationshipType{
		RelationshipTypeSharedKernel,
		RelationshipTypeCustomerSupplier,
		RelationshipTypeConformist,
		RelationshipTypeAnticorruptionLayer,
		RelationshipTypeOpenHostService,
		RelationshipTypePublishedLanguage,
		RelationshipTypePartnership,
		RelationshipTypeSeparateWays,
	}
}

// String returns the string representation of a RelationshipType.
func (r RelationshipType) String() string {
	return string(r)
}

// Validate checks whether the RelationshipType holds a valid value.
func (r RelationshipType) Validate() error {
	if _, ok := validRelationshipTypes[r]; !ok {
		return fmt.Errorf("invalid relationship type %q: %w", string(r), domainerrors.ErrInvariantViolation)
	}

	return nil
}

// MarshalText implements encoding.TextMarshaler.
func (r RelationshipType) MarshalText() ([]byte, error) {
	if err := r.Validate(); err != nil {
		return nil, fmt.Errorf("marshaling relationship type: %w", err)
	}

	return []byte(r), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (r *RelationshipType) UnmarshalText(data []byte) error {
	parsed, err := NewRelationshipType(string(data))
	if err != nil {
		return err
	}

	*r = parsed

	return nil
}
