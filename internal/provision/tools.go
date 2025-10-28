package provision

import (
	"fmt"
	"log"

	"github.com/corymacd/StratusShell/internal/audit"
)

// Provisioner handles tool installation and configuration
type Provisioner struct {
	pm       PackageManager
	username string
	config   *Config
}

// NewProvisioner creates a new Provisioner instance
func NewProvisioner(username string, config *Config) (*Provisioner, error) {
	pm, err := DetectPackageManager()
	if err != nil {
		auditLogger.Log(audit.Entry{
			Action:  audit.ActionToolInstall,
			Actor:   "system",
			Target:  username,
			Outcome: audit.OutcomeFailure,
			Error:   fmt.Sprintf("package manager detection failed: %v", err),
		})
		return nil, err
	}

	auditLogger.Log(audit.Entry{
		Action:  audit.ActionToolInstall,
		Actor:   "system",
		Target:  username,
		Outcome: audit.OutcomeSuccess,
		Details: map[string]interface{}{
			"package_manager": pm.String(),
		},
	})

	return &Provisioner{
		pm:       pm,
		username: username,
		config:   config,
	}, nil
}

// InstallBasePackages installs base development packages
func (p *Provisioner) InstallBasePackages() error {
	log.Println("Installing base packages...")

	if len(p.config.Base) == 0 {
		log.Println("No base packages configured")
		return nil
	}

	// Translate package names for different package managers
	packages := p.translatePackageNames(p.config.Base)

	auditLogger.Log(audit.Entry{
		Action: audit.ActionToolInstall,
		Actor:  "system",
		Target: p.username,
		Details: map[string]interface{}{
			"stage":    "base_packages",
			"packages": packages,
		},
	})

	if err := p.pm.Install(packages...); err != nil {
		auditLogger.Log(audit.Entry{
			Action:  audit.ActionToolInstall,
			Actor:   "system",
			Target:  p.username,
			Outcome: audit.OutcomeFailure,
			Error:   fmt.Sprintf("base package installation failed: %v", err),
			Details: map[string]interface{}{
				"packages": packages,
			},
		})
		return fmt.Errorf("failed to install base packages: %w", err)
	}

	auditLogger.Log(audit.Entry{
		Action:  audit.ActionToolInstall,
		Actor:   "system",
		Target:  p.username,
		Outcome: audit.OutcomeSuccess,
		Details: map[string]interface{}{
			"stage":    "base_packages",
			"packages": packages,
			"count":    len(packages),
		},
	})

	log.Printf("Installed %d base packages", len(packages))
	return nil
}

// translatePackageNames converts generic package names to package manager-specific names
func (p *Provisioner) translatePackageNames(packages []string) []string {
	translated := make([]string, 0, len(packages))

	for _, pkg := range packages {
		switch pkg {
		case "build-essential":
			if p.pm == YUM || p.pm == DNF {
				// RHEL/CentOS/Fedora equivalent
				translated = append(translated, "gcc", "gcc-c++", "make")
			} else if p.pm == PACMAN {
				// Arch Linux equivalent
				translated = append(translated, "base-devel")
			} else {
				// Debian/Ubuntu
				translated = append(translated, pkg)
			}
		default:
			translated = append(translated, pkg)
		}
	}

	return translated
}

// InstallCloudTools installs cloud development tools
func (p *Provisioner) InstallCloudTools() error {
	log.Println("Installing cloud tools...")

	installed := 0
	failed := 0

	if p.config.Cloud.AWS {
		if err := p.installAWSCLI(); err != nil {
			log.Printf("Warning: failed to install AWS CLI: %v", err)
			failed++
		} else {
			installed++
		}
	}

	if p.config.Cloud.Docker {
		if err := p.installDocker(); err != nil {
			log.Printf("Warning: failed to install Docker: %v", err)
			failed++
		} else {
			installed++
		}
	}

	if p.config.Cloud.Kubectl {
		if err := p.installKubectl(); err != nil {
			log.Printf("Warning: failed to install kubectl: %v", err)
			failed++
		} else {
			installed++
		}
	}

	if p.config.Cloud.Terraform {
		if err := p.installTerraform(); err != nil {
			log.Printf("Warning: failed to install Terraform: %v", err)
			failed++
		} else {
			installed++
		}
	}

	if p.config.Cloud.GCloud {
		if err := p.installGCloud(); err != nil {
			log.Printf("Warning: failed to install gcloud: %v", err)
			failed++
		} else {
			installed++
		}
	}

	total := p.countEnabledCloudTools()
	auditLogger.Log(audit.Entry{
		Action:  audit.ActionToolInstall,
		Actor:   "system",
		Target:  p.username,
		Outcome: audit.OutcomeSuccess,
		Details: map[string]interface{}{
			"stage":     "cloud_tools",
			"installed": installed,
			"failed":    failed,
			"total":     total,
		},
	})

	log.Printf("Installed %d/%d cloud tools", installed, total)
	return nil
}

// installAWSCLI installs AWS CLI
func (p *Provisioner) installAWSCLI() error {
	log.Println("Installing AWS CLI...")

	// Try to install via package manager
	err := p.pm.Install("awscli")

	auditLogger.Log(audit.Entry{
		Action:  audit.ActionToolInstall,
		Actor:   "system",
		Target:  p.username,
		Outcome: audit.OutcomeFromError(err),
		Details: map[string]interface{}{
			"tool": "awscli",
		},
		Error: audit.ErrorString(err),
	})

	return err
}

// installDocker installs Docker
func (p *Provisioner) installDocker() error {
	log.Println("Installing Docker...")

	// Try different package names based on package manager
	var err error
	if p.pm == APT {
		// Debian/Ubuntu uses docker.io or docker-ce
		err = p.pm.Install("docker.io")
		if err != nil {
			err = p.pm.Install("docker-ce")
		}
	} else {
		err = p.pm.Install("docker")
	}

	if err == nil {
		// Add user to docker group
		if groupErr := AddUserToGroup(p.username, "docker"); groupErr != nil {
			log.Printf("Warning: failed to add user %s to docker group: %v", p.username, groupErr)
			log.Printf("Note: User %s may need to be added to docker group manually", p.username)
		} else {
			log.Printf("✓ User %s added to docker group (re-login required for changes to take effect)", p.username)
		}
	}

	auditLogger.Log(audit.Entry{
		Action:  audit.ActionToolInstall,
		Actor:   "system",
		Target:  p.username,
		Outcome: audit.OutcomeFromError(err),
		Details: map[string]interface{}{
			"tool": "docker",
		},
		Error: audit.ErrorString(err),
	})

	return err
}

// installKubectl installs kubectl
func (p *Provisioner) installKubectl() error {
	log.Println("Installing kubectl...")

	err := p.pm.Install("kubectl")

	auditLogger.Log(audit.Entry{
		Action:  audit.ActionToolInstall,
		Actor:   "system",
		Target:  p.username,
		Outcome: audit.OutcomeFromError(err),
		Details: map[string]interface{}{
			"tool": "kubectl",
		},
		Error: audit.ErrorString(err),
	})

	return err
}

// installTerraform installs Terraform
func (p *Provisioner) installTerraform() error {
	log.Println("Installing Terraform...")

	err := p.pm.Install("terraform")

	auditLogger.Log(audit.Entry{
		Action:  audit.ActionToolInstall,
		Actor:   "system",
		Target:  p.username,
		Outcome: audit.OutcomeFromError(err),
		Details: map[string]interface{}{
			"tool": "terraform",
		},
		Error: audit.ErrorString(err),
	})

	return err
}

// installGCloud installs Google Cloud SDK
func (p *Provisioner) installGCloud() error {
	log.Println("Installing gcloud...")

	err := p.pm.Install("google-cloud-sdk")

	auditLogger.Log(audit.Entry{
		Action:  audit.ActionToolInstall,
		Actor:   "system",
		Target:  p.username,
		Outcome: audit.OutcomeFromError(err),
		Details: map[string]interface{}{
			"tool": "google-cloud-sdk",
		},
		Error: audit.ErrorString(err),
	})

	return err
}

// countEnabledCloudTools returns the number of enabled cloud tools
func (p *Provisioner) countEnabledCloudTools() int {
	count := 0
	if p.config.Cloud.AWS {
		count++
	}
	if p.config.Cloud.GCloud {
		count++
	}
	if p.config.Cloud.Kubectl {
		count++
	}
	if p.config.Cloud.Docker {
		count++
	}
	if p.config.Cloud.Terraform {
		count++
	}
	return count
}
