package interfaces

import "context"

// Comment represents a single comment with basic information
type Comment struct {
	Author    string
	Body      string
	CreatedAt string
	URL       string
}

// CommentProvider defines the interface for fetching comments from various sources
type CommentProvider interface {
	// GetComments retrieves comments from the provider
	GetComments(ctx context.Context) ([]Comment, error)
	// Configure configures the provider with specific options
	Configure(opts interface{}) error
}