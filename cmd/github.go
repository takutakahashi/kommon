package cmd

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/go-github/v57/github"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	githubCmd = &cobra.Command{
		Use:   "github",
		Short: "GitHub App webhook server commands",
		Long:  `GitHub App webhook server that handles various GitHub events`,
		RunE:  runServe,
	}
)

func init() {
	cobra.OnInitialize(initConfig)

	githubCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.kommon.yaml)")

	githubCmd.Flags().String("port", "8080", "Port to run the server on")
	githubCmd.Flags().Duration("shutdown-timeout", 30*time.Second, "Shutdown timeout duration")
	githubCmd.Flags().Int64("app-id", 0, "GitHub App ID")
	githubCmd.Flags().String("private-key-file", "", "Path to GitHub App private key file")
	githubCmd.Flags().String("webhook-secret", "", "GitHub webhook secret for request validation")

	if err := viper.BindPFlag("github.port", githubCmd.Flags().Lookup("port")); err != nil {
		cobra.CheckErr(err)
	}
	if err := viper.BindPFlag("github.shutdown_timeout", githubCmd.Flags().Lookup("shutdown-timeout")); err != nil {
		cobra.CheckErr(err)
	}
	if err := viper.BindPFlag("github.app_id", githubCmd.Flags().Lookup("app-id")); err != nil {
		cobra.CheckErr(err)
	}
	if err := viper.BindPFlag("github.private_key_file", githubCmd.Flags().Lookup("private-key-file")); err != nil {
		cobra.CheckErr(err)
	}
	if err := viper.BindPFlag("github.webhook_secret", githubCmd.Flags().Lookup("webhook-secret")); err != nil {
		cobra.CheckErr(err)
	}

	viper.SetDefault("github.port", "8080")
	viper.SetDefault("github.shutdown_timeout", 30*time.Second)
	viper.SetDefault("github.app_id", 0)
	viper.SetDefault("github.private_key_file", "")
	viper.SetDefault("github.webhook_secret", "")
}

func init() {
	rootCmd.AddCommand(githubCmd)
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".kommon")
	}

	viper.AutomaticEnv()
	viper.SetEnvPrefix("KOMMON")

	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

type WebhookServer struct {
	log           *logrus.Logger
	server        *http.Server
	appID         int64
	privateKey    *rsa.PrivateKey
	webhookSecret string
}

type Config struct {
	Port            string
	WebhookSecret   string
	AppID           int64
	PrivateKeyFile  string
	ShutdownTimeout time.Duration
}

// generateJWT generates a JWT for GitHub App authentication
func (ws *WebhookServer) generateJWT() (string, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(10 * time.Minute)),
		Issuer:    strconv.FormatInt(ws.appID, 10),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(ws.privateKey)
}

// getInstallationClient creates a new GitHub client for a specific installation
func (ws *WebhookServer) getInstallationClient(ctx context.Context, installationID int64) (*github.Client, error) {
	jwt, err := ws.generateJWT()
	if err != nil {
		return nil, fmt.Errorf("failed to generate JWT: %v", err)
	}

	// Create a temporary client using the JWT
	jwtClient := github.NewTokenClient(ctx, jwt)

	// Get an installation token
	token, _, err := jwtClient.Apps.CreateInstallationToken(
		ctx,
		installationID,
		&github.InstallationTokenOptions{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create installation token: %v", err)
	}

	// Create a new client using the installation token
	return github.NewTokenClient(ctx, token.GetToken()), nil
}

func NewWebhookServer(cfg Config) (*WebhookServer, error) {
	log := logrus.New()
	log.SetFormatter(&logrus.JSONFormatter{})

	// Read private key
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

	mux := http.NewServeMux()
	ws := &WebhookServer{
		log:           log,
		appID:         cfg.AppID,
		privateKey:    privateKey,
		webhookSecret: cfg.WebhookSecret,
		server: &http.Server{
			Addr:              ":" + cfg.Port,
			Handler:           mux,
			ReadHeaderTimeout: 10 * time.Second,
		},
	}

	mux.HandleFunc("/webhook", ws.handleWebhook)
	return ws, nil
}

func (ws *WebhookServer) Start() error {
	done := make(chan bool)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		ws.log.Info("Server is shutting down...")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := ws.server.Shutdown(ctx); err != nil {
			ws.log.Fatalf("Could not gracefully shutdown the server: %v\n", err)
		}
		close(done)
	}()

	ws.log.Infof("Server is starting on port%s", ws.server.Addr)
	if err := ws.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("could not listen on %s: %v", ws.server.Addr, err)
	}

	<-done
	ws.log.Info("Server stopped")
	return nil
}

func (ws *WebhookServer) handleWebhook(w http.ResponseWriter, r *http.Request) {
	logrus.Info("hello")
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	logrus.Info("hello")

	payload, err := github.ValidatePayload(r, []byte(ws.webhookSecret))
	if err != nil {
		ws.log.Errorf("Error validating webhook payload: %v", err)
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}
	logrus.Info("hello")

	event, err := github.ParseWebHook(github.WebHookType(r), payload)
	if err != nil {
		ws.log.Errorf("Error parsing webhook: %v", err)
		http.Error(w, "Error parsing webhook", http.StatusBadRequest)
		return
	}
	logrus.Info("hello")

	// Get installation ID from the event
	var installationID int64
	switch e := event.(type) {
	case *github.PushEvent:
		installationID = e.GetInstallation().GetID()
		ws.handlePushEvent(r.Context(), e, installationID)
	case *github.PullRequestEvent:
		installationID = e.GetInstallation().GetID()
		ws.handlePullRequestEvent(r.Context(), e, installationID)
	default:
		ws.log.WithFields(logrus.Fields{
			"event_type": github.WebHookType(r),
		}).Info("Received unhandled event type")
	}

	w.WriteHeader(http.StatusOK)
}

func (ws *WebhookServer) handlePushEvent(ctx context.Context, event *github.PushEvent, installationID int64) {
	client, err := ws.getInstallationClient(ctx, installationID)
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

	// Example: Get repository details using the installation client
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

func (ws *WebhookServer) handlePullRequestEvent(ctx context.Context, event *github.PullRequestEvent, installationID int64) {
	client, err := ws.getInstallationClient(ctx, installationID)
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

	// Example: Add a comment to the PR
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

func runServe(cmd *cobra.Command, args []string) error {
	cfg := Config{
		Port:            viper.GetString("github.port"),
		WebhookSecret:   viper.GetString("github.webhook_secret"),
		AppID:           viper.GetInt64("github.app_id"),
		PrivateKeyFile:  viper.GetString("github.private_key_file"),
		ShutdownTimeout: viper.GetDuration("github.shutdown_timeout"),
	}

	server, err := NewWebhookServer(cfg)
	if err != nil {
		return fmt.Errorf("failed to create webhook server: %v", err)
	}

	return server.Start()
}

func GetGitHubCommand() *cobra.Command {
	return githubCmd
}
