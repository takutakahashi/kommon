package main

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/takutakahashi/kommon/pkg/github"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Check command line arguments
	if len(os.Args) != 4 {
		return fmt.Errorf("Usage: %s <owner> <repo> <pr_number>", os.Args[0])
	}

	// Get GitHub token from environment
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return fmt.Errorf("GITHUB_TOKEN environment variable is required")
	}

	// Parse arguments
	owner := os.Args[1]
	repo := os.Args[2]
	prNum, err := strconv.Atoi(os.Args[3])
	if err != nil {
		return fmt.Errorf("invalid PR number: %s", os.Args[3])
	}

	// Create GitHub provider
	provider := github.NewProvider(token)

	// Get comments
	ctx := context.Background()
	comments, err := provider.GetComments(ctx, map[string]interface{}{
		"owner":     owner,
		"repo":      repo,
		"pr_number": prNum,
	})
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