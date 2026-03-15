package llm

import (
	"encoding/json"
	"path/filepath"
)

const defaultAnthropicModel = "claude-sonnet-4-20250514"

// DetectedCredential represents an LLM credential found during environment scanning.
type DetectedCredential struct {
	provider Provider
	apiKey   string
	baseURL  string
	model    string
	source   string
}

// NewDetectedCredential creates a DetectedCredential value object.
func NewDetectedCredential(provider Provider, apiKey, baseURL, model, source string) DetectedCredential {
	return DetectedCredential{
		provider: provider,
		apiKey:   apiKey,
		baseURL:  baseURL,
		model:    model,
		source:   source,
	}
}

// Provider returns the LLM provider.
func (d DetectedCredential) Provider() Provider { return d.provider }

// APIKey returns the API key.
func (d DetectedCredential) APIKey() string { return d.apiKey }

// BaseURL returns the base URL for the provider API.
func (d DetectedCredential) BaseURL() string { return d.baseURL }

// Model returns the model name.
func (d DetectedCredential) Model() string { return d.model }

// Source returns where the credential was found.
func (d DetectedCredential) Source() string { return d.source }

// CredentialDetector scans the environment for LLM API credentials.
type CredentialDetector struct {
	envReader  func(string) string
	fileReader func(string) ([]byte, error)
	homeDir    string
}

// NewCredentialDetector creates a CredentialDetector with injectable dependencies.
func NewCredentialDetector(envReader func(string) string, fileReader func(string) ([]byte, error), homeDir string) *CredentialDetector {
	return &CredentialDetector{
		envReader:  envReader,
		fileReader: fileReader,
		homeDir:    homeDir,
	}
}

// Detect scans environment variables and known config files for LLM credentials.
// Returns credentials in priority order: Anthropic env, OpenAI env, Claude JSON file.
func (d *CredentialDetector) Detect() []DetectedCredential {
	var creds []DetectedCredential
	hasAnthropic := false

	// 1. Check ANTHROPIC_API_KEY env var
	if key := d.envReader("ANTHROPIC_API_KEY"); key != "" {
		creds = append(creds, NewDetectedCredential(
			ProviderAnthropic, key, "", defaultAnthropicModel, "env:ANTHROPIC_API_KEY",
		))
		hasAnthropic = true
	}

	// 2. Check OPENAI_API_KEY env var
	if key := d.envReader("OPENAI_API_KEY"); key != "" {
		provider := ProviderOpenAI
		baseURL := d.envReader("OPENAI_BASE_URL")
		if baseURL != "" {
			provider = ProviderOpenAICompatible
		}
		creds = append(creds, NewDetectedCredential(
			provider, key, baseURL, "", "env:OPENAI_API_KEY",
		))
	}

	// 3. Check ~/.claude.json (skip if Anthropic already found via env)
	if !hasAnthropic {
		if cred, ok := d.detectClaudeJSON(); ok {
			creds = append(creds, cred)
		}
	}

	return creds
}

// claudeJSONConfig represents the structure of ~/.claude.json.
type claudeJSONConfig struct {
	APIKey string `json:"apiKey"`
}

func (d *CredentialDetector) detectClaudeJSON() (DetectedCredential, bool) {
	path := filepath.Join(d.homeDir, ".claude.json")
	data, err := d.fileReader(path)
	if err != nil {
		return DetectedCredential{}, false
	}

	var cfg claudeJSONConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return DetectedCredential{}, false
	}

	if cfg.APIKey == "" {
		return DetectedCredential{}, false
	}

	return NewDetectedCredential(
		ProviderAnthropic, cfg.APIKey, "", defaultAnthropicModel, "file:~/.claude.json",
	), true
}
