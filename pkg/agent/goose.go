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
	active    bool
}

// NewGooseAgent creates a new instance of GooseAgent
func NewGooseAgent(opts AgentOptions) (Agent, error) {
	return &GooseAgent{
		options:   opts,
		sessionID: opts.SessionID,
		active:    false,
	}, nil
}

// StartSession implements Agent.StartSession
func (a *GooseAgent) StartSession(ctx context.Context) error {
	var cmd *exec.Cmd
	if a.sessionID != "" {
		// Start specified session
		cmd = exec.CommandContext(ctx, "goose", "session", a.sessionID)
	} else {
		// Start new session
		cmd = exec.CommandContext(ctx, "goose", "session")
		// Capture new session ID if needed
		// TODO: Parse output to get session ID
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to start goose session: %v, output: %s", err, string(output))
	}

	a.active = true
	return nil
}

// Execute implements Agent.Execute
func (a *GooseAgent) Execute(ctx context.Context, input string) (string, error) {
	if !a.active {
		return "", fmt.Errorf("no active session")
	}
	if input == "" {
		return "", fmt.Errorf("input is required")
	}

	// Execute command in current session
	cmd := exec.CommandContext(ctx, "goose", "execute", "--text", input)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to execute goose command: %v, output: %s", err, string(output))
	}

	return strings.TrimSpace(string(output)), nil
}

// CloseSession implements Agent.CloseSession
func (a *GooseAgent) CloseSession(ctx context.Context) error {
	if !a.active {
		return nil // Already closed or never started
	}

	if a.sessionID != "" {
		cmd := exec.CommandContext(ctx, "goose", "session", "--close", a.sessionID)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to close goose session: %v, output: %s", err, string(output))
		}
	}

	a.active = false
	return nil
}