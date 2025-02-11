package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/takutakahashi/kommon/pkg/agent"
	"github.com/takutakahashi/kommon/pkg/client"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run a command with the specified agent",
	Long: `Run a command using the specified AI agent.
For example:
  # Use Goose agent with a specific session name
  kommon run --agent goose --session-id 123 "Your prompt here"
  
  # Use OpenAI agent
  kommon run --agent openai "Your prompt here"`,
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
	// Create agent options from viper config
	opts := agent.AgentOptions{
		APIKey:    viper.GetString("api_key"),
		BaseURL:   viper.GetString("base_url"),
		SessionID: viper.GetString("session_id"),
	}

	// Create agent
	agentClient, err := agent.NewAgent(opts)
	if err != nil {
		return fmt.Errorf("failed to create agent: %w", err)
	}

	// Initialize client helper
	ctx := context.Background()
	helper, err := client.NewClientHelper(ctx, viper.GetString("data_dir"), agentClient)
	if err != nil {
		return fmt.Errorf("failed to create client helper: %w", err)
	}
	defer func() {
		if err := helper.Close(); err != nil {
			fmt.Printf("Warning: failed to close helper: %v\n", err)
		}
	}()

	// Initialize session
	if err := agentClient.StartSession(ctx); err != nil {
		return fmt.Errorf("failed to start session: %w", err)
	}

	// Execute command
	output, err := agentClient.Execute(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to execute command: %w", err)
	}

	// Display result
	fmt.Println(output)

	return nil
}
