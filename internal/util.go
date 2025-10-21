package internal

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
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
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "mode=") {
			mode := strings.TrimPrefix(line, "mode=")
			if mode == "client" || mode == "server" {
				return mode
			}
		}
	}

	return ""
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

// ExtractFlagValue extracts a flag value from arguments
func ExtractFlagValue(args []string, flag string) string {
	for i := 0; i < len(args); i++ {
		if args[i] == flag && i+1 < len(args) {
			return args[i+1]
		}
	}
	return ""
}

// ExtractAppName extracts app name from arguments (no environment concept)
func ExtractAppName(args []string) string {
	app, _ := ExtractAppFlag(args)
	return app
}
