package goose

import (
	"fmt"

	"github.com/takutakahashi/kommon/pkg/github"
)

// GitHubOptions contains configuration for GitHub client creation
type GitHubOptions struct {
	Token  string
	Owner  string
	Repo   string
	Number int
	Type   github.ReferenceType
}

// NewGitHubClient creates a new Goose client configured for GitHub
func NewGitHubClient(opts GitHubOptions) (*Client, error) {
	provider, err := github.NewProvider(github.Options{
		Token:  opts.Token,
		Owner:  opts.Owner,
		Repo:   opts.Repo,
		Number: opts.Number,
		Type:   opts.Type,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create GitHub provider: %w", err)
	}

	return NewClient(provider), nil
}