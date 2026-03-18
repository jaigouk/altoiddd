package valueobjects_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
)

func TestTechStackCreation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		lang       string
		pkgMgr     string
		wantLang   string
		wantPkgMgr string
	}{
		{"python uv", "python", "uv", "python", "uv"},
		{"unknown empty", "unknown", "", "unknown", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ts := vo.NewTechStack(tt.lang, tt.pkgMgr)
			assert.Equal(t, tt.wantLang, ts.Language())
			assert.Equal(t, tt.wantPkgMgr, ts.PackageManager())
		})
	}
}

func TestTechStackEquality(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		a     vo.TechStack
		b     vo.TechStack
		equal bool
	}{
		{
			"equal values",
			vo.NewTechStack("python", "uv"),
			vo.NewTechStack("python", "uv"),
			true,
		},
		{
			"different language",
			vo.NewTechStack("python", "uv"),
			vo.NewTechStack("rust", "uv"),
			false,
		},
		{
			"different package manager",
			vo.NewTechStack("python", "uv"),
			vo.NewTechStack("python", "pip"),
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.equal, tt.a.Equal(tt.b))
		})
	}
}
