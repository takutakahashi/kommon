package agent

import (
	"context"
)

// Agent represents the interface for AI agents
type Agent interface {
	// StartSession initializes a new session for the agent
	StartSession(ctx context.Context) error

	// Execute runs a command within the current session
	Execute(ctx context.Context, input string) (string, error)

	// CloseSession closes the current session
	CloseSession(ctx context.Context) error
}

// AgentOptions contains configuration options for creating a new agent
type AgentOptions struct {
	BaseURL   string            // Base URL for API endpoint
	APIKey    string           // API key for authentication
	Headers   map[string]string // Additional headers
	SessionID string           // Session ID
}

// NewAgent creates a new instance of an AI agent with the specified options
type NewAgentFunc func(opts AgentOptions) (Agent, error)