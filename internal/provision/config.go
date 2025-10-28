package provision

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the full provisioning configuration
type Config struct {
	User      UserConfig      `yaml:"user"`
	Base      []string        `yaml:"base_packages"`
	Cloud     CloudConfig     `yaml:"cloud"`
	Languages LanguagesConfig `yaml:"languages"`
	Shell     ShellConfig     `yaml:"shell"`
}

// UserConfig contains user-specific settings
type UserConfig struct {
	Shell string `yaml:"shell"`
}

// CloudConfig contains cloud tool installation flags
type CloudConfig struct {
	AWS       bool `yaml:"aws"`
	GCloud    bool `yaml:"gcloud"`
	Kubectl   bool `yaml:"kubectl"`
	Docker    bool `yaml:"docker"`
	Terraform bool `yaml:"terraform"`
}

// LanguagesConfig contains language toolchain configurations
type LanguagesConfig struct {
	Go   GoConfig   `yaml:"go"`
	Node NodeConfig `yaml:"node"`
}

// GoConfig contains Go toolchain configuration
type GoConfig struct {
	Enabled bool     `yaml:"enabled"`
	Version string   `yaml:"version"`
	Tools   []string `yaml:"tools"`
}

// NodeConfig contains Node.js toolchain configuration
type NodeConfig struct {
	Enabled        bool     `yaml:"enabled"`
	Version        string   `yaml:"version"`
	PackageManager string   `yaml:"package_manager"`
	GlobalPackages []string `yaml:"global_packages"`
}

// ShellConfig contains shell environment settings
type ShellConfig struct {
	Zsh     bool `yaml:"zsh"`
	OhMyZsh bool `yaml:"oh_my_zsh"`
	Tmux    bool `yaml:"tmux"`
}

// LoadConfig loads and parses a YAML configuration file
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &config, nil
}
