package internal

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// LoadConfig loads the configuration from the config file
func LoadConfig() (*Config, error) {
	configPath := GetConfigPath()

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{Apps: []App{}}, nil
		}
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func (c *Config) GetAppConfig(appName string) *App {
	for _, app := range c.Apps {
		if app.Name == appName {
			return &app
		}
	}
	return nil
}

// SaveConfig saves the configuration to the config file
func SaveConfig(config *Config) error {
	configPath := GetConfigPath()

	// Ensure directory exists
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

// GetRemoteInfo extracts info from git remote
// Example: ubuntu@server:api
// Returns: RemoteInfo{Host: "ubuntu@server", BaseDir: "/opt/gokku", App: "api"}
func GetRemoteInfo(remoteName string) (*RemoteInfo, error) {
	// Get remote URL
	cmd := exec.Command("git", "remote", "get-url", remoteName)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git remote '%s' not found. Add it with: git remote add %s user@host:/opt/gokku/repos/<app>.git", remoteName, remoteName)
	}

	remoteURL := strings.TrimSpace(string(output))

	// Parse: user@host:/path/to/repos/app.git
	parts := strings.Split(remoteURL, ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid remote URL format: %s (expected user@host:/path)", remoteURL)
	}

	host := parts[0]
	path := parts[1]

	// Extract app name from path: api -> api
	pathParts := strings.Split(path, "/")
	if len(pathParts) < 2 {
		return nil, fmt.Errorf("invalid remote path: %s", path)
	}

	appFile := pathParts[len(pathParts)-1]         // api.git
	appName := strings.TrimSuffix(appFile, ".git") // api

	// Extract base dir: api -> /opt/gokku
	baseDir := strings.TrimSuffix(path, "/repos/"+appFile)

	return &RemoteInfo{
		Host:    host,
		BaseDir: baseDir,
		App:     appName,
	}, nil
}
