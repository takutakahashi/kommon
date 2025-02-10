package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/takutakahashi/kommon/pkg/agent"
	"github.com/takutakahashi/kommon/pkg/client"
)

var histCmd = &cobra.Command{
	Use:   "hist",
	Short: "Show command history",
	Long: `Display the history of commands and their outputs.
For example:
  kommon hist`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return showHistory()
	},
}

func init() {
	rootCmd.AddCommand(histCmd)
}

func showHistory() error {
	// Initialize client helper with dummy agent
	opts := agent.AgentOptions{}
	ctx := context.Background()
	helper, err := client.NewClientHelper(ctx, opts, viper.GetString("data_dir"), agent.NewBaseAgent)
	if err != nil {
		return fmt.Errorf("failed to create client helper: %w", err)
	}
	defer helper.Cleanup()

	// Get history
	history, err := helper.GetHistory()
	if err != nil {
		return fmt.Errorf("failed to get history: %w", err)
	}

	if history == "" {
		fmt.Println("No history found")
		return nil
	}

	fmt.Println(history)
	return nil
}