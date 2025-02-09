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

// GetComments retrieves comments from a GitHub PR
// Required options:
// - owner: repository owner
// - repo: repository name
// - pr_number: pull request number
func (p *Provider) GetComments(ctx context.Context, opts map[string]interface{}) ([]interfaces.Comment, error) {
	owner, ok := opts["owner"].(string)
	if !ok {
		return nil, fmt.Errorf("owner is required")
	}

	repo, ok := opts["repo"].(string)
	if !ok {
		return nil, fmt.Errorf("repo is required")
	}

	prNum, ok := opts["pr_number"].(int)
	if !ok {
		return nil, fmt.Errorf("pr_number is required")
	}

	comments, _, err := p.client.Issues.ListComments(ctx, owner, repo, prNum, nil)
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