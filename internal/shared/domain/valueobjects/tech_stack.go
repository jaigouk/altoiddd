package valueobjects

import (
	"encoding/json"
	"fmt"
)

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

// MarshalJSON implements json.Marshaler for event bus serialization.
func (ts TechStack) MarshalJSON() ([]byte, error) {
	type proxy struct {
		Language       string `json:"language"`
		PackageManager string `json:"package_manager"`
	}
	data, err := json.Marshal(proxy{
		Language:       ts.language,
		PackageManager: ts.packageManager,
	})
	if err != nil {
		return nil, fmt.Errorf("marshaling TechStack: %w", err)
	}
	return data, nil
}

// UnmarshalJSON implements json.Unmarshaler for event bus deserialization.
func (ts *TechStack) UnmarshalJSON(data []byte) error {
	type proxy struct {
		Language       string `json:"language"`
		PackageManager string `json:"package_manager"`
	}
	var p proxy
	if err := json.Unmarshal(data, &p); err != nil {
		return fmt.Errorf("unmarshaling TechStack: %w", err)
	}
	ts.language = p.Language
	ts.packageManager = p.PackageManager
	return nil
}
