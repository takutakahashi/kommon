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
}

// AgentOptions contains configuration options for creating a new agent
type AgentOptions struct {
	// Add any necessary configuration options here
	Model    string            // AI model to use (for OpenAI)
	BaseURL  string           // Base URL for API endpoint
	APIKey   string           // API key for authentication
	Headers  map[string]string // Additional headers
	IssueID  string           // Issue ID for Goose agent
}

// NewAgent creates a new instance of an AI agent with the specified options
type NewAgentFunc func(opts AgentOptions) (Agent, error)