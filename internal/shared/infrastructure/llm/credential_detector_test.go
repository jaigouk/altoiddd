package llm_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alty-cli/alty/internal/shared/infrastructure/llm"
)

// failReader always returns an error, simulating missing files.
func failReader(_ string) ([]byte, error) {
	return nil, fmt.Errorf("file not found")
}

// emptyEnv always returns empty string, simulating no env vars set.
func emptyEnv(_ string) string { return "" }

// ---------------------------------------------------------------------------
// DetectedCredential VO
// ---------------------------------------------------------------------------

func TestDetectedCredential_Accessors(t *testing.T) {
	t.Parallel()
	cred := llm.NewDetectedCredential(llm.ProviderAnthropic, "sk-key", "https://base.url", "model-1", "env:FOO")
	assert.Equal(t, llm.ProviderAnthropic, cred.Provider())
	assert.Equal(t, "sk-key", cred.APIKey())
	assert.Equal(t, "https://base.url", cred.BaseURL())
	assert.Equal(t, "model-1", cred.Model())
	assert.Equal(t, "env:FOO", cred.Source())
}

// ---------------------------------------------------------------------------
// CredentialDetector.Detect
// ---------------------------------------------------------------------------

func TestCredentialDetector_Detect_WhenAnthropicKeyInEnv_ExpectAnthropicCredential(t *testing.T) {
	t.Parallel()
	envReader := func(key string) string {
		if key == "ANTHROPIC_API_KEY" {
			return "sk-ant-test"
		}
		return ""
	}
	detector := llm.NewCredentialDetector(envReader, failReader, "/home/test")
	creds := detector.Detect()
	require.Len(t, creds, 1)
	assert.Equal(t, llm.ProviderAnthropic, creds[0].Provider())
	assert.Equal(t, "sk-ant-test", creds[0].APIKey())
	assert.Empty(t, creds[0].BaseURL())
	assert.Equal(t, "claude-sonnet-4-20250514", creds[0].Model())
	assert.Equal(t, "env:ANTHROPIC_API_KEY", creds[0].Source())
}

func TestCredentialDetector_Detect_WhenOpenAIKeyInEnv_ExpectOpenAICredential(t *testing.T) {
	t.Parallel()
	envReader := func(key string) string {
		if key == "OPENAI_API_KEY" {
			return "sk-openai-test"
		}
		return ""
	}
	detector := llm.NewCredentialDetector(envReader, failReader, "/home/test")
	creds := detector.Detect()
	require.Len(t, creds, 1)
	assert.Equal(t, llm.ProviderOpenAI, creds[0].Provider())
	assert.Equal(t, "sk-openai-test", creds[0].APIKey())
	assert.Empty(t, creds[0].BaseURL())
	assert.Empty(t, creds[0].Model())
	assert.Equal(t, "env:OPENAI_API_KEY", creds[0].Source())
}

func TestCredentialDetector_Detect_WhenOpenAIKeyWithBaseURL_ExpectOpenAICompatible(t *testing.T) {
	t.Parallel()
	envReader := func(key string) string {
		switch key {
		case "OPENAI_API_KEY":
			return "sk-openai-test"
		case "OPENAI_BASE_URL":
			return "https://custom.api.com/v1"
		}
		return ""
	}
	detector := llm.NewCredentialDetector(envReader, failReader, "/home/test")
	creds := detector.Detect()
	require.Len(t, creds, 1)
	assert.Equal(t, llm.ProviderOpenAICompatible, creds[0].Provider())
	assert.Equal(t, "sk-openai-test", creds[0].APIKey())
	assert.Equal(t, "https://custom.api.com/v1", creds[0].BaseURL())
	assert.Empty(t, creds[0].Model())
	assert.Equal(t, "env:OPENAI_API_KEY", creds[0].Source())
}

func TestCredentialDetector_Detect_WhenClaudeJSONExists_ExpectAnthropicCredential(t *testing.T) {
	t.Parallel()
	fileReader := func(path string) ([]byte, error) {
		if path == "/home/test/.claude.json" {
			return []byte(`{"apiKey":"sk-ant-from-file"}`), nil
		}
		return nil, fmt.Errorf("not found")
	}
	detector := llm.NewCredentialDetector(emptyEnv, fileReader, "/home/test")
	creds := detector.Detect()
	require.Len(t, creds, 1)
	assert.Equal(t, llm.ProviderAnthropic, creds[0].Provider())
	assert.Equal(t, "sk-ant-from-file", creds[0].APIKey())
	assert.Equal(t, "claude-sonnet-4-20250514", creds[0].Model())
	assert.Equal(t, "file:~/.claude.json", creds[0].Source())
}

func TestCredentialDetector_Detect_WhenMultipleSources_ExpectAllReturned(t *testing.T) {
	t.Parallel()
	envReader := func(key string) string {
		switch key {
		case "ANTHROPIC_API_KEY":
			return "sk-ant-env"
		case "OPENAI_API_KEY":
			return "sk-openai-env"
		}
		return ""
	}
	detector := llm.NewCredentialDetector(envReader, failReader, "/home/test")
	creds := detector.Detect()
	require.Len(t, creds, 2)
	assert.Equal(t, llm.ProviderAnthropic, creds[0].Provider())
	assert.Equal(t, llm.ProviderOpenAI, creds[1].Provider())
}

func TestCredentialDetector_Detect_WhenNoCredentials_ExpectEmptySlice(t *testing.T) {
	t.Parallel()
	detector := llm.NewCredentialDetector(emptyEnv, failReader, "/home/test")
	creds := detector.Detect()
	assert.Empty(t, creds)
}

func TestCredentialDetector_Detect_WhenClaudeJSONMalformed_ExpectSkippedSilently(t *testing.T) {
	t.Parallel()
	fileReader := func(path string) ([]byte, error) {
		if path == "/home/test/.claude.json" {
			return []byte(`{not valid json`), nil
		}
		return nil, fmt.Errorf("not found")
	}
	detector := llm.NewCredentialDetector(emptyEnv, fileReader, "/home/test")
	creds := detector.Detect()
	assert.Empty(t, creds)
}

func TestCredentialDetector_Detect_WhenClaudeJSONMissing_ExpectSkippedSilently(t *testing.T) {
	t.Parallel()
	detector := llm.NewCredentialDetector(emptyEnv, failReader, "/home/test")
	creds := detector.Detect()
	assert.Empty(t, creds)
}

func TestCredentialDetector_Detect_WhenBothEnvAndClaudeJSONForAnthropic_ExpectDedup(t *testing.T) {
	t.Parallel()
	envReader := func(key string) string {
		if key == "ANTHROPIC_API_KEY" {
			return "sk-ant-env"
		}
		return ""
	}
	fileReader := func(path string) ([]byte, error) {
		if path == "/home/test/.claude.json" {
			return []byte(`{"apiKey":"sk-ant-from-file"}`), nil
		}
		return nil, fmt.Errorf("not found")
	}
	detector := llm.NewCredentialDetector(envReader, fileReader, "/home/test")
	creds := detector.Detect()
	// Only env credential, Claude JSON skipped due to dedup
	require.Len(t, creds, 1)
	assert.Equal(t, "sk-ant-env", creds[0].APIKey())
	assert.Equal(t, "env:ANTHROPIC_API_KEY", creds[0].Source())
}

func TestCredentialDetector_Detect_WhenClaudeJSONEmptyAPIKey_ExpectSkipped(t *testing.T) {
	t.Parallel()
	fileReader := func(path string) ([]byte, error) {
		if path == "/home/test/.claude.json" {
			return []byte(`{"apiKey":""}`), nil
		}
		return nil, fmt.Errorf("not found")
	}
	detector := llm.NewCredentialDetector(emptyEnv, fileReader, "/home/test")
	creds := detector.Detect()
	assert.Empty(t, creds)
}
