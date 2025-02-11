package docker

import (
	"context"
	"fmt"
	"log"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

type Manager struct {
	client     *client.Client
	image      string
	containers map[string]string // key: issueID, value: containerID
}

func NewManager(image string) (*Manager, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %w", err)
	}

	return &Manager{
		client:     cli,
		image:      image,
		containers: make(map[string]string),
	}, nil
}

// StartContainer starts a new goose container for the given issue
func (m *Manager) StartContainer(ctx context.Context, issueID string) error {
	// Check if container already exists
	if _, exists := m.containers[issueID]; exists {
		return fmt.Errorf("container for issue %s already exists", issueID)
	}

	// Create container
	resp, err := m.client.ContainerCreate(ctx,
		&container.Config{
			Image: m.image,
			Cmd:   []string{"goose", "session"},
			Tty:   true,
		},
		nil, nil, nil,
		fmt.Sprintf("goose-%s", issueID),
	)
	if err != nil {
		return fmt.Errorf("failed to create container: %w", err)
	}

	// Start container
	if err := m.client.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return fmt.Errorf("failed to start container: %w", err)
	}

	m.containers[issueID] = resp.ID
	log.Printf("Started container %s for issue %s", resp.ID[:12], issueID)
	return nil
}

// ExecuteCommand executes a command in the issue's container
func (m *Manager) ExecuteCommand(ctx context.Context, issueID string, command string) (string, error) {
	containerID, exists := m.containers[issueID]
	if !exists {
		return "", fmt.Errorf("no container found for issue %s", issueID)
	}

	// Create exec instance
	execConfig := container.ExecOptions{
		Cmd:          []string{"goose", "--name", issueID, command},
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
	}

	execIDResp, err := m.client.ContainerExecCreate(ctx, containerID, execConfig)
	if err != nil {
		return "", fmt.Errorf("failed to create exec instance: %w", err)
	}

	// Start exec instance
	resp, err := m.client.ContainerExecAttach(ctx, execIDResp.ID, container.ExecStartOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to start exec instance: %w", err)
	}
	defer resp.Close()

	// Read output
	output := make([]byte, 0)
	buf := make([]byte, 1024)
	for {
		n, err := resp.Reader.Read(buf)
		if n > 0 {
			output = append(output, buf[:n]...)
		}
		if err != nil {
			break
		}
	}

	// Get exit code
	execInspect, err := m.client.ContainerExecInspect(ctx, execIDResp.ID)
	if err != nil {
		return "", fmt.Errorf("failed to inspect exec instance: %w", err)
	}

	if execInspect.ExitCode != 0 {
		return "", fmt.Errorf("command failed with exit code %d: %s", execInspect.ExitCode, string(output))
	}

	return string(output), nil
}

// StopContainer stops and removes the container for the given issue
func (m *Manager) StopContainer(ctx context.Context, issueID string) error {
	containerID, exists := m.containers[issueID]
	if !exists {
		return fmt.Errorf("no container found for issue %s", issueID)
	}

	// Stop container with 10s timeout
	timeout := int(10)
	if err := m.client.ContainerStop(ctx, containerID, container.StopOptions{Timeout: &timeout}); err != nil {
		return fmt.Errorf("failed to stop container: %w", err)
	}

	// Remove container
	if err := m.client.ContainerRemove(ctx, containerID, container.RemoveOptions{}); err != nil {
		return fmt.Errorf("failed to remove container: %w", err)
	}

	delete(m.containers, issueID)
	log.Printf("Stopped and removed container %s for issue %s", containerID[:12], issueID)
	return nil
}

// Close closes the docker client
func (m *Manager) Close() error {
	return m.client.Close()
}
