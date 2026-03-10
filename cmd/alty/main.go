package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/alty-cli/alty/cmd/alty/commands"
	"github.com/alty-cli/alty/internal/composition"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	app, err := composition.NewApp()
	if err != nil {
		return fmt.Errorf("initializing app: %w", err)
	}
	defer func() { _ = app.Close() }()

	rootCmd := newRootCmd(app)
	if err := rootCmd.Execute(); err != nil {
		return fmt.Errorf("executing command: %w", err)
	}
	return nil
}

func newRootCmd(app *composition.App) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "alty",
		Short: "Guided project bootstrapper with DDD + TDD + SOLID",
		Long:  "alty turns a simple idea into a structured, production-ready project.",
	}

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println(app.Version)
		},
	}

	rootCmd.AddCommand(
		versionCmd,
		commands.NewInitCmd(app),
		commands.NewGuideCmd(app),
		commands.NewDetectCmd(app),
		commands.NewCheckCmd(app),
		commands.NewKBCmd(app),
		commands.NewDocHealthCmd(app),
		commands.NewDocReviewCmd(app),
		commands.NewTicketHealthCmd(app),
		commands.NewTicketVerifyCmd(app),
		commands.NewGenerateCmd(app),
		commands.NewPersonaCmd(app),
	)

	return rootCmd
}
