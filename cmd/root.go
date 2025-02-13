package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	rootCmd = &cobra.Command{
		Use:   "kommon",
		Short: "Kommon is a developer and adviser CLI tool",
		Long: `Kommon is a CLI tool that helps developers by providing
AI-powered assistance using various agents like OpenAI GPT and Goose.`,
	}
)

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.kommon.yaml)")

	// Common flags
	rootCmd.PersistentFlags().String("agent", "openai", "Agent type (openai or goose)")
	rootCmd.PersistentFlags().String("session-id", "", "Session/Issue ID")
	rootCmd.PersistentFlags().String("api-key", "", "API key for the AI service")
	rootCmd.PersistentFlags().String("base-url", "", "Base URL for the AI service")
	rootCmd.PersistentFlags().String("data-dir", getDefaultDataDir(), "Directory for storing data")

	// GitHub App related flags
	rootCmd.PersistentFlags().String("github-app-id", "", "GitHub App ID")
	rootCmd.PersistentFlags().String("github-app-private-key", "", "Path to GitHub App private key file")
	rootCmd.PersistentFlags().String("github-app-webhook-secret", "", "GitHub App webhook secret")

	// Bind flags to viper with error checking
	if err := viper.BindPFlag("agent", rootCmd.PersistentFlags().Lookup("agent")); err != nil {
		fmt.Printf("Failed to bind agent flag: %v\n", err)
		os.Exit(1)
	}
	if err := viper.BindPFlag("session_id", rootCmd.PersistentFlags().Lookup("session-id")); err != nil {
		fmt.Printf("Failed to bind session_id flag: %v\n", err)
		os.Exit(1)
	}
	if err := viper.BindPFlag("api_key", rootCmd.PersistentFlags().Lookup("api-key")); err != nil {
		fmt.Printf("Failed to bind api_key flag: %v\n", err)
		os.Exit(1)
	}
	if err := viper.BindPFlag("base_url", rootCmd.PersistentFlags().Lookup("base-url")); err != nil {
		fmt.Printf("Failed to bind base_url flag: %v\n", err)
		os.Exit(1)
	}
	if err := viper.BindPFlag("data_dir", rootCmd.PersistentFlags().Lookup("data-dir")); err != nil {
		fmt.Printf("Failed to bind data_dir flag: %v\n", err)
		os.Exit(1)
	}

	// Bind GitHub App related flags
	if err := viper.BindPFlag("github_app_id", rootCmd.PersistentFlags().Lookup("github-app-id")); err != nil {
		fmt.Printf("Failed to bind github_app_id flag: %v\n", err)
		os.Exit(1)
	}
	if err := viper.BindPFlag("github_app_private_key", rootCmd.PersistentFlags().Lookup("github-app-private-key")); err != nil {
		fmt.Printf("Failed to bind github_app_private_key flag: %v\n", err)
		os.Exit(1)
	}
	if err := viper.BindPFlag("github_app_webhook_secret", rootCmd.PersistentFlags().Lookup("github-app-webhook-secret")); err != nil {
		fmt.Printf("Failed to bind github_app_webhook_secret flag: %v\n", err)
		os.Exit(1)
	}

	// Environment variables with error handling
	if err := viper.BindEnv("api_key", "KOMMON_API_KEY"); err != nil {
		fmt.Printf("Warning: failed to bind KOMMON_API_KEY environment variable: %v\n", err)
	}

	// GitHub App related environment variables
	if err := viper.BindEnv("github_app_id", "KOMMON_GITHUB_APP_ID"); err != nil {
		fmt.Printf("Warning: failed to bind KOMMON_GITHUB_APP_ID environment variable: %v\n", err)
	}
	if err := viper.BindEnv("github_app_private_key", "KOMMON_GITHUB_APP_PRIVATE_KEY"); err != nil {
		fmt.Printf("Warning: failed to bind KOMMON_GITHUB_APP_PRIVATE_KEY environment variable: %v\n", err)
	}
	if err := viper.BindEnv("github_app_webhook_secret", "KOMMON_GITHUB_APP_WEBHOOK_SECRET"); err != nil {
		fmt.Printf("Warning: failed to bind KOMMON_GITHUB_APP_WEBHOOK_SECRET environment variable: %v\n", err)
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".kommon" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".kommon")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func getDefaultDataDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ".kommon"
	}
	return filepath.Join(homeDir, ".kommon")
}
