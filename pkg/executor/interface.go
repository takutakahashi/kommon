package executor

import (
	"context"
	"fmt"

	"github.com/takutakahashi/kommon/pkg/agent"
)

// ExecutorType represents the type of executor
type ExecutorType string

const (
	ExecutorTypeLocal     ExecutorType = "local"
	ExecutorTypeDocker    ExecutorType = "docker"
	ExecutorTypeKubernetes ExecutorType = "kubernetes"
)

var (
	ErrUnsupportedExecutorType = fmt.Errorf("unsupported executor type")
)

// Executor represents the interface for agent executors
type Executor interface {
	// Initialize sets up the executor environment
	Initialize(ctx context.Context) error

	// CreateAgent creates and starts a new agent instance
	// Returns the agent interface and any error that occurred
	CreateAgent(ctx context.Context, opts agent.AgentOptions) (agent.Agent, error)

	// DestroyAgent stops and cleans up the specified agent
	DestroyAgent(ctx context.Context, agentID string) error

	// ListAgents returns a list of currently running agents
	ListAgents(ctx context.Context) ([]string, error)

	// GetStatus returns the current status of the executor
	GetStatus(ctx context.Context) (*ExecutorStatus, error)
}

// ExecutorStatus represents the current status of an executor
type ExecutorStatus struct {
	Type           ExecutorType      `json:"type"`
	IsReady        bool             `json:"is_ready"`
	ActiveAgents   int              `json:"active_agents"`
	ResourceStatus *ResourceStatus   `json:"resource_status,omitempty"`
}

// ResourceStatus represents resource usage information
type ResourceStatus struct {
	CPUUsage    float64 `json:"cpu_usage,omitempty"`    // Percentage of CPU usage
	MemoryUsage float64 `json:"memory_usage,omitempty"` // Memory usage in bytes
	DiskUsage   float64 `json:"disk_usage,omitempty"`   // Disk usage in bytes
}

// ExecutorOptions contains configuration options for creating a new executor
type ExecutorOptions struct {
	Type      ExecutorType         `json:"type"`
	ConfigDir string              `json:"config_dir"`
	Resources *ResourceRequirements `json:"resources,omitempty"`
}

// ResourceRequirements specifies resource limits and requests
type ResourceRequirements struct {
	Image       string `json:"image,omitempty"`        // Docker image (for Docker executor)
	CPULimit    string `json:"cpu_limit,omitempty"`    // CPU limit (e.g., "1.0")
	MemoryLimit string `json:"memory_limit,omitempty"` // Memory limit (e.g., "1Gi")
	DiskLimit   string `json:"disk_limit,omitempty"`   // Disk limit (e.g., "10Gi")
}

// NewExecutor creates a new instance of the specified executor type
func NewExecutor(opts ExecutorOptions) (Executor, error) {
	switch opts.Type {
	case ExecutorTypeLocal:
		return NewLocalExecutor(opts)
	case ExecutorTypeDocker:
		return NewDockerExecutor(opts)
	case ExecutorTypeKubernetes:
		return NewKubernetesExecutor(opts)
	default:
		return nil, ErrUnsupportedExecutorType
	}
}