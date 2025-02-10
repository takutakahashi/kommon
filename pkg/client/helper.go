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

// SaveHistory saves command history to a file
func (c *ClientHelper) SaveHistory(input, output string) error {
	historyFile := fmt.Sprintf("%s/history.txt", c.dataDir)
	entry := fmt.Sprintf("Input: %s\nOutput: %s\n---\n", input, output)
	
	f, err := os.OpenFile(historyFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open history file: %w", err)
	}
	defer f.Close()

	if _, err := f.WriteString(entry); err != nil {
		return fmt.Errorf("failed to write history: %w", err)
	}
	return nil
}

// GetHistory retrieves command history
func (c *ClientHelper) GetHistory() (string, error) {
	historyFile := fmt.Sprintf("%s/history.txt", c.dataDir)
	data, err := os.ReadFile(historyFile)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("failed to read history: %w", err)
	}
	return string(data), nil
}

// Cleanup performs cleanup operations
func (c *ClientHelper) Cleanup() error {
	return c.agent.CloseSession(c.ctx)
}