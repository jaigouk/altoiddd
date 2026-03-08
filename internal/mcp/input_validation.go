// Package mcp provides MCP server utilities including input validation
// helpers for path traversal prevention in MCP tool inputs.
package mcp

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

// SafeComponent validates that a component name is safe — no path separators,
// no traversal sequences, no null bytes, and non-empty.
func SafeComponent(name string) error {
	if name == "" {
		return fmt.Errorf("component name must not be empty")
	}
	if strings.ContainsAny(name, "/\\") {
		return fmt.Errorf("component name must not contain path separators: %q", name)
	}
	if strings.Contains(name, "..") {
		return fmt.Errorf("component name must not contain path traversal: %q", name)
	}
	if strings.ContainsRune(name, 0) {
		return fmt.Errorf("component name must not contain null bytes: %q", name)
	}
	return nil
}

// safeTicketIDPattern matches valid beads ticket IDs: alphanumeric start,
// then alphanumeric, dots, hyphens, or underscores, up to 64 chars.
var safeTicketIDPattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._-]{0,63}$`)

// SafeTicketID validates that a ticket ID matches the expected beads format.
// Rejects shell metacharacters, path traversal, and empty strings.
func SafeTicketID(id string) error {
	if id == "" {
		return fmt.Errorf("ticket ID must not be empty")
	}
	if !safeTicketIDPattern.MatchString(id) {
		return fmt.Errorf("invalid ticket ID format: %q", id)
	}
	return nil
}

// SafeProjectPath resolves a relative path under one of the allowed roots
// and rejects any traversal above the root. Symlinks are resolved to detect
// escape attempts.
func SafeProjectPath(path string, allowedRoots []string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("path must not be empty")
	}
	if len(allowedRoots) == 0 {
		return "", fmt.Errorf("no allowed roots specified")
	}
	if filepath.IsAbs(path) {
		return "", fmt.Errorf("absolute paths are not allowed: %q", path)
	}

	// Try each allowed root.
	for _, root := range allowedRoots {
		absRoot, err := filepath.Abs(root)
		if err != nil {
			continue
		}
		// Resolve symlinks on the root itself.
		absRoot, err = filepath.EvalSymlinks(absRoot)
		if err != nil {
			continue
		}

		candidate := filepath.Join(absRoot, path)

		// Resolve symlinks on the full candidate path to detect escapes.
		resolved, err := filepath.EvalSymlinks(candidate)
		if err != nil {
			// If the path doesn't exist yet, clean it and check prefix.
			resolved = filepath.Clean(candidate)
		}

		// Verify the resolved path is under the root.
		if strings.HasPrefix(resolved, absRoot+string(filepath.Separator)) || resolved == absRoot {
			return resolved, nil
		}
	}

	return "", fmt.Errorf("path %q resolves outside allowed roots", path)
}
