package server

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/go-github/v57/github"
	"github.com/sirupsen/logrus"
	"github.com/takutakahashi/kommon/pkg/config"
	"github.com/takutakahashi/kommon/pkg/github/agent"
	"github.com/takutakahashi/kommon/pkg/github/auth"
)

// WebhookServer は GitHub webhook サーバーを表します
type WebhookServer struct {
	log           *logrus.Logger
	server        *http.Server
	auth          *auth.GitHubAuth
	webhookSecret string
	appSlug       string
	agentManager  *agent.Manager
}

// NewWebhookServer は新しい WebhookServer インスタンスを作成します
func NewWebhookServer(cfg *config.GithubConfig) (*WebhookServer, error) {
	log := logrus.New()
	log.SetFormatter(&logrus.JSONFormatter{})

	// 秘密鍵の読み込み
	privateKeyBytes, err := os.ReadFile(cfg.PrivateKeyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key file: %v", err)
	}

	block, _ := pem.Decode(privateKeyBytes)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %v", err)
	}

	githubAuth := auth.NewGitHubAuth(cfg.AppID, privateKey)

	ws := &WebhookServer{
		log:           log,
		auth:          githubAuth,
		webhookSecret: cfg.WebhookSecret,
		server: &http.Server{
			Addr:              ":" + cfg.Port,
			Handler:           nil,
			ReadHeaderTimeout: 10 * time.Second,
		},
		agentManager: agent.NewManager(),
	}

	// GitHub App 情報の取得
	jwt, err := githubAuth.GenerateJWT()
	if err != nil {
		return nil, fmt.Errorf("failed to generate JWT: %v", err)
	}

	ctx := context.Background()
	client := github.NewTokenClient(ctx, jwt)
	app, _, err := client.Apps.Get(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get app information: %v", err)
	}
	ws.appSlug = app.GetSlug()
	log.Infof("GitHub App name (@%s) retrieved successfully", ws.appSlug)

	// Create mux and set handlers
	mux := http.NewServeMux()
	mux.HandleFunc("/webhook", ws.handleWebhook)
	ws.server.Handler = mux

	return ws, nil
}

// Start はサーバーを起動します
func (ws *WebhookServer) Start(ctx context.Context) error {
	ws.log.Infof("Server is starting on port%s", ws.server.Addr)
	if err := ws.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("could not listen on %s: %v", ws.server.Addr, err)
	}
	return nil
}

// Shutdown はサーバーを安全に停止します
func (ws *WebhookServer) Shutdown(ctx context.Context) error {
	ws.log.Info("Server is shutting down...")
	return ws.server.Shutdown(ctx)
}

// isKommonCommand はコマンドが kommon コマンドかどうかを判定します
func (ws *WebhookServer) isKommonCommand(text string) bool {
	mentionText := "@" + ws.appSlug
	return strings.Contains(text, "/kommon") || strings.Contains(text, mentionText)
}

// handleWebhook は webhook リクエストを処理します
func (ws *WebhookServer) handleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	payload, err := github.ValidatePayload(r, []byte(ws.webhookSecret))
	if err != nil {
		ws.log.Errorf("Error validating webhook payload: %v", err)
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	event, err := github.ParseWebHook(github.WebHookType(r), payload)
	if err != nil {
		ws.log.Errorf("Error parsing webhook: %v", err)
		http.Error(w, "Error parsing webhook", http.StatusBadRequest)
		return
	}

	var installationID int64
	switch e := event.(type) {
	case *github.PushEvent:
		installationID = e.GetInstallation().GetID()
		ws.handlePushEvent(r.Context(), e, installationID)
	case *github.PullRequestEvent:
		installationID = e.GetInstallation().GetID()
		ws.handlePullRequestEvent(r.Context(), e, installationID)
	case *github.IssueCommentEvent:
		installationID = e.GetInstallation().GetID()
		ws.handleIssueCommentEvent(r.Context(), e, installationID)
	default:
		ws.log.WithFields(logrus.Fields{
			"event_type": github.WebHookType(r),
		}).Info("Received unhandled event type")
	}

	w.WriteHeader(http.StatusOK)
}