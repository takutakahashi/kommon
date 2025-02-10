package agent

import (
	"context"
	"fmt"
)

// BaseAgent provides a basic implementation of the Agent interface
type BaseAgent struct {
	options AgentOptions
	active  bool
}

// NewBaseAgent creates a new instance of BaseAgent
func NewBaseAgent(opts AgentOptions) (Agent, error) {
	if opts.APIKey == "" {
		return nil, fmt.Errorf("API key is required")
	}
	return &BaseAgent{
		options: opts,
		active:  false,
	}, nil
}

// StartSession implements Agent.StartSession
func (a *BaseAgent) StartSession(ctx context.Context) error {
	a.active = true
	return nil
}

// Execute implements Agent.Execute
func (a *BaseAgent) Execute(ctx context.Context, input string) (string, error) {
	if !a.active {
		return "", fmt.Errorf("no active session")
	}
	if input == "" {
		return "", fmt.Errorf("input is required")
	}
	// Here you would implement the actual execution logic
	// This might involve calling an AI API, processing the input, etc.
	return fmt.Sprintf("Processed input: %s", input), nil
}

// CloseSession implements Agent.CloseSession
func (a *BaseAgent) CloseSession(ctx context.Context) error {
	a.active = false
	return nil
}