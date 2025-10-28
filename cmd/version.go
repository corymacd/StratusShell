package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	versionInfo = struct {
		version string
		commit  string
		date    string
	}{
		version: "dev",
		commit:  "none",
		date:    "unknown",
	}
)

// SetVersionInfo sets the version information from main package
func SetVersionInfo(version, commit, date string) {
	versionInfo.version = version
	versionInfo.commit = commit
	versionInfo.date = date
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Long:  `Display the version, commit, and build date of StratusShell.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("StratusShell %s\n", versionInfo.version)
		fmt.Printf("Commit: %s\n", versionInfo.commit)
		fmt.Printf("Built: %s\n", versionInfo.date)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
