package server

import (
	"context"

	"github.com/google/go-github/v57/github"
	"github.com/sirupsen/logrus"
)

// handlePushEvent は push イベントを処理します
func (ws *WebhookServer) handlePushEvent(ctx context.Context, event *github.PushEvent, installationID int64) {
	client, err := ws.auth.GetInstallationClient(ctx, installationID)
	if err != nil {
		ws.log.Errorf("Failed to get installation client: %v", err)
		return
	}

	ws.log.WithFields(logrus.Fields{
		"repo":    event.GetRepo().GetFullName(),
		"ref":     event.GetRef(),
		"commits": len(event.Commits),
		"pusher":  event.GetPusher().GetName(),
	}).Info("Received push event")

	// リポジトリ詳細の取得
	repo, _, err := client.Repositories.Get(ctx, event.GetRepo().GetOwner().GetLogin(), event.GetRepo().GetName())
	if err != nil {
		ws.log.Errorf("Failed to get repository details: %v", err)
		return
	}

	ws.log.WithFields(logrus.Fields{
		"stars":          repo.GetStargazersCount(),
		"visibility":     repo.GetVisibility(),
		"default_branch": repo.GetDefaultBranch(),
	}).Info("Repository details")
}

// handlePullRequestEvent は pull request イベントを処理します
func (ws *WebhookServer) handlePullRequestEvent(ctx context.Context, event *github.PullRequestEvent, installationID int64) {
	client, err := ws.auth.GetInstallationClient(ctx, installationID)
	if err != nil {
		ws.log.Errorf("Failed to get installation client: %v", err)
		return
	}

	ws.log.WithFields(logrus.Fields{
		"repo":      event.GetRepo().GetFullName(),
		"pr_number": event.GetPullRequest().GetNumber(),
		"action":    event.GetAction(),
		"title":     event.GetPullRequest().GetTitle(),
	}).Info("Received pull request event")

	if event.GetAction() == "opened" {
		comment := &github.IssueComment{
			Body: github.String("Thank you for your contribution! 🎉"),
		}
		_, _, err := client.Issues.CreateComment(
			ctx,
			event.GetRepo().GetOwner().GetLogin(),
			event.GetRepo().GetName(),
			event.GetPullRequest().GetNumber(),
			comment,
		)
		if err != nil {
			ws.log.Errorf("Failed to create comment: %v", err)
		}
	}
}

// handleIssueCommentEvent は issue comment イベントを処理します
func (ws *WebhookServer) handleIssueCommentEvent(ctx context.Context, event *github.IssueCommentEvent, installationID int64) {
	client, err := ws.auth.GetInstallationClient(ctx, installationID)
	if err != nil {
		ws.log.Errorf("Failed to get installation client: %v", err)
		return
	}

	comment := event.GetComment()
	if comment == nil {
		return
	}

	if !ws.isKommonCommand(comment.GetBody()) {
		return
	}

	ws.log.WithFields(logrus.Fields{
		"repo":       event.GetRepo().GetFullName(),
		"issue":      event.GetIssue().GetNumber(),
		"comment_id": comment.GetID(),
		"comment_by": event.GetSender().GetLogin(),
		"mentioned":  "@" + ws.appSlug,
		"comment":    comment.GetBody(),
	}).Info("Received mention in issue comment")

	agent := ws.agentManager.GetAgent(event.GetRepo().GetFullName(), event.GetIssue().GetNumber(), "")
	res, err := agent.Execute(ctx, comment.GetBody())
	if err != nil {
		ws.log.Errorf("Failed to execute prompt: %v", err)
		return
	}

	newComment := &github.IssueComment{
		Body: github.String(res),
	}

	_, _, err = client.Issues.CreateComment(
		ctx,
		event.GetRepo().GetOwner().GetLogin(),
		event.GetRepo().GetName(),
		event.GetIssue().GetNumber(),
		newComment,
	)
	if err != nil {
		ws.log.Errorf("Failed to post comment: %v", err)
		return
	}

	ws.log.Info("Successfully responded to mention")
}