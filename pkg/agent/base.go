package agent

import (
	"context"
	"fmt"
)

// BaseAgent implements a basic agent for testing and development
type BaseAgent struct {
	opts AgentOptions
}

// NewBaseAgent creates a new base agent
func NewBaseAgent(opts AgentOptions) (Agent, error) {
	return &BaseAgent{opts: opts}, nil
}

// StartSession implements Agent.StartSession
func (a *BaseAgent) StartSession(ctx context.Context) error {
	return nil
}

// Execute implements Agent.Execute
func (a *BaseAgent) Execute(ctx context.Context, input string) (string, error) {
	return fmt.Sprintf("Executing in session %s: %s", a.opts.SessionID, input), nil
}

// GetSessionID implements Agent.GetSessionID
func (a *BaseAgent) GetSessionID() string {
	return a.opts.SessionID
}