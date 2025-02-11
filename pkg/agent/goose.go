package agent

import (
	"context"
	"fmt"
)

// GooseAgent implements the agent interface for Goose
type GooseAgent struct {
	opts AgentOptions
}

// NewGooseAgent creates a new Goose agent
func NewGooseAgent(opts AgentOptions) (Agent, error) {
	if opts.SessionID == "" {
		return nil, fmt.Errorf("session ID is required for Goose agent")
	}

	return &GooseAgent{
		opts: opts,
	}, nil
}

// StartSession initializes a new session with Goose
func (a *GooseAgent) StartSession(ctx context.Context) error {
	// TODO: Implement actual Goose session initialization
	return nil
}

// Execute sends a command to Goose
func (a *GooseAgent) Execute(ctx context.Context, input string) (string, error) {
	// TODO: Implement actual Goose command execution
	return fmt.Sprintf("[Goose %s] %s", a.opts.SessionID, input), nil
}

// GetSessionID returns the current session ID
func (a *GooseAgent) GetSessionID() string {
	return a.opts.SessionID
}
