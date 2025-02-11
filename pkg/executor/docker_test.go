package executor

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/takutakahashi/kommon/pkg/agent"
)

func TestDockerExecutor(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping Docker-based test in short mode")
	}

	// Create executor options with test configuration
	opts := ExecutorOptions{
		Type: ExecutorTypeDocker,
		Resources: &ResourceRequirements{
			Image:       "goose:latest", // Make sure this image exists
			CPULimit:    "1.0",
			MemoryLimit: "512Mi",
		},
	}

	// Create executor
	executor, err := NewDockerExecutor(opts)
	require.NoError(t, err)
	require.NotNil(t, executor)
	defer executor.Close()

	// Test initialization
	ctx := context.Background()
	err = executor.Initialize(ctx)
	require.NoError(t, err)

	// Test agent creation
	agentOpts := agent.AgentOptions{
		SessionID: "test-docker-agent-1",
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

		// Cleanup
		err = executor.DestroyAgent(ctx, agentOpts.SessionID)
		require.NoError(t, err)
	})

	t.Run("ListAgents", func(t *testing.T) {
		// Create test agent
		agent, err := executor.CreateAgent(ctx, agentOpts)
		require.NoError(t, err)
		require.NotNil(t, agent)

		// List agents
		agents, err := executor.ListAgents(ctx)
		require.NoError(t, err)
		assert.Len(t, agents, 1)
		assert.Contains(t, agents, agentOpts.SessionID)

		// Cleanup
		err = executor.DestroyAgent(ctx, agentOpts.SessionID)
		require.NoError(t, err)
	})

	t.Run("GetStatus", func(t *testing.T) {
		// Create test agent
		agent, err := executor.CreateAgent(ctx, agentOpts)
		require.NoError(t, err)
		require.NotNil(t, agent)

		// Get status
		status, err := executor.GetStatus(ctx)
		require.NoError(t, err)
		assert.Equal(t, ExecutorTypeDocker, status.Type)
		assert.True(t, status.IsReady)
		assert.Equal(t, 1, status.ActiveAgents)
		assert.NotNil(t, status.ResourceStatus)

		// Cleanup
		err = executor.DestroyAgent(ctx, agentOpts.SessionID)
		require.NoError(t, err)
	})

	t.Run("DestroyAgent", func(t *testing.T) {
		// Create test agent
		agent, err := executor.CreateAgent(ctx, agentOpts)
		require.NoError(t, err)
		require.NotNil(t, agent)

		// Destroy agent
		err = executor.DestroyAgent(ctx, agentOpts.SessionID)
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

func TestDockerExecutorResourceParsing(t *testing.T) {
	executor := &DockerExecutor{}

	t.Run("ParseMemoryLimit", func(t *testing.T) {
		tests := []struct {
			input    string
			expected int64
		}{
			{"1Gi", 1024 * 1024 * 1024},
			{"512Mi", 512 * 1024 * 1024},
			{"1024Ki", 1024 * 1024},
			{"1024", 1024},
			{"invalid", 0},
		}

		for _, tt := range tests {
			t.Run(tt.input, func(t *testing.T) {
				result := executor.parseMemoryLimit(tt.input)
				assert.Equal(t, tt.expected, result)
			})
		}
	})

	t.Run("ParseCPUQuota", func(t *testing.T) {
		tests := []struct {
			input    string
			expected int64
		}{
			{"1.0", 100000},
			{"0.5", 50000},
			{"2.0", 200000},
			{"invalid", 0},
		}

		for _, tt := range tests {
			t.Run(tt.input, func(t *testing.T) {
				result := executor.parseCPUQuota(tt.input)
				assert.Equal(t, tt.expected, result)
			})
		}
	})
}