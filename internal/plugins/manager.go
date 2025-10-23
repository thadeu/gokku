package plugins

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// PluginManager manages plugins and their lifecycle
type PluginManager struct {
	pluginsDir string
}

// GetPluginsDir returns the plugins directory
func (pm *PluginManager) GetPluginsDir() string {
	return pm.pluginsDir
}

// NewPluginManager creates a new plugin manager
func NewPluginManager() *PluginManager {
	// For development, use local directory if /opt/gokku doesn't exist
	pluginsDir := "/opt/gokku/plugins"

	if _, err := os.Stat("/opt/gokku"); os.IsNotExist(err) {
		pluginsDir = "dev-plugins"
		os.MkdirAll(pluginsDir, 0755)
	}

	return &PluginManager{
		pluginsDir: pluginsDir,
	}
}

// InstallOfficialPlugin installs an official plugin from gokku-vm organization
func (pm *PluginManager) InstallOfficialPlugin(pluginName string) error {
	// Check if plugin already exists
	if pm.pluginExists(pluginName) {
		return fmt.Errorf("plugin '%s' already exists", pluginName)
	}

	// Create plugin directory
	pluginDir := filepath.Join(pm.pluginsDir, pluginName)
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		return fmt.Errorf("failed to create plugin directory: %v", err)
	}

	// Official plugins are in gokku-vm organization with gokku- prefix
	repoName := fmt.Sprintf("gokku-%s", pluginName)
	gitURL := fmt.Sprintf("https://github.com/gokku-vm/%s", repoName)

	// Clone repository
	if err := pm.cloneRepository(gitURL, pluginDir); err != nil {
		// Cleanup on error
		os.RemoveAll(pluginDir)
		return fmt.Errorf("failed to clone official plugin: %v", err)
	}

	// Create plugin config.json
	if err := pm.createPluginConfig(pluginDir, pluginName, gitURL); err != nil {
		return fmt.Errorf("failed to create plugin config: %v", err)
	}

	// Make scripts executable
	if err := pm.makeScriptsExecutable(pluginDir); err != nil {
		return fmt.Errorf("failed to make scripts executable: %v", err)
	}

	return nil
}

// ListPlugins returns a list of installed plugins
func (pm *PluginManager) ListPlugins() ([]string, error) {
	var plugins []string

	entries, err := os.ReadDir(pm.pluginsDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			plugins = append(plugins, entry.Name())
		}
	}

	return plugins, nil
}

// RemovePlugin removes a plugin
func (pm *PluginManager) RemovePlugin(pluginName string) error {
	if !pm.pluginExists(pluginName) {
		return fmt.Errorf("plugin '%s' not found", pluginName)
	}

	pluginDir := filepath.Join(pm.pluginsDir, pluginName)
	return os.RemoveAll(pluginDir)
}

// GetPluginCommands returns available commands for a plugin
func (pm *PluginManager) GetPluginCommands(pluginName string) ([]string, error) {
	commandsDir := filepath.Join(pm.pluginsDir, pluginName, "commands")

	var commands []string
	entries, err := os.ReadDir(commandsDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			commands = append(commands, entry.Name())
		}
	}

	return commands, nil
}

// PluginExists checks if a plugin exists
func (pm *PluginManager) PluginExists(pluginName string) bool {
	return pm.pluginExists(pluginName)
}

// CommandExists checks if a plugin command exists
func (pm *PluginManager) CommandExists(pluginName, command string) bool {
	commandPath := filepath.Join(pm.pluginsDir, pluginName, "commands", command)
	_, err := os.Stat(commandPath)
	return !os.IsNotExist(err)
}

// makeScriptsExecutable makes all plugin scripts executable
func (pm *PluginManager) makeScriptsExecutable(pluginDir string) error {
	// Make bin/install script executable
	installPath := filepath.Join(pluginDir, "bin", "install")
	if _, err := os.Stat(installPath); err == nil {
		if err := os.Chmod(installPath, 0755); err != nil {
			return fmt.Errorf("failed to make install script executable: %v", err)
		}
	}

	// Make bin/uninstall script executable
	uninstallPath := filepath.Join(pluginDir, "bin", "uninstall")
	if _, err := os.Stat(uninstallPath); err == nil {
		if err := os.Chmod(uninstallPath, 0755); err != nil {
			return fmt.Errorf("failed to make uninstall script executable: %v", err)
		}
	}

	// Make all command scripts executable
	commandsDir := filepath.Join(pluginDir, "commands")
	if _, err := os.Stat(commandsDir); err == nil {
		entries, err := os.ReadDir(commandsDir)
		if err != nil {
			return fmt.Errorf("failed to read commands directory: %v", err)
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				commandPath := filepath.Join(commandsDir, entry.Name())
				if err := os.Chmod(commandPath, 0755); err != nil {
					return fmt.Errorf("failed to make command script executable: %v", err)
				}
			}
		}
	}

	return nil
}

// pluginExists checks if a plugin exists
func (pm *PluginManager) pluginExists(pluginName string) bool {
	pluginPath := filepath.Join(pm.pluginsDir, pluginName)
	_, err := os.Stat(pluginPath)
	return !os.IsNotExist(err)
}

// IsValidGitURL checks if the string is a valid Git URL
func (pm *PluginManager) IsValidGitURL(url string) bool {
	// Check for common Git URL patterns
	patterns := []string{
		"https://github.com/",
		"https://gitlab.com/",
		"https://bitbucket.org/",
		"git@github.com:",
		"git@gitlab.com:",
		"git@bitbucket.org:",
		"git://",
		"ssh://",
	}

	for _, pattern := range patterns {
		if strings.HasPrefix(url, pattern) {
			return true
		}
	}

	return false
}

// ExtractPluginNameFromURL extracts plugin name from Git URL
func (pm *PluginManager) ExtractPluginNameFromURL(url string) string {
	// Extract repository name from URL
	// https://github.com/user/gokku-nginx -> nginx
	// https://gitlab.com/user/gokku-postgres -> postgres

	parts := strings.Split(url, "/")
	if len(parts) == 0 {
		return "unknown"
	}

	repoName := parts[len(parts)-1]

	// Remove .git suffix if present
	repoName = strings.TrimSuffix(repoName, ".git")

	// Remove gokku- prefix if present
	pluginName := strings.TrimPrefix(repoName, "gokku-")

	return pluginName
}

// InstallPluginFromGit installs a plugin from Git repository
func (pm *PluginManager) InstallPluginFromGit(gitURL, pluginName string) error {
	// Check if plugin already exists
	if pm.pluginExists(pluginName) {
		return fmt.Errorf("plugin '%s' already exists", pluginName)
	}

	// Create plugin directory
	pluginDir := filepath.Join(pm.pluginsDir, pluginName)
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		return fmt.Errorf("failed to create plugin directory: %v", err)
	}

	// Clone repository
	if err := pm.cloneRepository(gitURL, pluginDir); err != nil {
		// Cleanup on error
		os.RemoveAll(pluginDir)
		return fmt.Errorf("failed to clone repository: %v", err)
	}

	// Create plugin config.json
	if err := pm.createPluginConfig(pluginDir, pluginName, gitURL); err != nil {
		return fmt.Errorf("failed to create plugin config: %v", err)
	}

	// Make scripts executable
	if err := pm.makeScriptsExecutable(pluginDir); err != nil {
		return fmt.Errorf("failed to make scripts executable: %v", err)
	}

	return nil
}

// cloneRepository clones a Git repository
func (pm *PluginManager) cloneRepository(gitURL, targetDir string) error {
	// Use git clone command
	cmd := exec.Command("git", "clone", "--depth", "1", gitURL, targetDir)

	// Capture output for better error messages
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git clone failed: %v\nOutput: %s", err, string(output))
	}

	return nil
}

// createPluginConfig creates a config.json file for the plugin
func (pm *PluginManager) createPluginConfig(pluginDir, pluginName, gitURL string) error {
	configJSON := fmt.Sprintf(`{
  "name": "%s",
  "url": "%s"
}`, pluginName, gitURL)

	configPath := filepath.Join(pluginDir, "config.json")
	return os.WriteFile(configPath, []byte(configJSON), 0644)
}

// UpdatePlugin updates a plugin by re-cloning from its source URL
func (pm *PluginManager) UpdatePlugin(pluginName string) error {
	// Check if plugin exists
	if !pm.pluginExists(pluginName) {
		return fmt.Errorf("plugin '%s' not found", pluginName)
	}

	pluginDir := filepath.Join(pm.pluginsDir, pluginName)
	configPath := filepath.Join(pluginDir, "config.json")

	// Read plugin config to get source URL
	configData, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read plugin config: %v", err)
	}

	// Extract URL from config (simple parsing)
	configStr := string(configData)
	var gitURL string

	// Find URL in config.json
	lines := strings.Split(configStr, "\n")
	for _, line := range lines {
		if strings.Contains(line, `"url":`) {
			// Extract URL from line like: "url": "https://github.com/user/repo"
			parts := strings.Split(line, `"url": "`)
			if len(parts) > 1 {
				gitURL = strings.TrimSuffix(strings.TrimSpace(parts[1]), `"`)
				break
			}
		}
	}

	if gitURL == "" {
		return fmt.Errorf("plugin source URL not found in config.json")
	}

	// Remove existing plugin directory
	if err := os.RemoveAll(pluginDir); err != nil {
		return fmt.Errorf("failed to remove existing plugin: %v", err)
	}

	// Recreate plugin directory
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		return fmt.Errorf("failed to create plugin directory: %v", err)
	}

	// Clone repository again
	if err := pm.cloneRepository(gitURL, pluginDir); err != nil {
		// Cleanup on error
		os.RemoveAll(pluginDir)
		return fmt.Errorf("failed to clone repository: %v", err)
	}

	// Recreate plugin config.json
	if err := pm.createPluginConfig(pluginDir, pluginName, gitURL); err != nil {
		return fmt.Errorf("failed to create plugin config: %v", err)
	}

	// Make scripts executable
	if err := pm.makeScriptsExecutable(pluginDir); err != nil {
		return fmt.Errorf("failed to make scripts executable: %v", err)
	}

	return nil
}
