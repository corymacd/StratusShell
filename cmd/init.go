package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Provision a development environment",
	Long:  `Create a system user and install complete development toolchain.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("not yet implemented")
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().StringP("user", "u", "", "Username to create (required)")
	initCmd.MarkFlagRequired("user")
	initCmd.Flags().String("shell", "/bin/bash", "Default shell")
	initCmd.Flags().String("config", "/etc/stratusshell/default.yaml", "Config file path")
	initCmd.Flags().Bool("skip-tools", false, "Create user only, skip tool installation")
}
