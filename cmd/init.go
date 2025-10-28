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

		username, err := cmd.Flags().GetString("user")
		if err != nil {
			return fmt.Errorf("failed to get user flag: %w", err)
		}
		shell, err := cmd.Flags().GetString("shell")
		if err != nil {
			return fmt.Errorf("failed to get shell flag: %w", err)
		}
		configPath, err := cmd.Flags().GetString("config")
		if err != nil {
			return fmt.Errorf("failed to get config flag: %w", err)
		}
		skipTools, err := cmd.Flags().GetBool("skip-tools")
		if err != nil {
			return fmt.Errorf("failed to get skip-tools flag: %w", err)
		}

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
			if delErr := provision.DeleteUser(username); delErr != nil {
				log.Printf("ERROR: Failed to rollback user creation: %v", delErr)
				log.Printf("Please manually remove user '%s' with: sudo userdel -r %s", username, username)
			}
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
			log.Printf("ERROR: Failed to create provisioner: %v", err)
			log.Printf("User '%s' was created but tool installation will be skipped", username)
			log.Println("You may need to manually install tools or run provisioning again")
			// Return error to indicate partial failure
			return fmt.Errorf("provisioner creation failed: %w", err)
		}

		var provisionErrors []string

		// Install base packages
		if err := p.InstallBasePackages(); err != nil {
			log.Printf("Warning: base package installation failed: %v", err)
			provisionErrors = append(provisionErrors, "base packages")
		}

		// Install cloud tools
		if err := p.InstallCloudTools(); err != nil {
			log.Printf("Warning: cloud tools installation failed: %v", err)
			provisionErrors = append(provisionErrors, "cloud tools")
		}

		// Install language toolchains
		if err := p.InstallLanguageToolchains(); err != nil {
			log.Printf("Warning: language toolchains installation failed: %v", err)
			provisionErrors = append(provisionErrors, "language toolchains")
		}

		// Setup shell environment
		if err := p.SetupShellEnvironment(); err != nil {
			log.Printf("Warning: shell setup failed: %v", err)
			provisionErrors = append(provisionErrors, "shell environment")
		}

		if len(provisionErrors) > 0 {
			log.Printf("✗ Provisioning completed with errors in: %s.", provisionErrors)
			return fmt.Errorf("one or more provisioning steps failed")
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
	initCmd.Flags().String("shell", "", "Shell to use (overrides config if specified)")
	initCmd.Flags().String("config", "/etc/stratusshell/default.yaml", "Config file path")
	initCmd.Flags().Bool("skip-tools", false, "Create user only, skip tool installation")
}
