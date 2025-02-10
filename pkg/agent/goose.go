package agent

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// GooseAgent implements the Agent interface for Goose
type GooseAgent struct {
	issueID string
	options AgentOptions
}

// NewGooseAgent creates a new instance of GooseAgent
func NewGooseAgent(opts AgentOptions) (Agent, error) {
	if opts.IssueID == "" {
		return nil, fmt.Errorf("issue ID is required")
	}
	return &GooseAgent{
		options: opts,
		issueID: opts.IssueID,
	}, nil
}

// StartSession implements Agent.StartSession
func (a *GooseAgent) StartSession(ctx context.Context) error {
	// Goose doesn't require explicit session initialization
	return nil
}

// Execute implements Agent.Execute
func (a *GooseAgent) Execute(ctx context.Context, input string) (string, error) {
	if input == "" {
		return "", fmt.Errorf("input is required")
	}

	// Execute goose command
	cmd := exec.CommandContext(ctx, "goose", "run", "--name", a.issueID, "--text", input)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to execute goose command: %v, output: %s", err, string(output))
	}

	return strings.TrimSpace(string(output)), nil
}