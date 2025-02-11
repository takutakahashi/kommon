package agent

import (
	"context"
)

// Agent represents the interface for AI agents
type Agent interface {
	// StartSession initializes a new session for the agent
	// Returns any error that occurred
	StartSession(ctx context.Context) error

	// Execute performs the specified action within the current session
	// Returns the result of the execution and any error that occurred
	Execute(ctx context.Context, input string) (string, error)

	// GetSessionID returns the current session ID
	GetSessionID() string
}

// AgentOptions contains configuration options for creating a new agent
type AgentOptions struct {
	SessionID string // Session/Issue ID
	APIKey    string // API key for authentication
	BaseURL   string // Base URL for API endpoint
	Headers   map[string]string // Additional headers
}

// NewAgent creates a new instance of an AI agent with the specified options
func NewAgent(opts AgentOptions) (Agent, error) {
	// For now, we'll create a base agent
	return NewBaseAgent(opts)
}