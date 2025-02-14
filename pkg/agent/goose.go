package agent

import (
	"context"
	"fmt"
	"os"
	"os/exec"
)

type GooseAPIType string

const (
	GooseAPITypeOpenRouter GooseAPIType = "openrouter"
)

// GooseAgent implements the agent interface for Goose
type GooseAgent struct {
	Opts GooseOptions
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
		Opts: opts,
	}, nil
}

func setFile(text string) (*os.File, error) {
	f, err := os.CreateTemp("", "goose-script-*.sh")
	if err != nil {
		return nil, err
	}
	defer os.Remove(f.Name())
	_, err = f.WriteString(text)
	if err != nil {
		return nil, err
	}
	return f, nil
}

// Execute sends a command to Goose
func (a *GooseAgent) Execute(ctx context.Context, input string) (string, error) {
	instruction := `gh command can be used. all edit is under new branch checkout from main and PR it.`
	i, err := setFile(instruction)
	if err != nil {
		return "", err
	}
	defer i.Close()
	defer os.Remove(i.Name())
	script := fmt.Sprintf(`
	#!/bin/bash
	goose run --name %s --text %s
	`, a.Opts.SessionID, input)
	f, err := os.CreateTemp("", "goose-script-*.sh")
	if err != nil {
		return "", err
	}
	defer os.Remove(f.Name())
	_, err = f.WriteString(script)
	if err != nil {
		return "", err
	}
	err = f.Close()
	if err != nil {
		return "", err
	}
	out, err := exec.CommandContext(ctx, "bash", "-c", f.Name()).Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// GetSessionID returns the current session ID
func (a *GooseAgent) GetSessionID() string {
	return a.Opts.SessionID
}
