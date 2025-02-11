package client

import (
	"context"
	"os"
	"path/filepath"

	"github.com/takutakahashi/kommon/pkg/agent"
)

// ClientHelper manages agent sessions and data persistence
type ClientHelper struct {
	dataDir string
	agent   agent.Agent
}

// NewClientHelper creates a new ClientHelper instance
func NewClientHelper(ctx context.Context, dataDir string, agent agent.Agent) (*ClientHelper, error) {
	// Create data directory if it doesn't exist
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}

	return &ClientHelper{
		dataDir: dataDir,
		agent:   agent,
	}, nil
}

// GetSessionDir returns the directory for the current session
func (h *ClientHelper) GetSessionDir() string {
	// Assuming agent has a method to get session ID
	if sessioner, ok := h.agent.(interface{ GetSessionID() string }); ok {
		return filepath.Join(h.dataDir, "sessions", sessioner.GetSessionID())
	}
	return filepath.Join(h.dataDir, "sessions", "default")
}

// Close cleans up resources
func (h *ClientHelper) Close() error {
	if closer, ok := h.agent.(interface{ Close() error }); ok {
		return closer.Close()
	}
	return nil
}
