package agent

import (
	"context"
	"fmt"
)

// BaseAgent provides a basic implementation of the Agent interface
type BaseAgent struct {
	options AgentOptions
}

// NewBaseAgent creates a new instance of BaseAgent
func NewBaseAgent(opts AgentOptions) (Agent, error) {
	if opts.APIKey == "" {
		return nil, fmt.Errorf("API key is required")
	}
	return &BaseAgent{
		options: opts,
	}, nil
}

// StartSession implements Agent.StartSession
func (a *BaseAgent) StartSession(ctx context.Context) error {
	// Add session initialization logic here if needed
	return nil
}

// Execute implements Agent.Execute
func (a *BaseAgent) Execute(ctx context.Context, input string) (string, error) {
	if input == "" {
		return "", fmt.Errorf("input is required")
	}
	// Here you would implement the actual execution logic
	// This might involve calling an AI API, processing the input, etc.
	return fmt.Sprintf("Processed input: %s", input), nil
}