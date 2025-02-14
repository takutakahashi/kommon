package agent

import (
	"context"
	"fmt"
	"log"
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
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	log.Printf("Created temporary file: %s", f.Name())
	
	_, err = f.WriteString(text)
	if err != nil {
		return nil, fmt.Errorf("failed to write to file: %w", err)
	}
	
	// Make the script executable
	if err := os.Chmod(f.Name(), 0755); err != nil {
		return nil, fmt.Errorf("failed to chmod file: %w", err)
	}
	log.Printf("Set executable permissions on file: %s", f.Name())
	
	return f, nil
}

// Execute sends a command to Goose
func (a *GooseAgent) Execute(ctx context.Context, input string) (string, error) {
	instruction := `gh command can be used. all edit is under new branch checkout from main and PR it.`
	i, err := setFile(instruction)
	if err != nil {
		log.Printf("Error setting instruction file: %v", err)
		return "", err
	}
	defer func() {
		i.Close()
		if err := os.Remove(i.Name()); err != nil {
			log.Printf("Failed to remove instruction file %s: %v", i.Name(), err)
		}
	}()

	script := fmt.Sprintf(`#!/bin/bash
goose run --text '%s'
`, input)
	
	f, err := os.CreateTemp("", "goose-script-*.sh")
	if err != nil {
		log.Printf("Error creating script file: %v", err)
		return "", fmt.Errorf("failed to create script file: %w", err)
	}
	
	log.Printf("Created script file: %s", f.Name())
	log.Printf("Script contents:\n%s", script)
	
	_, err = f.WriteString(script)
	if err != nil {
		log.Printf("Error writing to script file: %v", err)
		return "", fmt.Errorf("failed to write script: %w", err)
	}
	
	if err := f.Chmod(0755); err != nil {
		log.Printf("Error setting script permissions: %v", err)
		return "", fmt.Errorf("failed to set script permissions: %w", err)
	}
	
	if err := f.Close(); err != nil {
		log.Printf("Error closing script file: %v", err)
		return "", fmt.Errorf("failed to close script file: %w", err)
	}
	
	defer func() {
		if err := os.Remove(f.Name()); err != nil {
			log.Printf("Failed to remove script file %s: %v", f.Name(), err)
		}
	}()
	
	cmd := exec.CommandContext(ctx, "bash", f.Name())
	log.Printf("Executing command: %v", cmd.String())
	
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Command execution error: %v", err)
		log.Printf("Command output: %s", string(out))
		return "", fmt.Errorf("command execution failed: %w", err)
	}
	
	return string(out), nil
}

// GetSessionID returns the current session ID
func (a *GooseAgent) GetSessionID() string {
	return a.Opts.SessionID
}
