package provision

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
)

func CreateUser(username, shell string) error {
	// Check if user already exists
	if _, err := user.Lookup(username); err == nil {
		return fmt.Errorf("user %s already exists", username)
	}

	// Create user with home directory
	cmd := exec.Command("useradd",
		"-m",                    // Create home directory
		"-s", shell,            // Set shell
		"-c", "StratusShell User", // Comment
		username,
	)

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to create user: %w (output: %s)", err, output)
	}

	return nil
}

func DeleteUser(username string) error {
	cmd := exec.Command("userdel", "-r", username)
	return cmd.Run()
}

func UserExists(username string) bool {
	_, err := user.Lookup(username)
	return err == nil
}

func SetUserShell(username, shell string) error {
	cmd := exec.Command("chsh", "-s", shell, username)
	return cmd.Run()
}

func GetUserHomeDir(username string) (string, error) {
	u, err := user.Lookup(username)
	if err != nil {
		return "", err
	}
	return u.HomeDir, nil
}

func ChownRecursive(path, username string) error {
	u, err := user.Lookup(username)
	if err != nil {
		return err
	}

	cmd := exec.Command("chown", "-R", fmt.Sprintf("%s:%s", u.Uid, u.Gid), path)
	return cmd.Run()
}
