package config

import "time"

// GithubConfig は GitHub App の設定を保持する構造体
type GithubConfig struct {
	Port            string
	WebhookSecret   string
	AppID           int64
	PrivateKeyFile  string
	ShutdownTimeout time.Duration
}

// NewGithubConfig は新しい GithubConfig インスタンスを作成します
func NewGithubConfig(
	port string,
	webhookSecret string,
	appID int64,
	privateKeyFile string,
	shutdownTimeout time.Duration,
) *GithubConfig {
	return &GithubConfig{
		Port:            port,
		WebhookSecret:   webhookSecret,
		AppID:           appID,
		PrivateKeyFile:  privateKeyFile,
		ShutdownTimeout: shutdownTimeout,
	}
}