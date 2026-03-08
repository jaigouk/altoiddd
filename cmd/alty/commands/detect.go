package commands

import (
	"fmt"

	"github.com/alty-cli/alty/internal/composition"
	"github.com/spf13/cobra"
)

// NewDetectCmd creates the "alty detect" command.
func NewDetectCmd(app *composition.App) *cobra.Command {
	return &cobra.Command{
		Use:   "detect [project-dir]",
		Short: "Scan for installed AI coding tools and global settings",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectDir := "."
			if len(args) > 0 {
				projectDir = args[0]
			}

			result, err := app.DetectionHandler.Detect(projectDir)
			if err != nil {
				return fmt.Errorf("detection: %w", err)
			}

			tools := result.DetectedTools()
			if len(tools) == 0 {
				fmt.Println("No AI coding tools detected.")
				return nil
			}

			fmt.Println("Detected AI coding tools:")
			for _, tool := range tools {
				fmt.Printf("  - %s\n", tool.Name())
			}

			conflicts := result.Conflicts()
			if len(conflicts) > 0 {
				fmt.Println("\nConfiguration conflicts:")
				for _, c := range conflicts {
					fmt.Printf("  [WARNING] %s\n", c)
				}
			}

			return nil
		},
	}
}
