package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/takutakahashi/kommon/pkg/agent"
	"github.com/takutakahashi/kommon/pkg/client"
)

func main() {
	// コマンドライン引数の解析
	var (
		sessionID = flag.String("session", "", "Session ID to use")
		apiKey    = flag.String("api-key", os.Getenv("KOMMON_API_KEY"), "API key for the AI service")
		baseURL   = flag.String("base-url", "", "Base URL for the AI service")
		dataDir   = flag.String("data-dir", getDefaultDataDir(), "Directory for storing session data")
		text      = flag.String("text", "", "Input text for the agent")
	)
	flag.Parse()

	if *sessionID == "" {
		log.Fatal("Session ID is required")
	}

	// AgentOptionsの設定
	opts := agent.AgentOptions{
		SessionID: *sessionID,
		APIKey:    *apiKey,
		BaseURL:   *baseURL,
	}

	// Create agent client
	agentClient, err := agent.NewAgent(opts)
	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}

	// ClientHelperの初期化
	ctx := context.Background()
	helper, err := client.NewClientHelper(ctx, *dataDir, agentClient)
	if err != nil {
		log.Fatalf("Failed to create client helper: %v", err)
	}
	defer func() {
		if err := helper.Close(); err != nil {
			log.Printf("Failed to close helper: %v", err)
		}
	}()

	// セッションの初期化
	if err := agentClient.StartSession(ctx); err != nil {
		log.Fatalf("Failed to start session: %v", err)
	}

	// 入力の処理
	input := *text
	if input == "" {
		input = strings.Join(flag.Args(), " ")
	}
	if input == "" {
		fmt.Println("Please provide input using --text flag or as command line arguments")
		return
	}

	// コマンドの実行
	output, err := agentClient.Execute(ctx, input)
	if err != nil {
		log.Fatalf("Failed to execute command: %v", err)
	}

	// 結果の表示
	fmt.Println(output)
}

func getDefaultDataDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ".kommon"
	}
	return filepath.Join(homeDir, ".kommon")
}
