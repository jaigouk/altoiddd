package commands

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/alty-cli/alty/internal/composition"
)

// NewKBCmd creates the "alty kb" command.
func NewKBCmd(app *composition.App) *cobra.Command {
	return &cobra.Command{
		Use:   "kb [topic]",
		Short: "Look up a topic in the RLM knowledge base",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				// List categories
				fmt.Println("Knowledge Base Categories")
				fmt.Println("----------------------------------------")
				for _, cat := range app.KnowledgeLookupHandler.ListCategories() {
					topics, err := app.KnowledgeLookupHandler.ListTopics(
						context.Background(), cat, nil,
					)
					if err != nil {
						topics = nil
					}
					if len(topics) > 0 {
						fmt.Printf("  %s: %s\n", cat, joinStrings(topics))
					} else {
						fmt.Printf("  %s: (empty)\n", cat)
					}
				}
				return nil
			}

			topic := args[0]
			entry, err := app.KnowledgeLookupHandler.Lookup(
				context.Background(), topic, "",
			)
			if err != nil {
				return fmt.Errorf("lookup %q: %w", topic, err)
			}

			fmt.Println(entry.Content())
			return nil
		},
	}
}

func joinStrings(items []string) string {
	result := ""
	for i, s := range items {
		if i > 0 {
			result += ", "
		}
		result += s
	}
	return result
}
