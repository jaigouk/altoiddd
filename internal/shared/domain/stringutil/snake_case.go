// Package stringutil provides string manipulation utilities for domain operations.
package stringutil

import (
	"regexp"
	"strings"
	"unicode"
)

// Precompiled regexes for ToSnakeCase.
var (
	matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
	matchAllCap   = regexp.MustCompile("([a-z0-9])([A-Z])")
)

// ToSnakeCase converts PascalCase, camelCase, or "Space Separated" strings to snake_case.
//
// Examples:
//   - "OrderManagement" -> "order_management"
//   - "Order Processing" -> "order_processing"
//   - "APIClient" -> "api_client"
//   - "userID" -> "user_id"
func ToSnakeCase(s string) string {
	// Handle PascalCase/camelCase first
	snake := matchFirstCap.ReplaceAllString(s, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")

	// Then handle spaces and hyphens
	snake = strings.ReplaceAll(snake, " ", "_")
	snake = strings.ReplaceAll(snake, "-", "_")

	// Clean up any double underscores that might have been created
	for strings.Contains(snake, "__") {
		snake = strings.ReplaceAll(snake, "__", "_")
	}

	return strings.ToLower(snake)
}

// ToSnakeCaseSimple is a simpler implementation that handles PascalCase and spaces
// but may produce suboptimal results for consecutive uppercase letters.
// Use ToSnakeCase for better handling of acronyms like "APIClient".
func ToSnakeCaseSimple(s string) string {
	var result strings.Builder
	for i, r := range s {
		if r == ' ' || r == '-' {
			result.WriteRune('_')
			continue
		}
		if unicode.IsUpper(r) {
			if i > 0 {
				prev := rune(s[i-1])
				if prev != ' ' && prev != '-' && !unicode.IsUpper(prev) {
					result.WriteRune('_')
				}
			}
			result.WriteRune(unicode.ToLower(r))
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}
