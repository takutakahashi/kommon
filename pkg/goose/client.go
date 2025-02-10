package goose

import (
	"context"

	"github.com/takutakahashi/kommon/pkg/interfaces"
)

// Client provides a high-level interface for comment operations
type Client struct {
	provider interfaces.CommentProvider
}

// NewClient creates a new Goose client with the given provider
func NewClient(provider interfaces.CommentProvider) *Client {
	return &Client{
		provider: provider,
	}
}

// GetComments retrieves comments using the configured provider
func (c *Client) GetComments(ctx context.Context) ([]interfaces.Comment, error) {
	return c.provider.GetComments(ctx)
}