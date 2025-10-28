package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/corymacd/StratusShell/internal/service"
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install systemd service",
	Long:  `Generate and enable systemd service for StratusShell.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check if running as root
		if os.Geteuid() != 0 {
			return fmt.Errorf("install command must be run as root (use sudo)")
		}

		username, _ := cmd.Flags().GetString("user")
		port, _ := cmd.Flags().GetInt("port")

		log.Printf("Installing systemd service for user: %s", username)

		if err := service.InstallSystemdService(username, port); err != nil {
			return fmt.Errorf("failed to install service: %w", err)
		}

		log.Println("✓ Service installed successfully")
		log.Printf("Service: stratusshell-%s.service", username)
		log.Printf("URL: http://localhost:%d", port)
		log.Println()
		log.Println("Useful commands:")
		log.Printf("  sudo systemctl status stratusshell-%s", username)
		log.Printf("  sudo systemctl restart stratusshell-%s", username)
		log.Printf("  sudo systemctl stop stratusshell-%s", username)
		log.Printf("  sudo journalctl -u stratusshell-%s -f", username)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
	installCmd.Flags().StringP("user", "u", "", "User that service runs as (required)")
	installCmd.MarkFlagRequired("user")
	installCmd.Flags().IntP("port", "p", 8080, "HTTP port")
}
