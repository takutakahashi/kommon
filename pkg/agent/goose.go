package agent

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

type GooseAPIType string

const (
	GooseAPITypeOpenRouter GooseAPIType = "openrouter"
)

// GooseAgent implements the agent interface for Goose
type GooseAgent struct {
	Opts              GooseOptions
	Repo              string
	InstallationToken string
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

	_, writeErr := f.WriteString(text)
	if writeErr != nil {
		return nil, fmt.Errorf("failed to write to file: %w", writeErr)
	}

	// Make the script executable
	if chmodErr := os.Chmod(f.Name(), 0755); chmodErr != nil {
		return nil, fmt.Errorf("failed to chmod file: %w", chmodErr)
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
		removeErr := os.Remove(i.Name())
		if removeErr != nil {
			log.Printf("Failed to remove instruction file %s: %v", i.Name(), removeErr)
		}
	}()

	script := fmt.Sprintf(`#!/bin/bash
gh auth login --with-token <<< "%s"
gh auth setup-git
SESSION_ID=%s
SESSION_DIR=./tmp/$SESSION_ID
REPO=%s
INPUT="%s"
ls $SESSION_DIR || (mkdir -p $SESSION_DIR; git clone $REPO $SESSION_DIR/repo)
cd $SESSION_DIR/repo
goose run --name $SESSION_ID -r --text "$INPUT" || goose run --name $SESSION_ID --text "$INPUT" 
gh auth logout
`, a.InstallationToken, strings.ReplaceAll(a.Opts.SessionID, "/", "-"), fmt.Sprintf("https://github.com/%s", a.Repo), input)

	f, scriptErr := os.CreateTemp("", "goose-script-*.sh")
	if scriptErr != nil {
		log.Printf("Error creating script file: %v", scriptErr)
		return "", fmt.Errorf("failed to create script file: %w", scriptErr)
	}

	log.Printf("Created script file: %s", f.Name())
	log.Printf("Script contents:\n%s", script)

	_, writeErr := f.WriteString(script)
	if writeErr != nil {
		log.Printf("Error writing to script file: %v", writeErr)
		return "", fmt.Errorf("failed to write script: %w", writeErr)
	}

	if chmodErr := os.Chmod(f.Name(), 0755); chmodErr != nil {
		log.Printf("Error setting script permissions: %v", chmodErr)
		return "", fmt.Errorf("failed to set script permissions: %w", chmodErr)
	}

	if closeErr := f.Close(); closeErr != nil {
		log.Printf("Error closing script file: %v", closeErr)
		return "", fmt.Errorf("failed to close script file: %w", closeErr)
	}

	defer func() {
		removeErr := os.Remove(f.Name())
		if removeErr != nil {
			log.Printf("Failed to remove script file %s: %v", f.Name(), removeErr)
		}
	}()

	// #nosec G204 -- This is a controlled environment where we create the script
	cmd := exec.CommandContext(ctx, "bash", f.Name())
	log.Printf("Executing command: %v", cmd.String())

	out, execErr := cmd.CombinedOutput()
	if execErr != nil {
		log.Printf("Command execution error: %v", execErr)
		log.Printf("Command output: %s", string(out))
		return "", fmt.Errorf("command execution failed: %w", execErr)
	}

	return string(out), nil
}

// GetSessionID returns the current session ID
func (a *GooseAgent) GetSessionID() string {
	return a.Opts.SessionID
}
