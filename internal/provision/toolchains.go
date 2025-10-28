package provision

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/corymacd/StratusShell/internal/audit"
	"github.com/corymacd/StratusShell/internal/validation"
)

// InstallLanguageToolchains installs configured language toolchains
func (p *Provisioner) InstallLanguageToolchains() error {
	log.Println("Installing language toolchains...")

	if p.config.Languages.Go.Enabled {
		if err := p.installGoToolchain(); err != nil {
			log.Printf("Warning: failed to install Go toolchain: %v", err)
		}
	}

	if p.config.Languages.Node.Enabled {
		if err := p.installNodeToolchain(); err != nil {
			log.Printf("Warning: failed to install Node toolchain: %v", err)
		}
	}

	return nil
}

// installGoToolchain installs Go and related tools
func (p *Provisioner) installGoToolchain() error {
	log.Println("Installing Go toolchain...")

	// Install Go via package manager
	err := p.pm.Install("golang")
	if err != nil {
		// Try alternative package name
		err = p.pm.Install("go")
	}

	if err != nil {
		auditLogger.Log(audit.Entry{
			Action:  audit.ActionToolInstall,
			Actor:   "system",
			Target:  p.username,
			Outcome: audit.OutcomeFailure,
			Error:   fmt.Sprintf("go installation failed: %v", err),
			Details: map[string]interface{}{
				"tool": "golang",
			},
		})
		return fmt.Errorf("failed to install go: %w", err)
	}

	// Get user home directory
	homeDir, err := GetUserHomeDir(p.username)
	if err != nil {
		auditLogger.Log(audit.Entry{
			Action:  audit.ActionToolInstall,
			Actor:   "system",
			Target:  p.username,
			Outcome: audit.OutcomeFailure,
			Error:   fmt.Sprintf("failed to get home directory: %v", err),
			Details: map[string]interface{}{
				"tool": "golang",
			},
		})
		return err
	}

	// Install Go tools
	for _, tool := range p.config.Languages.Go.Tools {
		if err := p.installGoTool(tool); err != nil {
			log.Printf("Warning: failed to install Go tool %s: %v", tool, err)
		}
	}

	// Create stratusshell env file
	envDir := filepath.Join(homeDir, ".stratusshell")
	if err := validation.ValidateWorkingDir(envDir); err != nil {
		auditLogger.Log(audit.Entry{
			Action:  audit.ActionToolInstall,
			Actor:   "system",
			Target:  p.username,
			Outcome: audit.OutcomeFailure,
			Error:   fmt.Sprintf("env directory validation failed: %v", err),
		})
		return fmt.Errorf("invalid env directory: %w", err)
	}

	envFile := filepath.Join(envDir, "env.sh")
	envContent := `# StratusShell Go Environment
export GOPATH=$HOME/go
export PATH=$PATH:/usr/local/go/bin:$GOPATH/bin
`

	if err := os.MkdirAll(envDir, 0755); err != nil {
		auditLogger.Log(audit.Entry{
			Action:  audit.ActionToolInstall,
			Actor:   "system",
			Target:  p.username,
			Outcome: audit.OutcomeFailure,
			Error:   fmt.Sprintf("failed to create env directory: %v", err),
		})
		return err
	}

	if err := os.WriteFile(envFile, []byte(envContent), 0644); err != nil {
		auditLogger.Log(audit.Entry{
			Action:  audit.ActionToolInstall,
			Actor:   "system",
			Target:  p.username,
			Outcome: audit.OutcomeFailure,
			Error:   fmt.Sprintf("failed to write env file: %v", err),
		})
		return err
	}

	// Set ownership
	if err := ChownRecursive(envDir, p.username); err != nil {
		return err
	}

	auditLogger.Log(audit.Entry{
		Action:  audit.ActionToolInstall,
		Actor:   "system",
		Target:  p.username,
		Outcome: audit.OutcomeSuccess,
		Details: map[string]interface{}{
			"tool":     "golang",
			"go_tools": p.config.Languages.Go.Tools,
			"env_file": envFile,
		},
	})

	return nil
}

// installGoTool installs a specific Go tool
func (p *Provisioner) installGoTool(tool string) error {
	var packagePath string

	switch tool {
	case "golangci-lint":
		packagePath = "github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
	case "gopls":
		packagePath = "golang.org/x/tools/gopls@latest"
	case "delve":
		packagePath = "github.com/go-delve/delve/cmd/dlv@latest"
	default:
		return fmt.Errorf("unknown go tool: %s", tool)
	}

	cmd := exec.Command("sudo", "-u", p.username, "go", "install", packagePath)
	if err := cmd.Run(); err != nil {
		auditLogger.Log(audit.Entry{
			Action:  audit.ActionToolInstall,
			Actor:   "system",
			Target:  p.username,
			Outcome: audit.OutcomeFailure,
			Error:   fmt.Sprintf("go tool installation failed: %v", err),
			Details: map[string]interface{}{
				"go_tool": tool,
				"package": packagePath,
			},
		})
		return err
	}

	auditLogger.Log(audit.Entry{
		Action:  audit.ActionToolInstall,
		Actor:   "system",
		Target:  p.username,
		Outcome: audit.OutcomeSuccess,
		Details: map[string]interface{}{
			"go_tool": tool,
			"package": packagePath,
		},
	})

	return nil
}

// installNodeToolchain installs Node.js and related tools
func (p *Provisioner) installNodeToolchain() error {
	log.Println("Installing Node toolchain...")

	homeDir, err := GetUserHomeDir(p.username)
	if err != nil {
		auditLogger.Log(audit.Entry{
			Action:  audit.ActionToolInstall,
			Actor:   "system",
			Target:  p.username,
			Outcome: audit.OutcomeFailure,
			Error:   fmt.Sprintf("failed to get home directory: %v", err),
			Details: map[string]interface{}{
				"tool": "nodejs",
			},
		})
		return err
	}

	// Create .nvm directory
	nvmDir := filepath.Join(homeDir, ".nvm")
	if err := validation.ValidateWorkingDir(nvmDir); err != nil {
		auditLogger.Log(audit.Entry{
			Action:  audit.ActionToolInstall,
			Actor:   "system",
			Target:  p.username,
			Outcome: audit.OutcomeFailure,
			Error:   fmt.Sprintf("nvm directory validation failed: %v", err),
		})
		return fmt.Errorf("invalid nvm directory: %w", err)
	}

	if err := os.MkdirAll(nvmDir, 0755); err != nil {
		auditLogger.Log(audit.Entry{
			Action:  audit.ActionToolInstall,
			Actor:   "system",
			Target:  p.username,
			Outcome: audit.OutcomeFailure,
			Error:   fmt.Sprintf("failed to create nvm directory: %v", err),
		})
		return err
	}

	// Install Node.js via package manager (simplified approach)
	err = p.pm.Install("nodejs", "npm")
	if err != nil {
		auditLogger.Log(audit.Entry{
			Action:  audit.ActionToolInstall,
			Actor:   "system",
			Target:  p.username,
			Outcome: audit.OutcomeFailure,
			Error:   fmt.Sprintf("nodejs installation failed: %v", err),
		})
		return fmt.Errorf("failed to install nodejs: %w", err)
	}

	// Install global packages
	for _, pkg := range p.config.Languages.Node.GlobalPackages {
		if err := p.installNpmGlobal(pkg); err != nil {
			log.Printf("Warning: failed to install npm package %s: %v", pkg, err)
		}
	}

	// Install pnpm if configured
	if p.config.Languages.Node.PackageManager == "pnpm" {
		if err := p.installNpmGlobal("pnpm"); err != nil {
			log.Printf("Warning: failed to install pnpm: %v", err)
		}
	}

	// Set ownership
	if err := ChownRecursive(nvmDir, p.username); err != nil {
		return err
	}

	auditLogger.Log(audit.Entry{
		Action:  audit.ActionToolInstall,
		Actor:   "system",
		Target:  p.username,
		Outcome: audit.OutcomeSuccess,
		Details: map[string]interface{}{
			"tool":            "nodejs",
			"global_packages": p.config.Languages.Node.GlobalPackages,
			"package_manager": p.config.Languages.Node.PackageManager,
		},
	})

	return nil
}

// installNpmGlobal installs a global npm package
func (p *Provisioner) installNpmGlobal(packageName string) error {
	cmd := exec.Command("npm", "install", "-g", packageName)
	if err := cmd.Run(); err != nil {
		auditLogger.Log(audit.Entry{
			Action:  audit.ActionToolInstall,
			Actor:   "system",
			Target:  p.username,
			Outcome: audit.OutcomeFailure,
			Error:   fmt.Sprintf("npm package installation failed: %v", err),
			Details: map[string]interface{}{
				"npm_package": packageName,
			},
		})
		return err
	}

	auditLogger.Log(audit.Entry{
		Action:  audit.ActionToolInstall,
		Actor:   "system",
		Target:  p.username,
		Outcome: audit.OutcomeSuccess,
		Details: map[string]interface{}{
			"npm_package": packageName,
		},
	})

	return nil
}
