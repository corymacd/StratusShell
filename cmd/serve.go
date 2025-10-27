package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/corymacd/cloud-dev-cli-env/internal/server"
	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the web UI server",
	Long:  `Start HTTP server with GoTTY terminal management and HTMX UI.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		port, _ := cmd.Flags().GetInt("port")
		dbPath, _ := cmd.Flags().GetString("db")

		// Default DB path if not specified
		if dbPath == "" {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("failed to get home directory: %w", err)
			}
			dbPath = filepath.Join(homeDir, ".stratusshell", "data.db")
		}

		// Create and run server
		srv, err := server.NewServer(port, dbPath)
		if err != nil {
			return fmt.Errorf("failed to create server: %w", err)
		}

		return srv.Run()
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().StringP("user", "u", "", "Run as specific user (when started by root)")
	serveCmd.Flags().IntP("port", "p", 8080, "HTTP port")
	serveCmd.Flags().String("db", "", "Database path (default: ~/.stratusshell/data.db)")
}
