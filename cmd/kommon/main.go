package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/takutakahashi/kommon/pkg/agent"
	"github.com/takutakahashi/kommon/pkg/client"
)

func main() {
	// コマンドライン引数の解析
	var (
		model   = flag.String("model", "gpt-4", "AI model to use")
		apiKey  = flag.String("api-key", os.Getenv("KOMMON_API_KEY"), "API key for the AI service")
		baseURL = flag.String("base-url", "", "Base URL for the AI service")
		dataDir = flag.String("data-dir", getDefaultDataDir(), "Directory for storing session and history data")
	)
	flag.Parse()

	// 必須パラメータのチェック
	if *apiKey == "" {
		log.Fatal("API key is required. Set it via -api-key flag or KOMMON_API_KEY environment variable")
	}

	// AgentOptionsの設定
	opts := agent.AgentOptions{
		Model:   *model,
		APIKey:  *apiKey,
		BaseURL: *baseURL,
	}

	// ClientHelperの初期化
	ctx := context.Background()
	helper, err := client.NewClientHelper(ctx, opts, *dataDir)
	if err != nil {
		log.Fatalf("Failed to create client helper: %v", err)
	}
	defer helper.Cleanup()

	// セッションの初期化
	if err := helper.InitializeSession(); err != nil {
		log.Fatalf("Failed to initialize session: %v", err)
	}

	// 入力の処理
	input := flag.Arg(0)
	if input == "" {
		if history, err := helper.GetHistory(); err == nil && history != "" {
			fmt.Println("Previous interactions:")
			fmt.Println(history)
		}
		fmt.Println("Please provide input as a command line argument")
		return
	}

	// コマンドの実行
	output, err := helper.Execute(input)
	if err != nil {
		log.Fatalf("Failed to execute command: %v", err)
	}

	// 結果の表示と履歴の保存
	fmt.Println(output)
	if err := helper.SaveHistory(input, output); err != nil {
		log.Printf("Warning: Failed to save history: %v", err)
	}
}

func getDefaultDataDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ".kommon"
	}
	return filepath.Join(homeDir, ".kommon")
}