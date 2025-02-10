package client

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/takutakahashi/kommon/pkg/agent"
)

// ClientHelper provides helper functions for CLI client operations
type ClientHelper struct {
	agent    agent.Agent
	session  string
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

// InitializeSession starts a new session or resumes an existing one
func (c *ClientHelper) InitializeSession() error {
	// Try to load existing session
	sessionFile := fmt.Sprintf("%s/session.txt", c.dataDir)
	if data, err := os.ReadFile(sessionFile); err == nil && len(data) > 0 {
		sessionID := strings.TrimSpace(string(data))
		if err := c.agent.Resume(c.ctx, sessionID); err == nil {
			c.session = sessionID
			return nil
		}
	}

	// Start new session if loading failed
	sessionID, err := c.agent.StartSession(c.ctx)
	if err != nil {
		return fmt.Errorf("failed to start session: %w", err)
	}

	// Save session ID
	if err := os.WriteFile(sessionFile, []byte(sessionID), 0644); err != nil {
		return fmt.Errorf("failed to save session ID: %w", err)
	}

	c.session = sessionID
	return nil
}

// Execute runs a command through the agent
func (c *ClientHelper) Execute(input string) (string, error) {
	if c.session == "" {
		return "", fmt.Errorf("no active session")
	}
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
	// Add any cleanup operations here
	return nil
}