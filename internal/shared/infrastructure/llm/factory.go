package llm

// Factory creates the appropriate Client based on Config.
type Factory struct{}

// Create returns an LLM client for the given configuration.
// Degradation paths:
//   - ProviderNone → NoopClient
//   - ProviderAnthropic + empty API key → NoopClient
//   - Unsupported provider → NoopClient
func (f Factory) Create(config Config) Client {
	switch config.Provider() {
	case ProviderAnthropic:
		if config.APIKey() == "" {
			return &NoopClient{}
		}
		return NewAnthropicClient(config.APIKey(), config.Model(), config.Timeout())
	case ProviderOllama, ProviderVertexAI:
		return &NoopClient{}
	case ProviderNone:
		return &NoopClient{}
	}
	return &NoopClient{}
}
