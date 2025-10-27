package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "stratusshell",
	Short: "Cloud development environment provisioning tool",
	Long: `StratusShell provisions complete cloud development environments with
user creation, tool installation, and web-based terminal management.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Subcommands will be added here
}
