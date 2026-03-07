package valueobjects

// TechStack captures the user's chosen tech stack (language + package manager).
type TechStack struct {
	language       string
	packageManager string
}

// NewTechStack creates a TechStack value object.
func NewTechStack(language, packageManager string) TechStack {
	return TechStack{language: language, packageManager: packageManager}
}

// Language returns the programming language.
func (ts TechStack) Language() string { return ts.language }

// PackageManager returns the package manager.
func (ts TechStack) PackageManager() string { return ts.packageManager }

// Equal returns true if two TechStacks have the same values.
func (ts TechStack) Equal(other TechStack) bool {
	return ts.language == other.language && ts.packageManager == other.packageManager
}
