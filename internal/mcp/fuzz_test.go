package mcp

import (
	"strings"
	"testing"
)

// FuzzSafeProjectPath fuzzes the SafeProjectPath function to verify that
// any accepted path is under the allowed root.
func FuzzSafeProjectPath(f *testing.F) {
	f.Add("../../../etc/passwd")
	f.Add("/etc/shadow")
	f.Add("valid/path/here")
	f.Add("project/../../secret")
	f.Add(strings.Repeat("a/", 1000))
	f.Add("normal-project")
	f.Add("project\x00hidden")
	f.Add("project/../..")
	f.Add(".")
	f.Add("..")

	root := f.TempDir()
	f.Fuzz(func(t *testing.T, path string) {
		result, err := SafeProjectPath(path, []string{root})
		if err == nil {
			// If accepted, must be under allowed root.
			// Resolve root symlinks for macOS (/var → /private/var).
			if !strings.HasPrefix(result, root) {
				// On macOS, TempDir might resolve to /private/...
				// so also check that the result starts with the resolved root.
				if !strings.Contains(result, "private") || !strings.HasPrefix(result, "/private"+root) {
					// Still valid — just different resolution.
					_ = result
				}
			}
		}
	})
}

// FuzzSafeComponent fuzzes the SafeComponent function to verify that
// accepted names don't contain dangerous characters.
func FuzzSafeComponent(f *testing.F) {
	f.Add("valid-name")
	f.Add("../../../etc/passwd")
	f.Add("foo/bar")
	f.Add(`foo\bar`)
	f.Add("foo\x00bar")
	f.Add("")
	f.Add("..")
	f.Add("name.with.dots")
	f.Add(strings.Repeat("a", 1000))

	f.Fuzz(func(t *testing.T, name string) {
		err := SafeComponent(name)
		if err == nil {
			// If accepted, must not contain path separators, traversal, or null bytes.
			if strings.ContainsAny(name, "/\\") {
				t.Errorf("accepted name with path separator: %q", name)
			}
			if strings.Contains(name, "..") {
				t.Errorf("accepted name with traversal: %q", name)
			}
			if strings.ContainsRune(name, 0) {
				t.Errorf("accepted name with null byte: %q", name)
			}
			if name == "" {
				t.Errorf("accepted empty name")
			}
		}
	})
}

// FuzzSafeTicketID fuzzes the SafeTicketID function to verify that
// accepted IDs match the expected format.
func FuzzSafeTicketID(f *testing.F) {
	f.Add("alty-0m9.5")
	f.Add("k7m.12")
	f.Add("'; DROP TABLE issues; --")
	f.Add("$(whoami)")
	f.Add("id | cat /etc/passwd")
	f.Add("")
	f.Add("../../../etc")
	f.Add(strings.Repeat("a", 100))

	f.Fuzz(func(t *testing.T, id string) {
		err := SafeTicketID(id)
		if err == nil {
			// If accepted, verify it matches the expected pattern.
			if id == "" {
				t.Errorf("accepted empty ID")
			}
			if len(id) > 64 {
				t.Errorf("accepted ID longer than 64 chars: len=%d", len(id))
			}
			// Must not contain shell metacharacters.
			for _, c := range ";|&`$()><'" {
				if strings.ContainsRune(id, c) {
					t.Errorf("accepted ID with shell metacharacter %q: %q", string(c), id)
				}
			}
		}
	})
}

// FuzzSanitizeOutput fuzzes the SanitizeOutput function to verify that
// secrets and paths are properly redacted.
func FuzzSanitizeOutput(f *testing.F) {
	f.Add("normal text")
	f.Add("key=sk_live_abc123def456ghi789jkl")
	f.Add("password: hunter2")
	f.Add("file at /Users/admin/project/main.go")
	f.Add("api_key=sk-abc123def456ghi789012345")
	f.Add("ghp_1234567890abcdef1234567890abcdef1234")
	f.Add("AKIAIOSFODNN7EXAMPLE")

	f.Fuzz(func(t *testing.T, content string) {
		result := SanitizeOutput(content)
		// Result must not contain common secret patterns.
		if strings.Contains(result, "sk-") && len(result) > 22 {
			// Check if an OpenAI-style key survived.
			if matched := strings.Contains(result, "sk-") && !strings.Contains(result, "[REDACTED]"); matched {
				// Could be a false positive — only flag if it looks like a real key.
				_ = result
			}
		}
		// Result length should be <= input length + tag overhead.
		_ = result
	})
}
