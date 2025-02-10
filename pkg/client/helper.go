package client

import (
	"context"
	"fmt"
	"os"

	"github.com/takutakahashi/kommon/pkg/agent"
)

// ClientHelper provides helper functions for CLI client operations
type ClientHelper struct {
	agent    agent.Agent
	dataDir  string
	ctx      context.Context
}

// NewClientHelper creates a new instance of ClientHelper
func NewClientHelper(ctx context.Context, opts agent.AgentOptions, dataDir string, newAgentFunc agent.NewAgentFunc) (*ClientHelper, error) {
	// Create agent instance
	a, err := newAgentFunc(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create agent: %w", err)
	}

	// Create data directory if it doesn't exist
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	return &ClientHelper{
		agent:   a,
		dataDir: dataDir,
		ctx:     ctx,
	}, nil
}

// StartSession initializes a new session
func (c *ClientHelper) StartSession() error {
	return c.agent.StartSession(c.ctx)
}

// Execute runs a command through the agent
func (c *ClientHelper) Execute(input string) (string, error) {
	return c.agent.Execute(c.ctx, input)
}

// Cleanup performs cleanup operations
func (c *ClientHelper) Cleanup() error {
	// Add any cleanup operations here if needed
	return nil
}