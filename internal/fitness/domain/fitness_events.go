package domain

import (
	"encoding/json"
	"fmt"
)

// FitnessTestsGenerated is emitted when a FitnessTestSuite is approved.
type FitnessTestsGenerated struct {
	suiteID     string
	rootPackage string
	contracts   []Contract
	archRules   []ArchRule
}

// NewFitnessTestsGenerated creates a FitnessTestsGenerated event.
func NewFitnessTestsGenerated(suiteID, rootPackage string, contracts []Contract, archRules []ArchRule) FitnessTestsGenerated {
	c := make([]Contract, len(contracts))
	copy(c, contracts)
	r := make([]ArchRule, len(archRules))
	copy(r, archRules)
	return FitnessTestsGenerated{
		suiteID:     suiteID,
		rootPackage: rootPackage,
		contracts:   c,
		archRules:   r,
	}
}

// SuiteID returns the suite identifier.
func (e FitnessTestsGenerated) SuiteID() string { return e.suiteID }

// RootPackage returns the root package name.
func (e FitnessTestsGenerated) RootPackage() string { return e.rootPackage }

// Contracts returns a defensive copy.
func (e FitnessTestsGenerated) Contracts() []Contract {
	out := make([]Contract, len(e.contracts))
	copy(out, e.contracts)
	return out
}

// ArchRules returns a defensive copy.
func (e FitnessTestsGenerated) ArchRules() []ArchRule {
	out := make([]ArchRule, len(e.archRules))
	copy(out, e.archRules)
	return out
}

// MarshalJSON implements json.Marshaler for event bus serialization.
func (e FitnessTestsGenerated) MarshalJSON() ([]byte, error) {
	type proxy struct {
		SuiteID     string     `json:"suite_id"`
		RootPackage string     `json:"root_package"`
		Contracts   []Contract `json:"contracts"`
		ArchRules   []ArchRule `json:"arch_rules"`
	}
	data, err := json.Marshal(proxy{
		SuiteID:     e.suiteID,
		RootPackage: e.rootPackage,
		Contracts:   e.contracts,
		ArchRules:   e.archRules,
	})
	if err != nil {
		return nil, fmt.Errorf("marshaling FitnessTestsGenerated: %w", err)
	}
	return data, nil
}

// UnmarshalJSON implements json.Unmarshaler for event bus deserialization.
func (e *FitnessTestsGenerated) UnmarshalJSON(data []byte) error {
	type proxy struct {
		SuiteID     string     `json:"suite_id"`
		RootPackage string     `json:"root_package"`
		Contracts   []Contract `json:"contracts"`
		ArchRules   []ArchRule `json:"arch_rules"`
	}
	var p proxy
	if err := json.Unmarshal(data, &p); err != nil {
		return fmt.Errorf("unmarshaling FitnessTestsGenerated: %w", err)
	}
	e.suiteID = p.SuiteID
	e.rootPackage = p.RootPackage
	e.contracts = p.Contracts
	e.archRules = p.ArchRules
	return nil
}
