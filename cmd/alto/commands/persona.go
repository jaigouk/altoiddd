package commands

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/alto-cli/alto/internal/composition"
)

// NewPersonaCmd creates the "alto persona" command group.
func NewPersonaCmd(app *composition.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "persona",
		Short: "Manage AI agent persona configurations",
	}

	cmd.AddCommand(
		newPersonaListCmd(app),
		newPersonaGenerateCmd(app),
	)

	return cmd
}

func newPersonaListCmd(app *composition.App) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all available persona definitions",
		RunE: func(cmd *cobra.Command, args []string) error {
			personas := app.PersonaHandler.ListPersonas()
			if len(personas) == 0 {
				fmt.Println("No personas registered.")
				return nil
			}

			fmt.Println("Available Personas:")
			for _, p := range personas {
				fmt.Printf("  - %s: %s\n", p.Name(), p.Description())
			}
			return nil
		},
	}
}

func newPersonaGenerateCmd(app *composition.App) *cobra.Command {
	var tool string
	var yes bool

	cmd := &cobra.Command{
		Use:   "generate <persona-name>",
		Short: "Generate persona configuration files",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			personaName := args[0]

			preview, err := app.PersonaHandler.BuildPreview(personaName, tool)
			if err != nil {
				return fmt.Errorf("persona preview: %w", err)
			}

			fmt.Println(preview.Summary)
			fmt.Println()

			if !yes {
				fmt.Print("Write this persona file? [y/N] ")
				reader := bufio.NewReader(os.Stdin)
				answer, _ := reader.ReadString('\n')
				if strings.TrimSpace(strings.ToLower(answer)) != "y" {
					fmt.Println("Cancelled.")
					return nil
				}
			}

			if err := app.PersonaHandler.ApproveAndWrite(
				context.Background(), preview, ".",
			); err != nil {
				return fmt.Errorf("writing persona: %w", err)
			}

			fmt.Printf("Persona %q written to %s\n", personaName, preview.TargetPath)
			return nil
		},
	}

	cmd.Flags().StringVar(&tool, "tool", "claude-code",
		"Target tool: claude-code, cursor, roo-code, opencode")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false,
		"Skip confirmation prompt")

	return cmd
}
