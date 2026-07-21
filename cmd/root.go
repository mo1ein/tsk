package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/graph/task-manager/internal/config"
)

var cfg *config.Config

var rootCmd = &cobra.Command{
	Use:   "task-manager",
	Short: "A microservices task manager API",
	Long:  `A backend service that manages tasks (to-do items) via a REST API.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		cfg, err = config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
