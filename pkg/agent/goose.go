package agent

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// GooseAgent implements the Agent interface for Goose
type GooseAgent struct {
	sessionID string
	options   AgentOptions
}

// NewGooseAgent creates a new instance of GooseAgent
func NewGooseAgent(opts AgentOptions) (Agent, error) {
	return &GooseAgent{
		options: opts,
	}, nil
}

// StartSession implements Agent.StartSession
func (a *GooseAgent) StartSession(ctx context.Context) (string, error) {
	// In Goose's case, we'll use the issue number as the session ID
	if a.options.Model == "" {
		return "", fmt.Errorf("issue number (model) is required")
	}
	a.sessionID = a.options.Model
	return a.sessionID, nil
}

// Resume implements Agent.Resume
func (a *GooseAgent) Resume(ctx context.Context, sessionID string) error {
	if sessionID == "" {
		return fmt.Errorf("session ID is required")
	}
	a.sessionID = sessionID
	return nil
}

// Execute implements Agent.Execute
func (a *GooseAgent) Execute(ctx context.Context, input string) (string, error) {
	if a.sessionID == "" {
		return "", fmt.Errorf("no active session")
	}
	if input == "" {
		return "", fmt.Errorf("input is required")
	}

	// Execute goose command
	cmd := exec.CommandContext(ctx, "goose", "run", "--name", a.sessionID, "--text", input)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to execute goose command: %v, output: %s", err, string(output))
	}

	return strings.TrimSpace(string(output)), nil
}