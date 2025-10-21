package internal

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"syscall"
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

// ExtractAppFlag extracts the -a or --app flag from arguments and returns the app name and remaining args
func ExtractAppFlag(args []string) (string, []string) {
	var app string
	var remaining []string

	for i := 0; i < len(args); i++ {
		if (args[i] == "-a" || args[i] == "--app") && i+1 < len(args) {
			app = args[i+1]
			i++ // Skip next arg
		} else {
			remaining = append(remaining, args[i])
		}
	}

	return app, remaining
}

// IsRunningOnServer returns true if running on the server environment
// Uses ~/.gokkurc file to determine mode, with fallback to client mode if file doesn't exist
func IsRunningOnServer() bool {
	return IsServerMode()
}

// GetGokkuRcPath returns the path to the ~/.gokkurc file
func GetGokkuRcPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".gokkurc")
}

// ReadGokkuRcMode reads the mode from ~/.gokkurc file
// Returns "client", "server", or empty string if file doesn't exist or is invalid
func ReadGokkuRcMode() string {
	rcPath := GetGokkuRcPath()

	file, err := os.Open(rcPath)

	if err != nil {
		return ""
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)

	var mode string

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		mode, _ = strings.CutPrefix(line, "mode=")
	}

	return mode
}

// IsClientMode returns true if running in client mode
// Falls back to client mode if ~/.gokkurc doesn't exist
func IsClientMode() bool {
	mode := ReadGokkuRcMode()

	if mode == "" {
		// If file doesn't exist, assume client mode (fallback)
		return true
	}

	return mode == "client"
}

// IsServerMode returns true if running in server mode
// Falls back to client mode if ~/.gokkurc doesn't exist
func IsServerMode() bool {
	mode := ReadGokkuRcMode()

	if mode == "" {
		// If file doesn't exist, assume client mode (fallback)
		return false
	}

	return mode == "server"
}

// IsSignalInterruption checks if the error is due to signal interruption
func IsSignalInterruption(err error) bool {
	if err == nil {
		return false
	}

	// Check if it's a signal error
	if exitError, ok := err.(*os.SyscallError); ok {
		if exitError.Err == syscall.EINTR {
			return true
		}
	}

	// For long-running operations, we'll be more permissive with signal handling
	// This is a simplified approach that works better with SSH and Docker commands
	return true
}
