package provision

import (
	"errors"
	"os/exec"
)

type PackageManager int

const (
	APT PackageManager = iota
	YUM
	DNF
	PACMAN
)

func (pm PackageManager) String() string {
	switch pm {
	case APT:
		return "apt"
	case YUM:
		return "yum"
	case DNF:
		return "dnf"
	case PACMAN:
		return "pacman"
	default:
		return "unknown"
	}
}

func DetectPackageManager() (PackageManager, error) {
	managers := []struct {
		pm      PackageManager
		command string
	}{
		{APT, "apt-get"},
		{DNF, "dnf"},
		{YUM, "yum"},
		{PACMAN, "pacman"},
	}

	for _, m := range managers {
		if _, err := exec.LookPath(m.command); err == nil {
			return m.pm, nil
		}
	}

	return 0, errors.New("no supported package manager found")
}

func (pm PackageManager) Install(packages ...string) error {
	var cmd *exec.Cmd

	switch pm {
	case APT:
		args := append([]string{"install", "-y"}, packages...)
		cmd = exec.Command("apt-get", args...)
	case YUM:
		args := append([]string{"install", "-y"}, packages...)
		cmd = exec.Command("yum", args...)
	case DNF:
		args := append([]string{"install", "-y"}, packages...)
		cmd = exec.Command("dnf", args...)
	case PACMAN:
		args := append([]string{"-S", "--noconfirm"}, packages...)
		cmd = exec.Command("pacman", args...)
	default:
		return errors.New("unsupported package manager")
	}

	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run()
}
