package cmd

import (
	"fmt"

	"github.com/karajanchang/murmur-ai/internal/server"
	"github.com/spf13/cobra"
)

var servePort int

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the web dashboard server",
	Long: `Start a local web server to visualize stats and patterns.

The dashboard provides:
  • Usage statistics and trends
  • Pattern management
  • Tool health status
  • Configuration overview

Example:
  mur serve              # Start on default port 8383
  mur serve --port 9000  # Start on custom port`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("🗣️ Murmur Dashboard")
		fmt.Println()

		srv := server.New(server.Config{
			Port: servePort,
		})

		return srv.Start()
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().IntVarP(&servePort, "port", "p", 8383, "port to listen on")
}
