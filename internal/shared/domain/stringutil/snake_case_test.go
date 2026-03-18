package stringutil_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/alto-cli/alto/internal/shared/domain/stringutil"
)

func TestToSnakeCase(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"PascalCase", "OrderManagement", "order_management"},
		{"camelCase", "orderManagement", "order_management"},
		{"Space Separated", "Order Processing", "order_processing"},
		{"Hyphenated", "order-processing", "order_processing"},
		{"Single word", "Orders", "orders"},
		{"Lowercase", "orders", "orders"},
		{"Acronym at start", "APIClient", "api_client"},
		{"Acronym in middle", "userIDField", "user_id_field"},
		{"Multiple words", "OrderManagementSystem", "order_management_system"},
		{"Empty string", "", ""},
		{"Already snake_case", "order_management", "order_management"},
		{"Mixed spaces and caps", "Order Management System", "order_management_system"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := stringutil.ToSnakeCase(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestToSnakeCaseSimple(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"PascalCase", "OrderManagement", "order_management"},
		{"Space Separated", "Order Processing", "order_processing"},
		{"Hyphenated", "order-processing", "order_processing"},
		{"Single word", "Orders", "orders"},
		{"Empty string", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := stringutil.ToSnakeCaseSimple(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
