package internal

import (
	"os"
	"path/filepath"
	"runtime"
)

// GetConfigPath returns the path to the configuration file
func GetConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".gokku", "config.yml")
}

// ExtractRemoteFlag extracts the --remote flag from arguments and returns the remote name and remaining args
func ExtractRemoteFlag(args []string) (string, []string) {
	var remote string
	var remaining []string

	for i := 0; i < len(args); i++ {
		if args[i] == "--remote" && i+1 < len(args) {
			remote = args[i+1]
			i++ // Skip next arg
		} else {
			remaining = append(remaining, args[i])
		}
	}

	return remote, remaining
}

// IsRunningOnServer returns true if running on the server environment
// Server environment: Linux + systemd + /opt/gokku directory exists and is writable
func IsRunningOnServer() bool {
	// Check if running on Linux
	if runtime.GOOS != "linux" {
		return false
	}

	// Check if systemd directory exists (server indicator)
	if _, err := os.Stat("/etc/systemd/system"); os.IsNotExist(err) {
		return false
	}

	// Check if /opt/gokku exists and is writable
	info, err := os.Stat("/opt/gokku")
	if os.IsNotExist(err) {
		return false
	}
	if !info.IsDir() {
		return false
	}

	// Try to create a test file to check write permissions
	testFile := "/opt/gokku/.gokku-test-write"
	err = os.WriteFile(testFile, []byte("test"), 0644)
	if err != nil {
		return false
	}
	os.Remove(testFile) // Clean up

	return true
}
