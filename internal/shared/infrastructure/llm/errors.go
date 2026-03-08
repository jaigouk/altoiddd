package llm

import "errors"

// ErrLLMUnavailable is returned when the LLM service is not configured or unreachable.
var ErrLLMUnavailable = errors.New("LLM unavailable")
