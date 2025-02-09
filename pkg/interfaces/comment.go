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
	// GetComments retrieves comments based on the given options
	GetComments(ctx context.Context, opts GetCommentsOptions) ([]Comment, error)
}

// GetCommentsOptions contains options for retrieving comments
type GetCommentsOptions struct {
	Owner     string
	Repo      string
	PRNumber  int
	// 将来の拡張のために他のフィールドを追加可能
	// 例: StartDate time.Time
	// 例: MaxComments int
}

// CommentDestination defines the interface for sending comments to various destinations
type CommentDestination interface {
	// SendComment sends a comment to the destination
	SendComment(ctx context.Context, comment Comment) error
	// SendComments sends multiple comments to the destination
	SendComments(ctx context.Context, comments []Comment) error
}