package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/takutakahashi/kommon/pkg/config"
	"github.com/takutakahashi/kommon/pkg/github/server"
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

	rootCmd.AddCommand(githubCmd)
}

func runServe(cmd *cobra.Command, args []string) error {
	// github-specific フラグから設定を取得
	cfg := &config.GithubConfig{
		Port:            viper.GetString("github.port"),
		WebhookSecret:   viper.GetString("github.webhook_secret"),
		AppID:           viper.GetInt64("github.app_id"),
		PrivateKeyFile:  viper.GetString("github.private_key_file"),
		ShutdownTimeout: viper.GetDuration("github.shutdown_timeout"),
	}

	// ルートレベルの環境変数からも設定を取得
	if cfg.WebhookSecret == "" {
		cfg.WebhookSecret = viper.GetString("github_app_webhook_secret")
	}
	if cfg.AppID == 0 {
		cfg.AppID = viper.GetInt64("github_app_id")
	}
	if cfg.PrivateKeyFile == "" {
		cfg.PrivateKeyFile = viper.GetString("github_app_private_key")
	}

	webhookServer, err := server.NewWebhookServer(cfg)
	if err != nil {
		return fmt.Errorf("failed to create webhook server: %v", err)
	}

	// シグナルハンドリングの設定
	done := make(chan bool, 1)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	ctx := context.Background()
	go func() {
		<-quit
		ctx, cancel := context.WithTimeout(ctx, cfg.ShutdownTimeout)
		defer cancel()

		if err := webhookServer.Shutdown(ctx); err != nil {
			fmt.Printf("Could not gracefully shutdown the server: %v\n", err)
		}
		close(done)
	}()

	return webhookServer.Start(ctx)
}

func GetGitHubCommand() *cobra.Command {
	return githubCmd
}