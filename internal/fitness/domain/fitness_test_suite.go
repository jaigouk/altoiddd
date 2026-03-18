package domain

import (
	"fmt"
	"strings"

	domainerrors "github.com/alto-cli/alto/internal/shared/domain/errors"
	"github.com/alto-cli/alto/internal/shared/domain/identity"
	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
)

// BoundedContextInput is the input data for generating fitness tests.
type BoundedContextInput struct {
	Classification *vo.SubdomainClassification
	Name           string
	Responsibility string
}

// SnakeCase converts a PascalCase, space-separated, or hyphenated name to snake_case.
func SnakeCase(name string) string {
	var result []byte
	for i := 0; i < len(name); i++ {
		ch := name[i]
		switch {
		case ch >= 'A' && ch <= 'Z':
			if len(result) > 0 && result[len(result)-1] != '_' {
				prevLower := i > 0 && name[i-1] >= 'a' && name[i-1] <= 'z'
				nextLower := i+1 < len(name) && name[i+1] >= 'a' && name[i+1] <= 'z'
				if prevLower || (nextLower && i > 0 && name[i-1] >= 'A' && name[i-1] <= 'Z') {
					result = append(result, '_')
				}
			}
			result = append(result, ch+32) // toLower
		case ch == ' ' || ch == '-':
			if len(result) > 0 && result[len(result)-1] != '_' {
				result = append(result, '_')
			}
		default:
			result = append(result, ch)
		}
	}
	return string(result)
}

// FitnessTestSuite is the aggregate root for architecture fitness tests.
type FitnessTestSuite struct {
	suiteID     string
	rootPackage string
	contracts   []Contract
	archRules   []ArchRule
	events      []FitnessTestsGenerated
	approved    bool
}

// NewFitnessTestSuite creates a new FitnessTestSuite aggregate root.
func NewFitnessTestSuite(rootPackage string) *FitnessTestSuite {
	return &FitnessTestSuite{
		suiteID:     identity.NewID(),
		rootPackage: rootPackage,
	}
}

// SuiteID returns the unique suite identifier.
func (s *FitnessTestSuite) SuiteID() string { return s.suiteID }

// RootPackage returns the root package name.
func (s *FitnessTestSuite) RootPackage() string { return s.rootPackage }

// Contracts returns a defensive copy.
func (s *FitnessTestSuite) Contracts() []Contract {
	out := make([]Contract, len(s.contracts))
	copy(out, s.contracts)
	return out
}

// ArchRules returns a defensive copy.
func (s *FitnessTestSuite) ArchRules() []ArchRule {
	out := make([]ArchRule, len(s.archRules))
	copy(out, s.archRules)
	return out
}

// Events returns a defensive copy.
func (s *FitnessTestSuite) Events() []FitnessTestsGenerated {
	out := make([]FitnessTestsGenerated, len(s.events))
	copy(out, s.events)
	return out
}

// GenerateContracts generates contracts for each bounded context based on classification.
func (s *FitnessTestSuite) GenerateContracts(bcs []BoundedContextInput) error {
	if s.approved {
		return fmt.Errorf("cannot regenerate contracts on an approved suite: %w",
			domainerrors.ErrInvariantViolation)
	}
	if len(bcs) == 0 {
		return fmt.Errorf("no bounded contexts to generate fitness tests for: %w",
			domainerrors.ErrInvariantViolation)
	}

	s.contracts = nil
	s.archRules = nil

	for _, bc := range bcs {
		if bc.Classification == nil {
			return fmt.Errorf("bounded context '%s' has no subdomain classification: %w",
				bc.Name, domainerrors.ErrInvariantViolation)
		}

		strictness := StrictnessFromClassification(*bc.Classification)
		requiredTypes := RequiredContractTypes(strictness)
		modulePrefix := s.rootPackage + "." + SnakeCase(bc.Name)

		for _, ct := range requiredTypes {
			contract := buildContract(ct, bc.Name, modulePrefix)
			s.contracts = append(s.contracts, contract)
		}

		s.archRules = append(s.archRules, NewArchRule(
			bc.Name+" domain isolation",
			fmt.Sprintf("modules in %s.domain should not import from %s.infrastructure",
				modulePrefix, modulePrefix),
			bc.Name,
		))
	}
	return nil
}

// Preview returns a human-readable preview of generated contracts.
func (s *FitnessTestSuite) Preview() (string, error) {
	if len(s.contracts) == 0 {
		return "", fmt.Errorf("no contracts generated yet — call GenerateContracts() first: %w",
			domainerrors.ErrInvariantViolation)
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Fitness Test Suite: %s\n", s.rootPackage)
	fmt.Fprintf(&b, "Total contracts: %d\n", len(s.contracts))
	fmt.Fprintf(&b, "Total arch rules: %d\n\n", len(s.archRules))

	contextsSeen := make(map[string][]Contract)
	for _, c := range s.contracts {
		contextsSeen[c.contextName] = append(contextsSeen[c.contextName], c)
	}

	for ctxName, contracts := range contextsSeen {
		strictness := inferStrictness(contracts)
		types := make([]string, len(contracts))
		for i, c := range contracts {
			types[i] = string(c.contractType)
		}
		fmt.Fprintf(&b, "  %s (%s):\n", ctxName, strings.ToUpper(string(strictness)))
		fmt.Fprintf(&b, "    Contracts: %s\n\n", strings.Join(types, ", "))
	}

	return b.String(), nil
}

// Approve approves the suite, emitting FitnessTestsGenerated.
func (s *FitnessTestSuite) Approve() error {
	if s.approved {
		return fmt.Errorf("suite already approved: %w", domainerrors.ErrInvariantViolation)
	}
	if len(s.contracts) == 0 {
		return fmt.Errorf("cannot approve suite with no contracts: %w",
			domainerrors.ErrInvariantViolation)
	}

	if err := s.validateModuleBoundaries(); err != nil {
		return err
	}

	s.approved = true
	s.events = append(s.events, NewFitnessTestsGenerated(
		s.suiteID, s.rootPackage,
		s.contracts, s.archRules,
	))
	return nil
}

// RenderImportLinterTOML renders import-linter TOML configuration.
func (s *FitnessTestSuite) RenderImportLinterTOML() (string, error) {
	if len(s.contracts) == 0 {
		return "", fmt.Errorf("no contracts generated yet — call GenerateContracts() first: %w",
			domainerrors.ErrInvariantViolation)
	}

	var b strings.Builder
	b.WriteString("[tool.importlinter]\n")
	fmt.Fprintf(&b, "root_package = \"%s\"\n\n", s.rootPackage)

	for _, c := range s.contracts {
		b.WriteString("[[tool.importlinter.contracts]]\n")
		fmt.Fprintf(&b, "name = \"%s\"\n", c.name)
		fmt.Fprintf(&b, "type = \"%s\"\n", string(c.contractType))

		switch c.contractType {
		case ContractTypeLayers:
			b.WriteString("layers = [\n")
			for _, m := range c.modules {
				fmt.Fprintf(&b, "  \"%s\",\n", m)
			}
			b.WriteString("]\n")
		case ContractTypeForbidden:
			b.WriteString("source_modules = [\n")
			for _, m := range c.modules {
				fmt.Fprintf(&b, "  \"%s\",\n", m)
			}
			b.WriteString("]\n")
			if len(c.forbiddenModules) > 0 {
				b.WriteString("forbidden_modules = [\n")
				for _, m := range c.forbiddenModules {
					fmt.Fprintf(&b, "  \"%s\",\n", m)
				}
				b.WriteString("]\n")
			}
		case ContractTypeIndependence:
			b.WriteString("modules = [\n")
			for _, m := range c.modules {
				fmt.Fprintf(&b, "  \"%s\",\n", m)
			}
			b.WriteString("]\n")
		case ContractTypeAcyclicSiblings:
			if len(c.modules) > 0 {
				fmt.Fprintf(&b, "source_module = \"%s\"\n", c.modules[0])
			}
		}
		b.WriteString("\n")
	}

	return b.String(), nil
}

// RenderArchGoYAML renders arch-go.yml configuration from a BoundedContextMap.
// threshold specifies the compliance percentage (100 for greenfield, 80 for brownfield).
func (s *FitnessTestSuite) RenderArchGoYAML(bcMap *BoundedContextMap, threshold int) (string, error) {
	if bcMap == nil || len(bcMap.Contexts()) == 0 {
		return "", fmt.Errorf("bounded context map is empty: %w", domainerrors.ErrInvariantViolation)
	}

	rootPkg := bcMap.RootPackage()
	var b strings.Builder

	// Header
	b.WriteString("version: 1\n\n")
	b.WriteString("threshold:\n")
	fmt.Fprintf(&b, "  compliance: %d\n", threshold)
	fmt.Fprintf(&b, "  coverage: %d\n\n", threshold)

	b.WriteString("dependenciesRules:\n")

	// Build context name set for cross-context isolation
	contextModules := make(map[string]string) // name -> modulePath
	for _, ctx := range bcMap.Contexts() {
		contextModules[ctx.Name()] = ctx.ModulePath()
	}

	// Build allowed dependencies from relationships
	allowedDeps := make(map[string]map[string]bool) // source -> set of allowed targets
	for _, ctx := range bcMap.Contexts() {
		allowedDeps[ctx.Name()] = make(map[string]bool)
		for _, rel := range ctx.Relationships() {
			if rel.Direction() == RelationshipUpstream {
				// Upstream means this context depends on target
				allowedDeps[ctx.Name()][rel.Target()] = true
			}
		}
	}

	// Helper to build full package path
	fullPkg := func(modulePath, layer string) string {
		if layer == "" {
			return fmt.Sprintf("%s/internal/%s", rootPkg, modulePath)
		}
		return fmt.Sprintf("%s/internal/%s/%s", rootPkg, modulePath, layer)
	}

	// Generate layer rules for each context
	for _, ctx := range bcMap.Contexts() {
		modulePath := ctx.ModulePath()

		// Domain layer: shouldOnlyDependsOn + shouldNotDependsOn
		b.WriteString("  # " + ctx.Name() + " domain layer isolation\n")
		fmt.Fprintf(&b, "  - package: %q\n", fullPkg(modulePath, "domain"))
		b.WriteString("    shouldOnlyDependsOn:\n")
		b.WriteString("      internal:\n")
		fmt.Fprintf(&b, "        - %q\n", fullPkg(modulePath, "domain"))
		fmt.Fprintf(&b, "        - %q\n", fullPkg("shared/domain", ""))
		b.WriteString("      external:\n")
		b.WriteString("        - \"$gostd\"\n")
		b.WriteString("    shouldNotDependsOn:\n")
		b.WriteString("      internal:\n")
		fmt.Fprintf(&b, "        - %q\n", fullPkg(modulePath, "application"))
		fmt.Fprintf(&b, "        - %q\n", fullPkg(modulePath, "infrastructure"))
		b.WriteString("\n")

		// Application layer: shouldNotDependsOn infrastructure
		b.WriteString("  # " + ctx.Name() + " application layer\n")
		fmt.Fprintf(&b, "  - package: %q\n", fullPkg(modulePath, "application"))
		b.WriteString("    shouldNotDependsOn:\n")
		b.WriteString("      internal:\n")
		fmt.Fprintf(&b, "        - %q\n", fullPkg(modulePath, "infrastructure"))
		b.WriteString("\n")
	}

	// Generate cross-context isolation rules
	contexts := bcMap.Contexts()
	for i, srcCtx := range contexts {
		for j, tgtCtx := range contexts {
			if i == j {
				continue
			}
			// Check if srcCtx is allowed to depend on tgtCtx
			if allowedDeps[srcCtx.Name()][tgtCtx.Name()] {
				continue // Relationship allows this dependency
			}
			// Generate isolation rule
			b.WriteString("  # " + srcCtx.Name() + " must not depend on " + tgtCtx.Name() + "\n")
			fmt.Fprintf(&b, "  - package: %q\n", fullPkg(srcCtx.ModulePath(), ""))
			b.WriteString("    shouldNotDependsOn:\n")
			b.WriteString("      internal:\n")
			fmt.Fprintf(&b, "        - %q\n", fullPkg(tgtCtx.ModulePath(), ""))
			b.WriteString("\n")
		}
	}

	return b.String(), nil
}

// RenderPytestarchTests renders pytestarch test file content.
func (s *FitnessTestSuite) RenderPytestarchTests() (string, error) {
	if len(s.contracts) == 0 {
		return "", fmt.Errorf("no contracts generated yet — call GenerateContracts() first: %w",
			domainerrors.ErrInvariantViolation)
	}

	var b strings.Builder
	b.WriteString("\"\"\"Auto-generated architecture fitness tests.\n\n")
	b.WriteString("Generated by alto from bounded context map.\n")
	b.WriteString("\"\"\"\n\n")
	b.WriteString("import pytest\n")
	b.WriteString("from pytestarch import get_evaluable_architecture, Rule\n\n\n")
	fmt.Fprintf(&b, "@pytest.fixture(scope=\"session\")\n")
	fmt.Fprintf(&b, "def evaluable():\n")
	fmt.Fprintf(&b, "    return get_evaluable_architecture(\".\", \"%s\")\n\n", s.rootPackage)

	for _, rule := range s.archRules {
		fnName := "test_" + SnakeCase(rule.name)
		b.WriteString("\n")
		fmt.Fprintf(&b, "def %s(evaluable):\n", fnName)
		fmt.Fprintf(&b, "    \"\"\"%s\"\"\"\n", rule.assertion)

		ctxModule := SnakeCase(rule.contextName)
		domainMod := s.rootPackage + "." + ctxModule + ".domain"
		infraMod := s.rootPackage + "." + ctxModule + ".infrastructure"

		b.WriteString("    rule = (\n")
		b.WriteString("        Rule()\n")
		b.WriteString("        .modules_that()\n")
		fmt.Fprintf(&b, "        .are_named(\"%s\")\n", infraMod)
		b.WriteString("        .should_not()\n")
		b.WriteString("        .be_imported_by_modules_that()\n")
		fmt.Fprintf(&b, "        .are_named(\"%s\")\n", domainMod)
		b.WriteString("    )\n")
		b.WriteString("    rule.assert_applies(evaluable)\n\n")
	}

	return b.String(), nil
}

// -- Private helpers --

func (s *FitnessTestSuite) validateModuleBoundaries() error {
	for _, contract := range s.contracts {
		expectedPrefix := s.rootPackage + "." + SnakeCase(contract.contextName)
		for _, module := range contract.modules {
			if !strings.HasPrefix(module, expectedPrefix) {
				return fmt.Errorf(
					"contract '%s' references module '%s' outside its bounded context '%s' (expected prefix: '%s'): %w",
					contract.name, module, contract.contextName, expectedPrefix,
					domainerrors.ErrInvariantViolation)
			}
		}
		for _, module := range contract.forbiddenModules {
			if !strings.HasPrefix(module, expectedPrefix) {
				return fmt.Errorf(
					"contract '%s' references forbidden module '%s' outside its bounded context '%s' (expected prefix: '%s'): %w",
					contract.name, module, contract.contextName, expectedPrefix,
					domainerrors.ErrInvariantViolation)
			}
		}
	}
	return nil
}

func buildContract(ct ContractType, contextName, modulePrefix string) Contract {
	switch ct {
	case ContractTypeLayers:
		return NewContract(
			contextName+" DDD layer contract",
			ContractTypeLayers,
			contextName,
			[]string{
				modulePrefix + ".infrastructure",
				modulePrefix + ".application",
				modulePrefix + ".domain",
			}, nil)
	case ContractTypeForbidden:
		return NewContract(
			contextName+" domain isolation",
			ContractTypeForbidden,
			contextName,
			[]string{modulePrefix + ".domain"},
			[]string{modulePrefix + ".infrastructure"})
	case ContractTypeIndependence:
		return NewContract(
			contextName+" independence",
			ContractTypeIndependence,
			contextName,
			[]string{modulePrefix}, nil)
	case ContractTypeAcyclicSiblings:
		return NewContract(
			contextName+" acyclic siblings",
			ContractTypeAcyclicSiblings,
			contextName,
			[]string{modulePrefix}, nil)
	default:
		return NewContract(
			contextName+" acyclic siblings",
			ContractTypeAcyclicSiblings,
			contextName,
			[]string{modulePrefix}, nil)
	}
}

func inferStrictness(contracts []Contract) ContractStrictness {
	types := make(map[ContractType]bool)
	for _, c := range contracts {
		types[c.contractType] = true
	}
	if len(types) >= 4 {
		return ContractStrictnessStrict
	}
	if types[ContractTypeLayers] {
		return ContractStrictnessModerate
	}
	return ContractStrictnessMinimal
}
