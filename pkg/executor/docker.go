package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"

	"github.com/takutakahashi/kommon/pkg/agent"
)

// DockerExecutor implements the Executor interface for Docker-based execution
type DockerExecutor struct {
	options      ExecutorOptions
	dockerClient *client.Client
	containers   map[string]string // agentID -> containerID
	mutex        sync.RWMutex
}

// NewDockerExecutor creates a new instance of DockerExecutor
func NewDockerExecutor(opts ExecutorOptions) (*DockerExecutor, error) {
	// Use custom Docker socket if running on macOS with Colima
	if _, err := os.Stat("/Users/owner/.colima/default/docker.sock"); err == nil {
		os.Setenv("DOCKER_HOST", "unix:///Users/owner/.colima/default/docker.sock")
	}

	// Create Docker client with specific API version
	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithVersion("1.46"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %w", err)
	}

	return &DockerExecutor{
		options:      opts,
		dockerClient: cli,
		containers:   make(map[string]string),
	}, nil
}

// Initialize implements Executor.Initialize
func (e *DockerExecutor) Initialize(ctx context.Context) error {
	// Check Docker daemon connectivity
	if _, err := e.dockerClient.Ping(ctx); err != nil {
		return fmt.Errorf("failed to connect to Docker daemon: %w", err)
	}

	return nil
}

// CreateAgent implements Executor.CreateAgent
func (e *DockerExecutor) CreateAgent(ctx context.Context, opts agent.AgentOptions) (agent.Agent, error) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	// Check if agent already exists
	if _, exists := e.containers[opts.SessionID]; exists {
		return nil, fmt.Errorf("agent with ID %s already exists", opts.SessionID)
	}

	// Create container config
	containerConfig := &container.Config{
		Image: e.options.Resources.Image,
		Env: []string{
			fmt.Sprintf("AGENT_SESSION_ID=%s", opts.SessionID),
			fmt.Sprintf("AGENT_BASE_URL=%s", opts.BaseURL),
			fmt.Sprintf("AGENT_API_KEY=%s", opts.APIKey),
		},
		Labels: map[string]string{
			"kommon.agent.id": opts.SessionID,
		},
	}

	// Create host config with resource limits if specified
	hostConfig := &container.HostConfig{}
	if e.options.Resources != nil {
		hostConfig.Resources = container.Resources{
			Memory:    e.parseMemoryLimit(e.options.Resources.MemoryLimit),
			CPUQuota:  e.parseCPUQuota(e.options.Resources.CPULimit),
			CPUPeriod: 100000, // Default period (100ms)
		}
	}

	// Create container
	resp, createErr := e.dockerClient.ContainerCreate(
		ctx,
		containerConfig,
		hostConfig,
		nil,
		nil,
		fmt.Sprintf("kommon-agent-%s", opts.SessionID),
	)
	if createErr != nil {
		return nil, fmt.Errorf("failed to create container: %w", createErr)
	}

	// Start container
	if startErr := e.dockerClient.ContainerStart(ctx, resp.ID, container.StartOptions{}); startErr != nil {
		return nil, fmt.Errorf("failed to start container: %w", startErr)
	}

	e.containers[opts.SessionID] = resp.ID

	// Create agent instance
	newAgent, agentErr := agent.NewAgent(opts)
	if agentErr != nil {
		// Cleanup container on agent creation failure
		if cleanupErr := e.DestroyAgent(ctx, opts.SessionID); cleanupErr != nil {
			return nil, fmt.Errorf("failed to create agent and cleanup failed: %w", agentErr)
		}
		return nil, fmt.Errorf("failed to create agent: %w", agentErr)
	}

	// Start agent session
	if sessionErr := newAgent.StartSession(ctx); sessionErr != nil {
		// Cleanup container on session start failure
		if cleanupErr := e.DestroyAgent(ctx, opts.SessionID); cleanupErr != nil {
			return nil, fmt.Errorf("failed to start session and cleanup failed: %w", sessionErr)
		}
		return nil, fmt.Errorf("failed to start agent session: %w", sessionErr)
	}

	return newAgent, nil
}

// DestroyAgent implements Executor.DestroyAgent
func (e *DockerExecutor) DestroyAgent(ctx context.Context, agentID string) error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	containerID, exists := e.containers[agentID]
	if !exists {
		return fmt.Errorf("agent with ID %s not found", agentID)
	}

	// Stop container with timeout
	timeout := int(10)
	if stopErr := e.dockerClient.ContainerStop(ctx, containerID, container.StopOptions{
		Timeout: &timeout,
	}); stopErr != nil {
		return fmt.Errorf("failed to stop container: %w", stopErr)
	}

	// Remove container
	if removeErr := e.dockerClient.ContainerRemove(ctx, containerID, container.RemoveOptions{
		Force: true,
	}); removeErr != nil {
		return fmt.Errorf("failed to remove container: %w", removeErr)
	}

	delete(e.containers, agentID)
	return nil
}

// ListAgents implements Executor.ListAgents
func (e *DockerExecutor) ListAgents(ctx context.Context) ([]string, error) {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	containers, listErr := e.dockerClient.ContainerList(ctx, container.ListOptions{
		Filters: e.buildAgentFilter(),
	})
	if listErr != nil {
		return nil, fmt.Errorf("failed to list containers: %w", listErr)
	}

	agents := make([]string, 0, len(containers))
	for _, c := range containers {
		if agentID, ok := c.Labels["kommon.agent.id"]; ok {
			agents = append(agents, agentID)
		}
	}

	return agents, nil
}

// GetStatus implements Executor.GetStatus
func (e *DockerExecutor) GetStatus(ctx context.Context) (*ExecutorStatus, error) {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	// Get Docker info
	info, infoErr := e.dockerClient.Info(ctx)
	if infoErr != nil {
		return nil, fmt.Errorf("failed to get Docker info: %w", infoErr)
	}

	// Get container stats
	containers, listErr := e.dockerClient.ContainerList(ctx, container.ListOptions{
		Filters: e.buildAgentFilter(),
	})
	if listErr != nil {
		return nil, fmt.Errorf("failed to list containers: %w", listErr)
	}

	var totalCPU float64
	var totalMemory float64
	for _, c := range containers {
		stats, statsErr := e.dockerClient.ContainerStats(ctx, c.ID, false)
		if statsErr != nil {
			continue
		}
		defer func() {
			if closeErr := stats.Body.Close(); closeErr != nil {
				log.Printf("Failed to close stats body: %v", closeErr)
			}
		}()

		var s container.StatsResponse
		if decodeErr := json.NewDecoder(stats.Body).Decode(&s); decodeErr != nil {
			continue
		}

		cpuPercent := calculateCPUPercentUnix(s)
		totalCPU += cpuPercent
		totalMemory += float64(s.MemoryStats.Usage)
	}

	status := &ExecutorStatus{
		Type:         ExecutorTypeDocker,
		IsReady:      info.ID != "",
		ActiveAgents: len(containers),
		ResourceStatus: &ResourceStatus{
			CPUUsage:    totalCPU,
			MemoryUsage: totalMemory,
		},
	}

	return status, nil
}

// Close implements io.Closer
func (e *DockerExecutor) Close() error {
	if e.dockerClient != nil {
		return e.dockerClient.Close()
	}
	return nil
}

// Helper functions

func (e *DockerExecutor) buildAgentFilter() filters.Args {
	f := filters.NewArgs()
	f.Add("label", "kommon.agent.id")
	return f
}

func (e *DockerExecutor) parseMemoryLimit(limit string) int64 {
	if limit == "" {
		return 0
	}
	// Parse memory limit (e.g., "1Gi" -> bytes)
	// This is a simplified version, you might want to add more unit support
	var multiplier int64
	switch {
	case strings.HasSuffix(limit, "Gi"):
		multiplier = 1024 * 1024 * 1024
		limit = strings.TrimSuffix(limit, "Gi")
	case strings.HasSuffix(limit, "Mi"):
		multiplier = 1024 * 1024
		limit = strings.TrimSuffix(limit, "Mi")
	case strings.HasSuffix(limit, "Ki"):
		multiplier = 1024
		limit = strings.TrimSuffix(limit, "Ki")
	default:
		multiplier = 1
	}

	value, parseErr := strconv.ParseInt(limit, 10, 64)
	if parseErr != nil {
		return 0
	}
	return value * multiplier
}

func (e *DockerExecutor) parseCPUQuota(limit string) int64 {
	if limit == "" {
		return 0
	}
	// Parse CPU limit (e.g., "1.0" -> CPU quota)
	// Docker uses microseconds for CPU quota
	value, parseErr := strconv.ParseFloat(limit, 64)
	if parseErr != nil {
		return 0
	}
	return int64(value * 100000) // 100000 is the default period
}

func calculateCPUPercentUnix(stats container.StatsResponse) float64 {
	cpuPercent := 0.0
	cpuDelta := float64(stats.CPUStats.CPUUsage.TotalUsage - stats.PreCPUStats.CPUUsage.TotalUsage)
	// Use CPU cycles for system delta as SystemCPUUsage is not directly available
	systemDelta := float64(100000) // Default CPU period

	if systemDelta > 0.0 && cpuDelta > 0.0 {
		cpuPercent = (cpuDelta / systemDelta) * 100.0
	}
	return cpuPercent
}