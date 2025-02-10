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

	// Select agent type
	var newAgentFunc agent.NewAgentFunc
	switch viper.GetString("agent") {
	case "goose":
		newAgentFunc = agent.NewGooseAgent
		if opts.SessionID == "" {
			return fmt.Errorf("session ID is required for Goose agent (used as --name parameter)")
		}
	case "openai":
		newAgentFunc = agent.NewBaseAgent
		if opts.APIKey == "" {
			return fmt.Errorf("API key is required for OpenAI agent")
		}
	default:
		return fmt.Errorf("unknown agent type: %s", viper.GetString("agent"))
	}

	// Initialize client helper
	ctx := context.Background()
	helper, err := client.NewClientHelper(ctx, opts, viper.GetString("data_dir"), newAgentFunc)
	if err != nil {
		return fmt.Errorf("failed to create client helper: %w", err)
	}
	defer helper.Cleanup()

	// Initialize session
	if err := helper.StartSession(); err != nil {
		return fmt.Errorf("failed to initialize session: %w", err)
	}

	// Execute command
	output, err := helper.Execute(input)
	if err != nil {
		return fmt.Errorf("failed to execute command: %w", err)
	}

	// Display result
	fmt.Println(output)

	return nil
}