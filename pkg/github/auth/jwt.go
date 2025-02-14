package auth

import (
	"context"
	"crypto/rsa"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/go-github/v57/github"
)

// GitHubAuth は GitHub 認証に関する機能を提供します
type GitHubAuth struct {
	appID      int64
	privateKey *rsa.PrivateKey
}

// NewGitHubAuth は新しい GitHubAuth インスタンスを作成します
func NewGitHubAuth(appID int64, privateKey *rsa.PrivateKey) *GitHubAuth {
	return &GitHubAuth{
		appID:      appID,
		privateKey: privateKey,
	}
}

// GenerateJWT は GitHub App 認証用の JWT を生成します
func (a *GitHubAuth) GenerateJWT() (string, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(10 * time.Minute)),
		Issuer:    strconv.FormatInt(a.appID, 10),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(a.privateKey)
}

// GetInstallationClient は特定のインストールのための GitHub クライアントを作成します
func (a *GitHubAuth) GetInstallationClient(ctx context.Context, installationID int64) (*github.Client, error) {
	jwt, err := a.GenerateJWT()
	if err != nil {
		return nil, fmt.Errorf("failed to generate JWT: %v", err)
	}

	// JWT を使用して一時的なクライアントを作成
	jwtClient := github.NewTokenClient(ctx, jwt)

	// インストールトークンを取得
	token, _, err := jwtClient.Apps.CreateInstallationToken(
		ctx,
		installationID,
		&github.InstallationTokenOptions{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create installation token: %v", err)
	}

	// インストールトークンを使用して新しいクライアントを作成
	return github.NewTokenClient(ctx, token.GetToken()), nil
}