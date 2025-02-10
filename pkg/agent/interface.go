package agent

import (
	"context"
)

// Agent represents the interface for AI agents
type Agent interface {
	// StartSession initializes a new session for the agent
	// Returns a unique session ID and any error that occurred
	StartSession(ctx context.Context) (string, error)

	// Resume restores a previous session using the provided session ID
	// Returns error if the session cannot be resumed
	Resume(ctx context.Context, sessionID string) error

	// Execute performs the specified action within the current session
	// Returns the result of the execution and any error that occurred
	Execute(ctx context.Context, input string) (string, error)
}

// AgentOptions contains configuration options for creating a new agent
type AgentOptions struct {
	// Add any necessary configuration options here
	Model     string            // AI model to use
	BaseURL   string           // Base URL for API endpoint
	APIKey    string           // API key for authentication
	Headers   map[string]string // Additional headers
	MaxTokens int              // Maximum tokens for response
}

// NewAgent creates a new instance of an AI agent with the specified options
type NewAgentFunc func(opts AgentOptions) (Agent, error)