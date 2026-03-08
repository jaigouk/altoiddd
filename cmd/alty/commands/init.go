package commands

import (
	"fmt"

	"github.com/alty-cli/alty/internal/composition"
	"github.com/spf13/cobra"
)

// NewInitCmd creates the "alty init" command.
func NewInitCmd(app *composition.App) *cobra.Command {
	var existing bool

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Bootstrap a new project from a README idea",
		Long: `Bootstrap a new project from a README idea.

Use --existing to rescue an existing project (alty init --existing).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if existing {
				return runRescue(app)
			}
			return runInit(app)
		},
	}

	cmd.Flags().BoolVar(&existing, "existing", false, "Rescue an existing project")

	return cmd
}

func runInit(app *composition.App) error {
	// Preview bootstrap actions
	session, err := app.BootstrapHandler.Preview(".")
	if err != nil {
		return fmt.Errorf("bootstrap preview: %w", err)
	}
	fmt.Printf("Bootstrap session created: %s\n", session.SessionID())
	fmt.Println("Run 'alty guide' to start the discovery flow.")
	return nil
}

func runRescue(app *composition.App) error {
	fmt.Println("Rescue mode: analyzing existing project...")
	fmt.Println("(Rescue flow requires interactive prompts -- not yet wired)")
	return nil
}
