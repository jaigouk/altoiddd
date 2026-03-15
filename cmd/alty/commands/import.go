package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/alty-cli/alty/internal/composition"
)

// NewImportCmd creates the "alty import" command.
func NewImportCmd(app *composition.App) *cobra.Command {
	var docsDir string

	cmd := &cobra.Command{
		Use:   "import",
		Short: "Import existing DDD documentation into alty's domain model",
		Long:  "Parses docs/DDD.md to extract bounded contexts, classifications, and context relationships.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			result, err := app.DocImportHandler.Import(cmd.Context(), docsDir)
			if err != nil {
				return fmt.Errorf("import: %w", err)
			}

			model := result.Model()
			contexts := model.BoundedContexts()
			relationships := model.ContextRelationships()

			w := cmd.OutOrStdout()
			_, _ = fmt.Fprintf(w, "Imported %d bounded context(s) from %s/DDD.md\n", len(contexts), docsDir)
			for _, bc := range contexts {
				classStr := "unclassified"
				if bc.Classification() != nil {
					classStr = string(*bc.Classification())
				}
				_, _ = fmt.Fprintf(w, "  - %s (%s)\n", bc.Name(), classStr)
			}

			if len(relationships) > 0 {
				_, _ = fmt.Fprintf(w, "\n%d context relationship(s):\n", len(relationships))
				for _, rel := range relationships {
					_, _ = fmt.Fprintf(w, "  %s -> %s [%s]\n", rel.Upstream(), rel.Downstream(), rel.IntegrationPattern())
				}
			}

			for _, warn := range result.Warnings() {
				_, _ = fmt.Fprintf(w, "  warning: %s — %s\n", warn.Section(), warn.Reason())
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&docsDir, "docs-dir", "docs", "Directory containing DDD.md")

	return cmd
}
