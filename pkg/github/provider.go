package github

import (
	"context"
	"fmt"
	"time"

	"github.com/google/go-github/v57/github"
	"github.com/takutakahashi/kommon/pkg/interfaces"
	"golang.org/x/oauth2"
)

// Provider implements the CommentProvider interface for GitHub
type Provider struct {
	client *github.Client
	opts   Options
}

// NewProvider creates a new GitHub comment provider
func NewProvider(token string) *Provider {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	return &Provider{
		client: client,
	}
}

// Configure sets up the provider with GitHub specific options
func (p *Provider) Configure(opts interface{}) error {
	ghOpts, ok := opts.(Options)
	if !ok {
		return fmt.Errorf("invalid options type: expected github.Options")
	}

	if ghOpts.Owner == "" {
		return fmt.Errorf("owner is required")
	}

	if ghOpts.Repo == "" {
		return fmt.Errorf("repo is required")
	}

	p.opts = ghOpts
	return nil
}

// GetComments retrieves comments from a GitHub PR
func (p *Provider) GetComments(ctx context.Context) ([]interfaces.Comment, error) {
	comments, _, err := p.client.Issues.ListComments(ctx, p.opts.Owner, p.opts.Repo, p.opts.PRNumber, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get PR comments: %w", err)
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