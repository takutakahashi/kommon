package executor

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/takutakahashi/kommon/pkg/agent"
)

// LocalExecutor implements the Executor interface for local execution
type LocalExecutor struct {
	options ExecutorOptions
	agents  map[string]agent.Agent
	mutex   sync.RWMutex
}

// NewLocalExecutor creates a new instance of LocalExecutor
func NewLocalExecutor(opts ExecutorOptions) (*LocalExecutor, error) {
	if opts.ConfigDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		opts.ConfigDir = filepath.Join(homeDir, ".kommon", "local-executor")
	}

	return &LocalExecutor{
		options: opts,
		agents:  make(map[string]agent.Agent),
	}, nil
}

// Initialize implements Executor.Initialize
func (e *LocalExecutor) Initialize(ctx context.Context) error {
	// Create config directory if it doesn't exist
	if err := os.MkdirAll(e.options.ConfigDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	return nil
}

// CreateAgent implements Executor.CreateAgent
func (e *LocalExecutor) CreateAgent(ctx context.Context, opts agent.AgentOptions) (agent.Agent, error) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	// Check if agent already exists
	if _, exists := e.agents[opts.SessionID]; exists {
		return nil, fmt.Errorf("agent with ID %s already exists", opts.SessionID)
	}

	// Create agent instance
	newAgent, err := agent.NewAgent(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create agent: %w", err)
	}

	// Start agent session
	if err := newAgent.StartSession(ctx); err != nil {
		return nil, fmt.Errorf("failed to start agent session: %w", err)
	}

	// Store agent
	e.agents[opts.SessionID] = newAgent
	return newAgent, nil
}

// DestroyAgent implements Executor.DestroyAgent
func (e *LocalExecutor) DestroyAgent(ctx context.Context, agentID string) error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	agent, exists := e.agents[agentID]
	if !exists {
		return fmt.Errorf("agent with ID %s not found", agentID)
	}

	// Cleanup agent resources if needed
	if closer, ok := agent.(interface{ Close() error }); ok {
		if err := closer.Close(); err != nil {
			return fmt.Errorf("failed to close agent: %w", err)
		}
	}

	delete(e.agents, agentID)
	return nil
}

// ListAgents implements Executor.ListAgents
func (e *LocalExecutor) ListAgents(ctx context.Context) ([]string, error) {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	agents := make([]string, 0, len(e.agents))
	for id := range e.agents {
		agents = append(agents, id)
	}
	return agents, nil
}

// GetStatus implements Executor.GetStatus
func (e *LocalExecutor) GetStatus(ctx context.Context) (*ExecutorStatus, error) {
	e.mutex.RLock()
	agentCount := len(e.agents)
	e.mutex.RUnlock()

	status := &ExecutorStatus{
		Type:         ExecutorTypeLocal,
		IsReady:      true,
		ActiveAgents: agentCount,
	}

	// Get resource usage
	resourceStatus, err := e.getResourceStatus()
	if err != nil {
		return nil, fmt.Errorf("failed to get resource status: %w", err)
	}
	status.ResourceStatus = resourceStatus

	return status, nil
}

// getResourceStatus collects system resource usage information
func (e *LocalExecutor) getResourceStatus() (*ResourceStatus, error) {
	// Get CPU usage
	cpuPercent, err := cpu.Percent(0, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get CPU usage: %w", err)
	}

	// Get memory usage
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		return nil, fmt.Errorf("failed to get memory info: %w", err)
	}

	// Get disk usage for the config directory
	diskInfo, err := disk.Usage(e.options.ConfigDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get disk usage: %w", err)
	}

	return &ResourceStatus{
		CPUUsage:    cpuPercent[0],
		MemoryUsage: float64(memInfo.Used),
		DiskUsage:   float64(diskInfo.Used),
	}, nil
}

// Close cleans up resources used by the executor
func (e *LocalExecutor) Close() error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	var errs []error
	for id, agent := range e.agents {
		if closer, ok := agent.(interface{ Close() error }); ok {
			if err := closer.Close(); err != nil {
				errs = append(errs, fmt.Errorf("failed to close agent %s: %w", id, err))
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors occurred while closing executor: %v", errs)
	}
	return nil
}
