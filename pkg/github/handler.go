package github

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-github/v57/github"
	"github.com/sirupsen/logrus"
)

// Handler handles GitHub issue events
type Handler struct {
	client   *github.Client
	log      *logrus.Logger
	appLogin string // GitHub Appのログイン名
}

// NewHandler creates a new Handler instance
func NewHandler(client *github.Client, log *logrus.Logger, appLogin string) *Handler {
	return &Handler{
		client:   client,
		log:      log,
		appLogin: appLogin,
	}
}

// HandleIssuesEvent processes GitHub issue events
func (h *Handler) HandleIssuesEvent(ctx context.Context, event *github.IssuesEvent) error {
	issue := event.GetIssue()
	repo := event.GetRepo()

	// ラベルイベントの処理
	if event.GetAction() == "labeled" {
		label := event.GetLabel()
		h.log.WithFields(logrus.Fields{
			"issue_number": issue.GetNumber(),
			"label":        label.GetName(),
			"repo":         repo.GetFullName(),
		}).Info("Issue labeled")

		switch label.GetName() {
		case "needs-review":
			return h.handleNeedsReview(ctx, repo, issue)
		case "bug":
			return h.handleBugLabel(ctx, repo, issue)
		}
	}

	// メンションの処理
	if event.GetAction() == "opened" || event.GetAction() == "edited" {
		return h.handleMention(ctx, repo, issue)
	}

	return nil
}

// handleMention processes mentions of the GitHub App in issues
func (h *Handler) handleMention(ctx context.Context, repo *github.Repository, issue *github.Issue) error {
	body := issue.GetBody()
	if body == "" {
		return nil
	}

	// @アプリ名 のメンションを確認
	mention := fmt.Sprintf("@%s", h.appLogin)
	if !strings.Contains(body, mention) {
		return nil
	}

	h.log.WithFields(logrus.Fields{
		"issue_number": issue.GetNumber(),
		"repo":        repo.GetFullName(),
		"body":        body,
	}).Info("App mentioned in issue")

	// メンションへの応答
	comment := &github.IssueComment{
		Body: github.String(fmt.Sprintf("こんにちは！ @%s です。\nメンションありがとうございます。\n\n"+
			"以下のコマンドが利用可能です:\n"+
			"- `help` - 利用可能なコマンドの一覧を表示\n"+
			"- `status` - 現在の状態を確認\n"+
			"- `assign` - 担当者の割り当て\n\n"+
			"コマンドを使用する場合は、`@%s command` の形式で指定してください。", 
			h.appLogin, h.appLogin)),
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

func (h *Handler) handleNeedsReview(ctx context.Context, repo *github.Repository, issue *github.Issue) error {
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

func (h *Handler) handleBugLabel(ctx context.Context, repo *github.Repository, issue *github.Issue) error {
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