package pkg

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"gokku/pkg/git"
)

// GetRemoteInfoOrDefault extracts remote info using --remote flag or defaults to "gokku"
// Returns nil if in server mode (local execution)
// Returns RemoteInfo if in client mode with valid remote
func GetRemoteInfoOrDefault(args []string) (*RemoteInfo, []string, error) {
	// If in server mode, return nil (execute locally)
	if IsServerMode() {
		// Remove --remote from args if present
		_, remainingArgs := ExtractRemoteFlag(args)
		return nil, remainingArgs, nil
	}

	// Client mode: extract --remote flag
	remote, remainingArgs := ExtractRemoteFlag(args)

	// If no remote specified, try to use "gokku" as default
	if remote == "" {
		// Try to get "gokku" remote first
		gokkuRemoteInfo, err := GetRemoteInfo("gokku")

		if err == nil {
			// Remote "gokku" exists, use it
			return gokkuRemoteInfo, remainingArgs, nil
		}

		// No "gokku" remote found, return error
		return nil, remainingArgs, fmt.Errorf("no remote specified and default remote 'gokku' not found. Run 'gokku remote setup user@server_ip' first")
	}

	// Remote specified, get its info
	remoteInfo, err := GetRemoteInfo(remote)
	if err != nil {
		return nil, remainingArgs, fmt.Errorf("remote '%s' not found: %v. Add it with: gokku remote add %s user@host", remote, err, remote)
	}

	return remoteInfo, remainingArgs, nil
}

// ExecuteRemoteCommand executes a command on remote server via SSH
// Automatically removes --remote flag from the command string
func ExecuteRemoteCommand(remoteInfo *RemoteInfo, command string) error {
	if remoteInfo == nil {
		return fmt.Errorf("remoteInfo is nil")
	}

	// Remove --remote flag from command string before executing
	cleanCommand := strings.Replace(command, " --remote", "", -1)
	cleanCommand = strings.Replace(cleanCommand, "--remote", "", -1)

	cmd := exec.Command("ssh", remoteInfo.Host, cleanCommand)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
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

// GetGokkuRcPath returns the path to the ~/.gokkurc file
func GetGokkuRcPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".gokkurc")
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

// ExtractIdentityFlag extracts the --identity flag from arguments
func ExtractIdentityFlag(args []string) (string, []string) {
	var identity string
	var remaining []string

	for i := 0; i < len(args); i++ {
		if args[i] == "--identity" && i+1 < len(args) {
			identity = args[i+1]
			i++ // Skip next arg
		} else {
			remaining = append(remaining, args[i])
		}
	}

	return identity, remaining
}

// GetRemoteInfo extracts info from git remote
// Example: ubuntu@server:api
// Returns: RemoteInfo{Host: "ubuntu@server", BaseDir: "/opt/gokku", App: "api"}
func GetRemoteInfo(remoteName string) (*RemoteInfo, error) {
	remoteInfo, err := git.GetRemoteInfoWithClient(&git.GitClient{}, remoteName)
	if err != nil {
		return nil, err
	}
	return &RemoteInfo{
		Host:    remoteInfo.Host,
		BaseDir: remoteInfo.BaseDir,
		App:     remoteInfo.App,
	}, nil
}

// LoadEnvFile loads environment variables from a file
func LoadEnvFile(envFile string) map[string]string {
	envVars := make(map[string]string)

	content, err := os.ReadFile(envFile)

	if err != nil {
		if os.IsNotExist(err) {
			return envVars // Return empty map if file doesn't exist
		}

		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	lines := strings.Split(string(content), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)

		if len(parts) == 2 {
			envVars[parts[0]] = parts[1]
		}
	}

	return envVars
}

// SaveEnvFile saves environment variables to a file
func SaveEnvFile(envFile string, envVars map[string]string) error {
	// Sort keys
	keys := make([]string, 0, len(envVars))

	for k := range envVars {
		keys = append(keys, k)
	}

	for i := 0; i < len(keys); i++ {
		for j := i + 1; j < len(keys); j++ {
			if keys[i] > keys[j] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}

	var content strings.Builder

	for _, key := range keys {
		content.WriteString(fmt.Sprintf("%s=%s\n", key, envVars[key]))
	}

	return os.WriteFile(envFile, []byte(content.String()), 0600)
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

	return false
}
