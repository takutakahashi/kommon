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
		agentType = flag.String("agent", "openai", "Agent type (openai or goose)")
		model     = flag.String("model", "", "AI model to use or issue number for Goose")
		apiKey    = flag.String("api-key", os.Getenv("KOMMON_API_KEY"), "API key for the AI service")
		baseURL   = flag.String("base-url", "", "Base URL for the AI service")
		dataDir   = flag.String("data-dir", getDefaultDataDir(), "Directory for storing session data")
		text      = flag.String("text", "", "Input text for the agent")
	)
	flag.Parse()

	// AgentOptionsの設定
	opts := agent.AgentOptions{
		Model:   *model,
		APIKey:  *apiKey,
		BaseURL: *baseURL,
	}

	// Create appropriate agent based on type
	var newAgentFunc agent.NewAgentFunc
	switch strings.ToLower(*agentType) {
	case "goose":
		newAgentFunc = agent.NewGooseAgent
		if *model == "" {
			log.Fatal("Issue number (model) is required for Goose agent")
		}
	case "openai":
		newAgentFunc = agent.NewBaseAgent
		if *apiKey == "" {
			log.Fatal("API key is required for OpenAI agent")
		}
	default:
		log.Fatalf("Unknown agent type: %s", *agentType)
	}

	// ClientHelperの初期化
	ctx := context.Background()
	helper, err := client.NewClientHelper(ctx, opts, *dataDir, newAgentFunc)
	if err != nil {
		log.Fatalf("Failed to create client helper: %v", err)
	}
	defer helper.Cleanup()

	// セッションの初期化
	if err := helper.InitializeSession(); err != nil {
		log.Fatalf("Failed to initialize session: %v", err)
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
	output, err := helper.Execute(input)
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