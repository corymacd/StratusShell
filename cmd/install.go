package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install systemd service",
	Long:  `Generate and enable systemd service for StratusShell.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("not yet implemented")
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
	installCmd.Flags().StringP("user", "u", "", "User that service runs as (required)")
	installCmd.MarkFlagRequired("user")
	installCmd.Flags().IntP("port", "p", 8080, "HTTP port")
}
