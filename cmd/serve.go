package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/takutakahashi/kommon/pkg/server"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the kommon server",
	Long: `Start the kommon server that manages Goose containers for issues.
The server provides a REST API to execute commands and manage containers.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		addr := fmt.Sprintf("%s:%d", viper.GetString("host"), viper.GetInt("port"))
		image := viper.GetString("goose_image")
		if image == "" {
			image = "ghcr.io/takutakahashi/kommon-goose-agent:latest"
		}

		server, err := server.NewServer(image)
		if err != nil {
			return fmt.Errorf("failed to create server: %w", err)
		}
		defer server.Close()

		return server.Start(addr)
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	// Server flags
	serveCmd.Flags().String("host", "0.0.0.0", "Host to bind the server to")
	serveCmd.Flags().Int("port", 8080, "Port to bind the server to")
	serveCmd.Flags().String("goose-image", "", "Goose container image to use (default: ghcr.io/takutakahashi/kommon-goose-agent:latest)")

	// Bind flags to viper
	viper.BindPFlag("host", serveCmd.Flags().Lookup("host"))
	viper.BindPFlag("port", serveCmd.Flags().Lookup("port"))
	viper.BindPFlag("goose_image", serveCmd.Flags().Lookup("goose-image"))
}
