package agent

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type GooseAPIType string

const (
	GooseAPITypeOpenRouter GooseAPIType = "openrouter"
)

// GooseAgent implements the agent interface for Goose
type GooseAgent struct {
	opts GooseOptions
	dir  string
}

type GooseOptions struct {
	SessionID   string
	APIType     GooseAPIType
	APIKey      string
	Instruction string
}

// NewGooseAgent creates a new Goose agent
func NewGooseAgent(opts GooseOptions) (Agent, error) {
	if opts.SessionID == "" {
		return nil, fmt.Errorf("session ID is required for Goose agent")
	}

	if opts.APIType == "" {
		opts.APIType = GooseAPITypeOpenRouter
	}

	if opts.APIKey == "" {
		return nil, fmt.Errorf("API key is required for Goose agent")
	}

	return &GooseAgent{
		opts: opts,
	}, nil
}

// StartSession initializes a new session with Goose
func (a *GooseAgent) StartSession(ctx context.Context) error {
	dir, err := os.MkdirTemp(a.dir, "goose-session-*")
	if err != nil {
		return err
	}
	os.Create(filepath.Join(dir, "goose.yaml"))
	return nil
}

// Execute sends a command to Goose
func (a *GooseAgent) Execute(ctx context.Context, input string) (string, error) {
	err := exec.CommandContext(ctx, "goose", "run", "--name", a.opts.SessionID, "--text", input).Run()
	if err != nil {
		return "", err
	}
	return "", nil
}

// GetSessionID returns the current session ID
func (a *GooseAgent) GetSessionID() string {
	return a.opts.SessionID
}
