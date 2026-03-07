package domain

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
