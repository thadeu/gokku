package internal

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
)

func ReadVersionFile() string {
	file, err := os.ReadFile(".goversion")
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(file))
}

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

// IsRegistryImage checks if the image is from a registry (pre-built image)
// Returns true if the image contains a registry URL (ghcr.io, ECR, docker.io, etc.)
// customRegistries is an optional list of custom registry patterns from gokku.yml
func IsRegistryImage(image string, customRegistries ...[]string) bool {
	if image == "" {
		return false
	}

	// Default registry patterns
	registryPatterns := []string{
		"ghcr.io/",
		"docker.io/",
		"registry.hub.docker.com/",
		"quay.io/",
		"gcr.io/",
		"us.gcr.io/",
		"eu.gcr.io/",
		"asia.gcr.io/",
		"k8s.gcr.io/",
		"registry.k8s.io/",
		"amazonaws.com/",
		"public.ecr.aws/",
		"registry.ecr.",
		"azurecr.io/",
		"registry.azurecr.io/",
		"registry.redhat.io/",
		"registry.access.redhat.com/",
		"registry.connect.redhat.com/",
		"registry.developers.redhat.com/",
	}

	// Add custom registries if provided
	if len(customRegistries) > 0 && len(customRegistries[0]) > 0 {
		for _, customRegistry := range customRegistries[0] {
			// Ensure custom registry ends with / for pattern matching
			if !strings.HasSuffix(customRegistry, "/") {
				customRegistry += "/"
			}
			registryPatterns = append(registryPatterns, customRegistry)
		}
	}

	// Check if image contains any registry pattern
	for _, pattern := range registryPatterns {
		if strings.Contains(image, pattern) {
			return true
		}
	}

	// Check if image contains a port (e.g., registry.example.com:5000)
	if strings.Contains(image, ":") && !strings.Contains(image, "/") {
		// This might be a local registry with port
		return false
	}

	// Check if image contains a domain-like pattern (contains dots and slashes)
	parts := strings.Split(image, "/")
	if len(parts) > 1 && strings.Contains(parts[0], ".") {
		return true
	}

	return false
}

// PullRegistryImage pulls a pre-built image from a registry
func PullRegistryImage(image string) error {
	fmt.Printf("-----> Pulling pre-built image: %s\n", image)

	cmd := exec.Command("docker", "pull", image)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to pull image %s: %v", image, err)
	}

	fmt.Printf("-----> Successfully pulled image: %s\n", image)
	return nil
}

// TagImageForApp tags a pulled image with the app name for deployment
func TagImageForApp(image, appName string) error {
	tag := fmt.Sprintf("%s:latest", appName)
	fmt.Printf("-----> Tagging image %s as %s\n", image, tag)

	cmd := exec.Command("docker", "tag", image, tag)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to tag image %s as %s: %v", image, tag, err)
	}

	fmt.Printf("-----> Successfully tagged image as %s\n", tag)
	return nil
}

// GetCustomRegistries loads custom registries from gokku.yml configuration for a specific app
func GetCustomRegistries(appName string) []string {
	// Load server config for the specific app
	if config, err := LoadServerConfigByApp(appName); err == nil && config.Docker != nil && len(config.Docker.Registry) > 0 {
		// Return the list of registries from docker.registry config
		return config.Docker.Registry
	}

	// Fallback to empty list if no config found
	return []string{}
}

// DetectRubyVersion detects Ruby version from .ruby-version or Gemfile
func DetectRubyVersion(releaseDir string) string {
	// Try .ruby-version first
	rubyVersionPath := filepath.Join(releaseDir, ".ruby-version")
	if data, err := os.ReadFile(rubyVersionPath); err == nil {
		version := strings.TrimSpace(string(data))
		if version != "" {
			return fmt.Sprintf("ruby:%s", version)
		}
	}

	// Try Gemfile
	gemfilePath := filepath.Join(releaseDir, "Gemfile")
	if data, err := os.ReadFile(gemfilePath); err == nil {
		content := string(data)
		// Look for ruby version in Gemfile
		re := regexp.MustCompile(`ruby\s+["']([^"']+)["']`)
		matches := re.FindStringSubmatch(content)
		if len(matches) > 1 {
			version := matches[1]
			// Convert version format (e.g., "3.1" -> "3.1")
			return fmt.Sprintf("ruby:%s", version)
		}
	}

	// Fallback to latest Ruby
	return "ruby:latest"
}

// DetectGoVersion detects Go version from go.mod
func DetectGoVersion(releaseDir string) string {
	goModPath := filepath.Join(releaseDir, "go.mod")
	if data, err := os.ReadFile(goModPath); err == nil {
		content := string(data)
		// Look for go version in go.mod
		re := regexp.MustCompile(`go\s+(\d+\.\d+)`)
		matches := re.FindStringSubmatch(content)

		if len(matches) > 1 {
			version := matches[1]
			return fmt.Sprintf("golang:%s-alpine", version)
		}
	}

	// Fallback to latest Go
	return "golang:latest"
}

// DetectNodeVersion detects Node.js version from .nvmrc or package.json
func DetectNodeVersion(releaseDir string) string {
	// Try .nvmrc first
	nvmrcPath := filepath.Join(releaseDir, ".nvmrc")
	if data, err := os.ReadFile(nvmrcPath); err == nil {
		version := strings.TrimSpace(string(data))
		if version != "" {
			// Remove 'v' prefix if present
			version = strings.TrimPrefix(version, "v")
			return fmt.Sprintf("node:%s", version)
		}
	}

	// Try package.json
	packageJsonPath := filepath.Join(releaseDir, "package.json")
	if data, err := os.ReadFile(packageJsonPath); err == nil {
		content := string(data)
		// Look for engines.node in package.json
		re := regexp.MustCompile(`"engines"\s*:\s*{[^}]*"node"\s*:\s*"([^"]+)"`)
		matches := re.FindStringSubmatch(content)
		if len(matches) > 1 {
			version := matches[1]
			// Remove version constraints (e.g., ">=18.0.0" -> "18")
			re2 := regexp.MustCompile(`(\d+)`)
			if versionMatch := re2.FindStringSubmatch(version); len(versionMatch) > 1 {
				return fmt.Sprintf("node:%s", versionMatch[1])
			}
		}
	}

	// Fallback to latest Node.js
	return "node:latest"
}

// DetectPythonVersion returns latest Python version (as requested)
func DetectPythonVersion(releaseDir string) string {
	// Always use latest Python as fallback
	return "python:latest"
}
