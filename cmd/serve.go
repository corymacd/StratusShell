package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the web UI server",
	Long:  `Start HTTP server with GoTTY terminal management and HTMX UI.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("not yet implemented")
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().StringP("user", "u", "", "Run as specific user (when started by root)")
	serveCmd.Flags().IntP("port", "p", 8080, "HTTP port")
	serveCmd.Flags().String("db", "", "Database path (default: ~/.stratusshell/data.db)")
}
