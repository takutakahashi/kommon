package executor

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/takutakahashi/kommon/pkg/agent"
)

func setupDockerClient() (*client.Client, error) {
	// Use custom Docker socket if running on macOS with Colima
	if _, err := os.Stat("/Users/owner/.colima/default/docker.sock"); err == nil {
		os.Setenv("DOCKER_HOST", "unix:///Users/owner/.colima/default/docker.sock")
	}

	return client.NewClientWithOpts(client.FromEnv, client.WithVersion("1.46"))
}

func cleanupTestContainers(t *testing.T) {
	cli, err := setupDockerClient()
	require.NoError(t, err)
	defer cli.Close()

	containers, err := cli.ContainerList(context.Background(), container.ListOptions{
		All: true,
	})
	require.NoError(t, err)

	for _, c := range containers {
		if _, ok := c.Labels["kommon.agent.id"]; ok {
			timeout := int(10)
			err = cli.ContainerStop(context.Background(), c.ID, container.StopOptions{
				Timeout: &timeout,
			})
			require.NoError(t, err)

			err = cli.ContainerRemove(context.Background(), c.ID, container.RemoveOptions{
				Force: true,
			})
			require.NoError(t, err)
		}
	}
}

func TestDockerExecutor(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping Docker-based test in short mode")
	}

	// Cleanup any leftover test containers
	cleanupTestContainers(t)

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
	defer func() {
		if err := executor.Close(); err != nil {
			t.Errorf("Failed to close executor: %v", err)
		}
	}()

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

func TestDockerExecutorIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping Docker integration test in short mode")
	}

	// Cleanup any leftover test containers
	cleanupTestContainers(t)

	ctx := context.Background()
	opts := ExecutorOptions{
		Type: ExecutorTypeDocker,
		Resources: &ResourceRequirements{
			Image:       "goose:latest",
			CPULimit:    "1.0",
			MemoryLimit: "512Mi",
		},
	}

	t.Run("ContainerLifecycle", func(t *testing.T) {
		executor, err := NewDockerExecutor(opts)
		require.NoError(t, err)
		require.NotNil(t, executor)
		defer executor.Close()

		err = executor.Initialize(ctx)
		require.NoError(t, err)

		// Create Docker client to verify container state
		cli, err := setupDockerClient()
		require.NoError(t, err)
		defer cli.Close()

		agentOpts := agent.AgentOptions{
			SessionID: "test-lifecycle-agent",
			BaseURL:   "http://localhost:8080",
			APIKey:    "test-key",
		}

		// Create agent and verify container exists
		agent, err := executor.CreateAgent(ctx, agentOpts)
		require.NoError(t, err)
		require.NotNil(t, agent)

		// Verify container is running
		containers, err := cli.ContainerList(ctx, container.ListOptions{})
		require.NoError(t, err)
		var found bool
		for _, c := range containers {
			for _, name := range c.Names {
				if name == "/kommon-agent-"+agentOpts.SessionID {
					found = true
					assert.Equal(t, "running", c.State)
					break
				}
			}
		}
		assert.True(t, found, "Container should be running")

		// Destroy agent and verify container is removed
		err = executor.DestroyAgent(ctx, agentOpts.SessionID)
		require.NoError(t, err)

		// Wait a bit for container to be removed
		time.Sleep(2 * time.Second)

		containers, err = cli.ContainerList(ctx, container.ListOptions{All: true})
		require.NoError(t, err)
		found = false
		for _, c := range containers {
			for _, name := range c.Names {
				if name == "/kommon-agent-"+agentOpts.SessionID {
					found = true
					break
				}
			}
		}
		assert.False(t, found, "Container should be removed")
	})

	t.Run("ContainerErrorHandling", func(t *testing.T) {
		// Test with invalid image
		invalidOpts := ExecutorOptions{
			Type: ExecutorTypeDocker,
			Resources: &ResourceRequirements{
				Image:       "non-existent-image:latest",
				CPULimit:    "1.0",
				MemoryLimit: "512Mi",
			},
		}

		executor, err := NewDockerExecutor(invalidOpts)
		require.NoError(t, err)
		require.NotNil(t, executor)
		defer executor.Close()

		err = executor.Initialize(ctx)
		require.NoError(t, err)

		agentOpts := agent.AgentOptions{
			SessionID: "test-error-agent",
			BaseURL:   "http://localhost:8080",
			APIKey:    "test-key",
		}

		// Attempt to create agent with invalid image
		_, err = executor.CreateAgent(ctx, agentOpts)
		assert.Error(t, err, "Should fail with invalid image")

		// Verify no containers are left running
		cli, err := setupDockerClient()
		require.NoError(t, err)
		defer cli.Close()

		containers, err := cli.ContainerList(ctx, container.ListOptions{All: true})
		require.NoError(t, err)
		for _, c := range containers {
			for _, name := range c.Names {
				assert.NotEqual(t, "/kommon-agent-"+agentOpts.SessionID, name, "No container should exist for failed agent")
			}
		}
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