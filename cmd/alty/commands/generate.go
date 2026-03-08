package commands

import (
	"fmt"

	"github.com/alty-cli/alty/internal/composition"
	"github.com/spf13/cobra"
)

// NewGenerateCmd creates the "alty generate" command group.
func NewGenerateCmd(app *composition.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate project artifacts",
		Long:  "Generate various project artifacts (DDD docs, fitness tests, tickets, configs).",
	}

	cmd.AddCommand(
		newGenerateArtifactsCmd(app),
		newGenerateFitnessCmd(app),
		newGenerateTicketsCmd(app),
		newGenerateConfigsCmd(app),
	)

	return cmd
}

func newGenerateArtifactsCmd(app *composition.App) *cobra.Command {
	return &cobra.Command{
		Use:   "artifacts",
		Short: "Generate DDD artifacts (PRD, DDD.md, ARCHITECTURE.md)",
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = app.ArtifactGenerationHandler
			fmt.Println("Artifact generation requires a completed discovery session.")
			fmt.Println("Run 'alty guide' first.")
			return nil
		},
	}
}

func newGenerateFitnessCmd(app *composition.App) *cobra.Command {
	return &cobra.Command{
		Use:   "fitness",
		Short: "Generate architecture fitness tests",
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = app.FitnessGenerationHandler
			fmt.Println("Fitness generation requires a domain model from discovery.")
			fmt.Println("Run 'alty guide' first.")
			return nil
		},
	}
}

func newGenerateTicketsCmd(app *composition.App) *cobra.Command {
	return &cobra.Command{
		Use:   "tickets",
		Short: "Generate dependency-ordered beads tickets from DDD",
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = app.TicketGenerationHandler
			fmt.Println("Ticket generation requires a domain model from discovery.")
			fmt.Println("Run 'alty guide' first.")
			return nil
		},
	}
}

func newGenerateConfigsCmd(app *composition.App) *cobra.Command {
	return &cobra.Command{
		Use:   "configs",
		Short: "Generate tool-native configurations for AI coding tools",
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = app.ConfigGenerationHandler
			fmt.Println("Config generation requires a domain model from discovery.")
			fmt.Println("Run 'alty guide' first.")
			return nil
		},
	}
}
