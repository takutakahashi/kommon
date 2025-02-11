package executor

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/takutakahashi/kommon/pkg/agent"
)

func TestLocalExecutor(t *testing.T) {
	// Create temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "local-executor-test-*")
	require.NoError(t, err)
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Errorf("Failed to remove temp dir: %v", err)
		}
	}()

	// Create executor options
	opts := ExecutorOptions{
		Type:      ExecutorTypeLocal,
		ConfigDir: tmpDir,
	}

	// Create executor
	executor, err := NewLocalExecutor(opts)
	require.NoError(t, err)
	require.NotNil(t, executor)

	// Test initialization
	ctx := context.Background()
	err = executor.Initialize(ctx)
	require.NoError(t, err)

	// Verify config directory was created
	_, err = os.Stat(tmpDir)
	require.NoError(t, err)

	// Test agent creation
	agentOpts := agent.AgentOptions{
		SessionID: "test-agent-1",
		BaseURL:   "http://localhost:8080",
		APIKey:    "test-key",
	}

	t.Run("CreateAgent", func(t *testing.T) {
		agent, err := executor.CreateAgent(ctx, agentOpts)
		require.NoError(t, err)
		require.NotNil(t, agent)

		// Try creating duplicate agent
		_, err = executor.CreateAgent(ctx, agentOpts)
		assert.Error(t, err)
	})

	t.Run("ListAgents", func(t *testing.T) {
		agents, err := executor.ListAgents(ctx)
		require.NoError(t, err)
		assert.Len(t, agents, 1)
		assert.Contains(t, agents, "test-agent-1")
	})

	t.Run("GetStatus", func(t *testing.T) {
		status, err := executor.GetStatus(ctx)
		require.NoError(t, err)
		assert.Equal(t, ExecutorTypeLocal, status.Type)
		assert.True(t, status.IsReady)
		assert.Equal(t, 1, status.ActiveAgents)
		assert.NotNil(t, status.ResourceStatus)
	})

	t.Run("DestroyAgent", func(t *testing.T) {
		err := executor.DestroyAgent(ctx, "test-agent-1")
		require.NoError(t, err)

		// Verify agent was removed
		agents, err := executor.ListAgents(ctx)
		require.NoError(t, err)
		assert.Len(t, agents, 0)

		// Try destroying non-existent agent
		err = executor.DestroyAgent(ctx, "non-existent")
		assert.Error(t, err)
	})
}

func TestNewExecutor(t *testing.T) {
	tests := []struct {
		name    string
		opts    ExecutorOptions
		wantErr bool
	}{
		{
			name: "Local executor",
			opts: ExecutorOptions{
				Type: ExecutorTypeLocal,
			},
			wantErr: false,
		},
		{
			name: "Docker executor",
			opts: ExecutorOptions{
				Type: ExecutorTypeDocker,
			},
			wantErr: false, // Docker executor is now implemented
		},
		{
			name: "Kubernetes executor",
			opts: ExecutorOptions{
				Type: ExecutorTypeKubernetes,
			},
			wantErr: true, // Not implemented yet
		},
		{
			name: "Unknown executor type",
			opts: ExecutorOptions{
				Type: "unknown",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor, err := NewExecutor(tt.opts)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, executor)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, executor)
			}
		})
	}
}
