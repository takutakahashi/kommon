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
	agentClient, initErr := agent.NewAgent(opts)
	if initErr != nil {
		return fmt.Errorf("failed to create agent: %w", initErr)
	}

	// Initialize client helper
	ctx := context.Background()
	helper, helperErr := client.NewClientHelper(ctx, viper.GetString("data_dir"), agentClient)
	if helperErr != nil {
		return fmt.Errorf("failed to create client helper: %w", helperErr)
	}
	defer func() {
		if closeErr := helper.Close(); closeErr != nil {
			fmt.Printf("Warning: failed to close helper: %v\n", closeErr)
		}
	}()

	// Initialize session
	if sessionErr := agentClient.StartSession(ctx); sessionErr != nil {
		return fmt.Errorf("failed to start session: %w", sessionErr)
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