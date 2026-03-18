package domain

import (
	"encoding/json"
	"fmt"

	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
)

// ContractType enumerates import-linter contract types.
type ContractType string

// Contract type constants.
const (
	ContractTypeLayers          ContractType = "layers"
	ContractTypeForbidden       ContractType = "forbidden"
	ContractTypeIndependence    ContractType = "independence"
	ContractTypeAcyclicSiblings ContractType = "acyclic_siblings"
)

// ContractStrictness maps SubdomainClassification to enforcement level.
type ContractStrictness string

// Contract strictness constants.
const (
	ContractStrictnessStrict   ContractStrictness = "strict"
	ContractStrictnessModerate ContractStrictness = "moderate"
	ContractStrictnessMinimal  ContractStrictness = "minimal"
)

// StrictnessFromClassification maps a SubdomainClassification to its ContractStrictness.
func StrictnessFromClassification(cl vo.SubdomainClassification) ContractStrictness {
	switch cl {
	case vo.SubdomainCore:
		return ContractStrictnessStrict
	case vo.SubdomainSupporting:
		return ContractStrictnessModerate
	case vo.SubdomainGeneric:
		return ContractStrictnessMinimal
	default:
		return ContractStrictnessMinimal
	}
}

// RequiredContractTypes returns the contract types required for a strictness level.
func RequiredContractTypes(s ContractStrictness) []ContractType {
	switch s {
	case ContractStrictnessStrict:
		return []ContractType{
			ContractTypeLayers,
			ContractTypeForbidden,
			ContractTypeIndependence,
			ContractTypeAcyclicSiblings,
		}
	case ContractStrictnessModerate:
		return []ContractType{ContractTypeLayers, ContractTypeForbidden}
	case ContractStrictnessMinimal:
		return []ContractType{ContractTypeForbidden}
	default:
		return []ContractType{ContractTypeForbidden}
	}
}

// Contract is an import-linter contract for architecture boundary enforcement.
type Contract struct {
	name             string
	contractType     ContractType
	contextName      string
	modules          []string
	forbiddenModules []string
}

// NewContract creates a Contract value object.
func NewContract(name string, ct ContractType, contextName string, modules, forbiddenModules []string) Contract {
	m := make([]string, len(modules))
	copy(m, modules)
	fm := make([]string, len(forbiddenModules))
	copy(fm, forbiddenModules)
	return Contract{
		name:             name,
		contractType:     ct,
		contextName:      contextName,
		modules:          m,
		forbiddenModules: fm,
	}
}

// Name returns the contract name.
func (c Contract) Name() string { return c.name }

// ContractType returns the contract type.
func (c Contract) ContractType() ContractType { return c.contractType }

// ContextName returns the bounded context name.
func (c Contract) ContextName() string { return c.contextName }

// Modules returns a defensive copy.
func (c Contract) Modules() []string {
	out := make([]string, len(c.modules))
	copy(out, c.modules)
	return out
}

// ForbiddenModules returns a defensive copy.
func (c Contract) ForbiddenModules() []string {
	out := make([]string, len(c.forbiddenModules))
	copy(out, c.forbiddenModules)
	return out
}

// MarshalJSON implements json.Marshaler for event bus serialization.
func (c Contract) MarshalJSON() ([]byte, error) {
	type proxy struct {
		Name             string       `json:"name"`
		ContractType     ContractType `json:"contract_type"`
		ContextName      string       `json:"context_name"`
		Modules          []string     `json:"modules"`
		ForbiddenModules []string     `json:"forbidden_modules"`
	}
	data, err := json.Marshal(proxy{
		Name:             c.name,
		ContractType:     c.contractType,
		ContextName:      c.contextName,
		Modules:          c.modules,
		ForbiddenModules: c.forbiddenModules,
	})
	if err != nil {
		return nil, fmt.Errorf("marshaling Contract: %w", err)
	}
	return data, nil
}

// UnmarshalJSON implements json.Unmarshaler for event bus deserialization.
func (c *Contract) UnmarshalJSON(data []byte) error {
	type proxy struct {
		Name             string       `json:"name"`
		ContractType     ContractType `json:"contract_type"`
		ContextName      string       `json:"context_name"`
		Modules          []string     `json:"modules"`
		ForbiddenModules []string     `json:"forbidden_modules"`
	}
	var p proxy
	if err := json.Unmarshal(data, &p); err != nil {
		return fmt.Errorf("unmarshaling Contract: %w", err)
	}
	c.name = p.Name
	c.contractType = p.ContractType
	c.contextName = p.ContextName
	c.modules = p.Modules
	c.forbiddenModules = p.ForbiddenModules
	return nil
}

// ArchRule is a pytestarch rule for architecture boundary enforcement.
type ArchRule struct {
	name        string
	assertion   string
	contextName string
}

// NewArchRule creates an ArchRule value object.
func NewArchRule(name, assertion, contextName string) ArchRule {
	return ArchRule{name: name, assertion: assertion, contextName: contextName}
}

// Name returns the rule name.
func (r ArchRule) Name() string { return r.name }

// Assertion returns the rule assertion text.
func (r ArchRule) Assertion() string { return r.assertion }

// ContextName returns the bounded context name.
func (r ArchRule) ContextName() string { return r.contextName }

// MarshalJSON implements json.Marshaler for event bus serialization.
func (r ArchRule) MarshalJSON() ([]byte, error) {
	type proxy struct {
		Name        string `json:"name"`
		Assertion   string `json:"assertion"`
		ContextName string `json:"context_name"`
	}
	data, err := json.Marshal(proxy{
		Name:        r.name,
		Assertion:   r.assertion,
		ContextName: r.contextName,
	})
	if err != nil {
		return nil, fmt.Errorf("marshaling ArchRule: %w", err)
	}
	return data, nil
}

// UnmarshalJSON implements json.Unmarshaler for event bus deserialization.
func (r *ArchRule) UnmarshalJSON(data []byte) error {
	type proxy struct {
		Name        string `json:"name"`
		Assertion   string `json:"assertion"`
		ContextName string `json:"context_name"`
	}
	var p proxy
	if err := json.Unmarshal(data, &p); err != nil {
		return fmt.Errorf("unmarshaling ArchRule: %w", err)
	}
	r.name = p.Name
	r.assertion = p.Assertion
	r.contextName = p.ContextName
	return nil
}
