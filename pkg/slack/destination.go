package slack

import (
	"context"

	"github.com/takutakahashi/kommon/pkg/interfaces"
)

// Destination implements the CommentDestination interface for Slack
type Destination struct {
	webhookURL string
}

// NewDestination creates a new Slack comment destination
func NewDestination(webhookURL string) *Destination {
	return &Destination{
		webhookURL: webhookURL,
	}
}

// SendComment sends a single comment to Slack
func (d *Destination) SendComment(ctx context.Context, comment interfaces.Comment) error {
	// TODO: Implement Slack message sending
	return nil
}

// SendComments sends multiple comments to Slack
func (d *Destination) SendComments(ctx context.Context, comments []interfaces.Comment) error {
	for _, comment := range comments {
		if err := d.SendComment(ctx, comment); err != nil {
			return err
		}
	}
	return nil
}