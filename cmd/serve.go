package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the kommon server",
	Long: `Start the kommon server that manages Goose containers for issues.
The server provides a REST API to execute commands and manage containers.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	// Server flags
	serveCmd.Flags().String("host", "0.0.0.0", "Host to bind the server to")
	serveCmd.Flags().Int("port", 8080, "Port to bind the server to")
	serveCmd.Flags().String("goose-image", "", "Goose container image to use (default: ghcr.io/takutakahashi/kommon-goose-agent:latest)")

	// Bind flags to viper with error handling
	if err := viper.BindPFlag("host", serveCmd.Flags().Lookup("host")); err != nil {
		fmt.Printf("Warning: failed to bind host flag: %v\n", err)
	}
	if err := viper.BindPFlag("port", serveCmd.Flags().Lookup("port")); err != nil {
		fmt.Printf("Warning: failed to bind port flag: %v\n", err)
	}
	if err := viper.BindPFlag("goose_image", serveCmd.Flags().Lookup("goose-image")); err != nil {
		fmt.Printf("Warning: failed to bind goose_image flag: %v\n", err)
	}
}
