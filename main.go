package main

import (
	"context"
	"fmt"
	"os"

	"github.com/google/go-github/v57/github"
	"golang.org/x/oauth2"
)

func main() {
	// GitHub トークンを環境変数から取得
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		fmt.Println("Please set GITHUB_TOKEN environment variable")
		os.Exit(1)
	}

	// GitHub クライアントの初期化
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	// コマンドライン引数からリポジトリ情報と PR 番号を取得
	if len(os.Args) != 4 {
		fmt.Println("Usage: go run main.go <owner> <repo> <pr_number>")
		os.Exit(1)
	}

	owner := os.Args[1]
	repo := os.Args[2]
	prNumber := os.Args[3]

	// PR 番号を整数に変換
	var prNum int
	_, err := fmt.Sscanf(prNumber, "%d", &prNum)
	if err != nil {
		fmt.Printf("Invalid PR number: %s\n", prNumber)
		os.Exit(1)
	}

	// PR のコメントを取得
	comments, _, err := client.Issues.ListComments(ctx, owner, repo, prNum, nil)
	if err != nil {
		fmt.Printf("Error getting PR comments: %v\n", err)
		os.Exit(1)
	}

	// コメントを出力
	for _, comment := range comments {
		fmt.Printf("Author: %s\n", *comment.User.Login)
		fmt.Printf("Comment: %s\n", *comment.Body)
		fmt.Println("---")
	}
}