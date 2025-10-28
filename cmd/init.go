package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/corymacd/StratusShell/internal/provision"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Provision a development environment",
	Long:  `Create a system user and install complete development toolchain.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check if running as root
		if os.Geteuid() != 0 {
			return fmt.Errorf("init command must be run as root (use sudo)")
		}

		username, _ := cmd.Flags().GetString("user")
		shell, _ := cmd.Flags().GetString("shell")
		configPath, _ := cmd.Flags().GetString("config")
		skipTools, _ := cmd.Flags().GetBool("skip-tools")

		log.Printf("Provisioning user: %s", username)

		// Load config
		config, err := provision.LoadConfig(configPath)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Override shell if specified via flag
		if shell != "" {
			config.User.Shell = shell
		}

		// Create user
		log.Println("Creating user...")
		if err := provision.CreateUser(username, config.User.Shell); err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}

		// Configure passwordless sudo
		log.Println("Configuring passwordless sudo...")
		if err := provision.ConfigurePasswordlessSudo(username); err != nil {
			// Rollback user creation
			log.Printf("Sudo configuration failed, rolling back user creation...")
			provision.DeleteUser(username)
			return fmt.Errorf("failed to configure sudo: %w", err)
		}

		if skipTools {
			log.Println("Skipping tool installation (--skip-tools)")
			log.Println("✓ User provisioning complete")
			return nil
		}

		// Create provisioner
		p, err := provision.NewProvisioner(username, config)
		if err != nil {
			log.Printf("Warning: failed to create provisioner: %v", err)
			log.Println("User created but tool installation skipped")
			return nil
		}

		// Install base packages
		if err := p.InstallBasePackages(); err != nil {
			log.Printf("Warning: base package installation failed: %v", err)
		}

		// Install cloud tools
		if err := p.InstallCloudTools(); err != nil {
			log.Printf("Warning: cloud tools installation failed: %v", err)
		}

		// Install language toolchains
		if err := p.InstallLanguageToolchains(); err != nil {
			log.Printf("Warning: language toolchains installation failed: %v", err)
		}

		// Setup shell environment
		if err := p.SetupShellEnvironment(); err != nil {
			log.Printf("Warning: shell setup failed: %v", err)
		}

		log.Println("✓ Provisioning complete")
		log.Printf("User %s is ready. Database will be created at ~/.stratusshell/data.db on first serve", username)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().StringP("user", "u", "", "Username to create (required)")
	initCmd.MarkFlagRequired("user")
	initCmd.Flags().String("shell", "", "Default shell (overrides config)")
	initCmd.Flags().String("config", "/etc/stratusshell/default.yaml", "Config file path")
	initCmd.Flags().Bool("skip-tools", false, "Create user only, skip tool installation")
}
