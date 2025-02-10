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
	rootCmd.PersistentFlags().String("model", "", "AI model to use or issue number for Goose")
	rootCmd.PersistentFlags().String("api-key", "", "API key for the AI service")
	rootCmd.PersistentFlags().String("base-url", "", "Base URL for the AI service")
	rootCmd.PersistentFlags().String("data-dir", getDefaultDataDir(), "Directory for storing session and history data")

	// Bind flags to viper
	viper.BindPFlag("agent", rootCmd.PersistentFlags().Lookup("agent"))
	viper.BindPFlag("model", rootCmd.PersistentFlags().Lookup("model"))
	viper.BindPFlag("api_key", rootCmd.PersistentFlags().Lookup("api-key"))
	viper.BindPFlag("base_url", rootCmd.PersistentFlags().Lookup("base-url"))
	viper.BindPFlag("data_dir", rootCmd.PersistentFlags().Lookup("data-dir"))

	// Environment variables
	viper.BindEnv("api_key", "KOMMON_API_KEY")
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