package github

import (
	"context"
	"fmt"

	"github.com/google/go-github/v57/github"
	"github.com/sirupsen/logrus"
)

// LabelHandler handles GitHub issue label events
type LabelHandler struct {
	client *github.Client
	log    *logrus.Logger
}

// NewLabelHandler creates a new LabelHandler instance
func NewLabelHandler(client *github.Client, log *logrus.Logger) *LabelHandler {
	return &LabelHandler{
		client: client,
		log:    log,
	}
}

// HandleIssuesEvent processes GitHub issue events
func (h *LabelHandler) HandleIssuesEvent(ctx context.Context, event *github.IssuesEvent) error {
	if event.GetAction() != "labeled" {
		return nil
	}

	issue := event.GetIssue()
	label := event.GetLabel()
	repo := event.GetRepo()

	h.log.WithFields(logrus.Fields{
		"issue_number": issue.GetNumber(),
		"label":        label.GetName(),
		"repo":         repo.GetFullName(),
	}).Info("Issue labeled")

	// カスタムロジックを追加
	switch label.GetName() {
	case "needs-review":
		return h.handleNeedsReview(ctx, repo, issue)
	case "bug":
		return h.handleBugLabel(ctx, repo, issue)
	}

	return nil
}

func (h *LabelHandler) handleNeedsReview(ctx context.Context, repo *github.Repository, issue *github.Issue) error {
	comment := &github.IssueComment{
		Body: github.String("このIssueはレビューが必要とマークされました。チームメンバーが確認します。"),
	}
	
	_, _, err := h.client.Issues.CreateComment(
		ctx,
		repo.GetOwner().GetLogin(),
		repo.GetName(),
		issue.GetNumber(),
		comment,
	)
	if err != nil {
		return fmt.Errorf("failed to create comment: %v", err)
	}
	
	return nil
}

func (h *LabelHandler) handleBugLabel(ctx context.Context, repo *github.Repository, issue *github.Issue) error {
	comment := &github.IssueComment{
		Body: github.String("このIssueはバグとしてマークされました。調査と優先度付けを行います。"),
	}
	
	_, _, err := h.client.Issues.CreateComment(
		ctx,
		repo.GetOwner().GetLogin(),
		repo.GetName(),
		issue.GetNumber(),
		comment,
	)
	if err != nil {
		return fmt.Errorf("failed to create comment: %v", err)
	}
	
	return nil
}