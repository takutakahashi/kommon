package main

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/takutakahashi/kommon/pkg/github"
	"github.com/takutakahashi/kommon/pkg/goose"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Check command line arguments
	if len(os.Args) != 5 {
		return fmt.Errorf("Usage: %s <owner> <repo> <number> <type>\ntype: issue or pr", os.Args[0])
	}

	// Get GitHub token from environment
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return fmt.Errorf("GITHUB_TOKEN environment variable is required")
	}

	// Parse arguments
	owner := os.Args[1]
	repo := os.Args[2]
	number, err := strconv.Atoi(os.Args[3])
	if err != nil {
		return fmt.Errorf("invalid number: %s", os.Args[3])
	}

	// Parse reference type
	var refType github.ReferenceType
	switch os.Args[4] {
	case "issue":
		refType = github.ReferenceTypeIssue
	case "pr":
		refType = github.ReferenceTypePR
	default:
		return fmt.Errorf("invalid type: %s (must be 'issue' or 'pr')", os.Args[4])
	}

	// Create Goose client for GitHub
	client, err := goose.NewGitHubClient(goose.GitHubOptions{
		Token:  token,
		Owner:  owner,
		Repo:   repo,
		Number: number,
		Type:   refType,
	})
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	// Get comments
	ctx := context.Background()
	comments, err := client.GetComments(ctx)
	if err != nil {
		return fmt.Errorf("failed to get comments: %w", err)
	}

	// Print comments
	for _, comment := range comments {
		fmt.Printf("Author: %s\n", comment.Author)
		fmt.Printf("Created At: %s\n", comment.CreatedAt)
		fmt.Printf("URL: %s\n", comment.URL)
		fmt.Printf("Comment: %s\n", comment.Body)
		fmt.Println("---")
	}

	return nil
}