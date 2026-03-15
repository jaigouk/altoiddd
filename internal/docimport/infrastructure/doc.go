package infrastructure

import "github.com/alty-cli/alty/internal/docimport/application"

// Compile-time check that MarkdownDocParser satisfies the DocImporter port.
var _ application.DocImporter = (*MarkdownDocParser)(nil)
