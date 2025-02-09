package github

import (
	"context"
	"fmt"
	"time"

	"github.com/google/go-github/v57/github"
	"github.com/takutakahashi/kommon/pkg/interfaces"
	"golang.org/x/oauth2"
)

// ReferenceType indicates whether the number refers to a PR or an Issue
type ReferenceType int

const (
	ReferenceTypeIssue ReferenceType = iota
	ReferenceTypePR
)

// Options contains GitHub specific options for comment retrieval
type Options struct {
	Token    string
	Owner    string
	Repo     string
	Number   int
	Type     ReferenceType
	// GitHub specific options can be added here
	// APIEndpoint string
	// RetryCount  int
	// etc...
}

// Provider implements the CommentProvider interface for GitHub
type Provider struct {
	client *github.Client
	opts   Options
}

// NewProvider creates a new GitHub comment provider with the given options
func NewProvider(opts Options) (*Provider, error) {
	if opts.Token == "" {
		return nil, fmt.Errorf("token is required")
	}
	if opts.Owner == "" {
		return nil, fmt.Errorf("owner is required")
	}
	if opts.Repo == "" {
		return nil, fmt.Errorf("repo is required")
	}
	if opts.Number <= 0 {
		return nil, fmt.Errorf("number must be positive")
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: opts.Token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	return &Provider{
		client: client,
		opts:   opts,
	}, nil
}

// GetComments retrieves comments from a GitHub Issue or PR
func (p *Provider) GetComments(ctx context.Context) ([]interfaces.Comment, error) {
	// Both Issues and PRs use the same API endpoint in GitHub
	comments, _, err := p.client.Issues.ListComments(ctx, p.opts.Owner, p.opts.Repo, p.opts.Number, nil)
	if err != nil {
		reference := "issue"
		if p.opts.Type == ReferenceTypePR {
			reference = "PR"
		}
		return nil, fmt.Errorf("failed to get %s comments: %w", reference, err)
	}

	var result []interfaces.Comment
	for _, comment := range comments {
		if comment.User == nil || comment.User.Login == nil || comment.Body == nil {
			continue
		}

		c := interfaces.Comment{
			Author: *comment.User.Login,
			Body:   *comment.Body,
			URL:    comment.GetHTMLURL(),
		}

		if comment.CreatedAt != nil {
			c.CreatedAt = comment.CreatedAt.Format(time.RFC3339)
		}

		result = append(result, c)
	}

	return result, nil
}