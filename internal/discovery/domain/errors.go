package domain

import "errors"

// ErrInferenceDismissed indicates the user declined an inference result during
// the --existing flow. The caller catches this with errors.Is and falls through
// to the normal storytelling discovery path.
var ErrInferenceDismissed = errors.New("user dismissed inference result")
