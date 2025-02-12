package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/takutakahashi/kommon/pkg/agent"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run a command with the specified agent",
	Long: `Run a command using the specified AI agent.
For example:
  # Use Goose agent with a specific session name
  kommon run --agent goose --session-id 123 "Your prompt here"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		text, err := cmd.Flags().GetString("text")
		if err != nil {
			return err
		}

		if text == "" && len(args) > 0 {
			text = args[0]
		}

		if text == "" {
			return fmt.Errorf("text is required either via --text flag or as an argument")
		}

		return runCommand(text)
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().String("text", "", "Input text for the agent")
}

func runCommand(input string) error {
	ctx := context.Background()

	// Create agent options from viper config
	opts := agent.GooseOptions{
		APIKey:      viper.GetString("api_key"),
		SessionID:   viper.GetString("session_id"),
		Instruction: input,
	}

	// Create agent
	agentClient, initErr := agent.NewGooseAgent(opts)
	if initErr != nil {
		return fmt.Errorf("failed to create agent: %w", initErr)
	}

	// Execute command
	output, execErr := agentClient.Execute(ctx, input)
	if execErr != nil {
		return fmt.Errorf("failed to execute command: %w", execErr)
	}

	// Display result
	fmt.Println(output)

	return nil
}
