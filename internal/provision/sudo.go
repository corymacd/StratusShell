package provision

import (
	"fmt"
	"os"
	"path/filepath"
)

func ConfigurePasswordlessSudo(username string) error {
	sudoersFile := filepath.Join("/etc/sudoers.d", fmt.Sprintf("stratusshell-%s", username))

	content := fmt.Sprintf("%s ALL=(ALL) NOPASSWD:ALL\n", username)

	// Write with 0440 permissions (required for sudoers files)
	if err := os.WriteFile(sudoersFile, []byte(content), 0440); err != nil {
		return fmt.Errorf("failed to write sudoers file: %w", err)
	}

	return nil
}

func RemoveSudoersConfig(username string) error {
	sudoersFile := filepath.Join("/etc/sudoers.d", fmt.Sprintf("stratusshell-%s", username))
	return os.Remove(sudoersFile)
}
