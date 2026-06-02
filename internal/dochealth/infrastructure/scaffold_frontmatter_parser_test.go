package infrastructure

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScaffoldFrontmatterParser_ParsesAllEightFields(t *testing.T) {
	t.Parallel()
	src := `---
name: foo
description: x
kind: command
phase: groom
when_to_use: testing
tools: Read, Bash
bash_substitution_policy: none
license: Apache-2.0
---
body content`

	p := newScaffoldFrontmatterParser()
	fm, body, _, hasFM, err := p.Parse(src)
	require.NoError(t, err)
	assert.True(t, hasFM)
	assert.Equal(t, "foo", fm["name"])
	assert.Equal(t, "command", fm["kind"])
	assert.Equal(t, "groom", fm["phase"])
	assert.Equal(t, "Apache-2.0", fm["license"])
	assert.Equal(t, "body content", body)
}

func TestScaffoldFrontmatterParser_MissingFrontmatter_ReturnsEmptyMap(t *testing.T) {
	t.Parallel()
	p := newScaffoldFrontmatterParser()
	fm, body, _, hasFM, err := p.Parse("no frontmatter\nbody")
	require.NoError(t, err)
	assert.False(t, hasFM)
	assert.Empty(t, fm)
	assert.Equal(t, "no frontmatter\nbody", body)
}

func TestScaffoldFrontmatterParser_UnclosedFrontmatter_TreatedAsNoFrontmatter(t *testing.T) {
	t.Parallel()
	p := newScaffoldFrontmatterParser()
	_, _, _, hasFM, err := p.Parse("---\nname: foo\nbody without close")
	require.NoError(t, err)
	assert.False(t, hasFM)
}

func TestScaffoldFrontmatterParser_ToolsRequiredAsScalar(t *testing.T) {
	t.Parallel()
	src := `---
tools: Read
---
body`
	p := newScaffoldFrontmatterParser()
	fm, _, _, _, err := p.Parse(src)
	require.NoError(t, err)
	assert.Equal(t, "Read", fm["tools"])
}

func TestLineCount(t *testing.T) {
	t.Parallel()
	assert.Equal(t, 0, lineCount(""))
	assert.Equal(t, 1, lineCount("a"))
	assert.Equal(t, 1, lineCount("a\n"))
	assert.Equal(t, 2, lineCount("a\nb"))
	assert.Equal(t, 2, lineCount("a\nb\n"))
}
