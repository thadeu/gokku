package plugins

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// PluginManager manages plugins and their lifecycle
type PluginManager struct {
	pluginsDir string
}

// NewPluginManager creates a new plugin manager
func NewPluginManager() *PluginManager {
	return &PluginManager{
		pluginsDir: "/opt/gokku/plugins",
	}
}

// DownloadPlugin downloads a plugin from GitHub
func (pm *PluginManager) DownloadPlugin(owner, repoName, pluginName string) error {
	// Check if plugin already exists
	if pm.pluginExists(pluginName) {
		return fmt.Errorf("plugin '%s' already exists", pluginName)
	}

	// Create plugin directory
	pluginDir := filepath.Join(pm.pluginsDir, pluginName)
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		return fmt.Errorf("failed to create plugin directory: %v", err)
	}

	// Download and extract plugin
	if err := pm.downloadAndExtractPlugin(owner, repoName, pluginDir); err != nil {
		// Cleanup on error
		os.RemoveAll(pluginDir)
		return fmt.Errorf("failed to download plugin: %v", err)
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

// downloadAndExtractPlugin downloads and extracts a plugin from GitHub
func (pm *PluginManager) downloadAndExtractPlugin(owner, repoName, pluginDir string) error {
	// Download latest release or main branch
	url := fmt.Sprintf("https://github.com/%s/%s/archive/refs/heads/main.tar.gz", owner, repoName)

	fmt.Printf("-----> Downloading from %s\n", url)

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download plugin: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download plugin: HTTP %d", resp.StatusCode)
	}

	// Extract tar.gz
	gzReader, err := gzip.NewReader(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %v", err)
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar: %v", err)
		}

		// Skip the root directory (repoName-main/)
		if header.Name == fmt.Sprintf("%s-main/", repoName) {
			continue
		}

		// Remove the root directory prefix
		targetPath := strings.TrimPrefix(header.Name, fmt.Sprintf("%s-main/", repoName))
		if targetPath == "" {
			continue
		}

		fullPath := filepath.Join(pluginDir, targetPath)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(fullPath, os.FileMode(header.Mode)); err != nil {
				return fmt.Errorf("failed to create directory: %v", err)
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
				return fmt.Errorf("failed to create parent directory: %v", err)
			}

			file, err := os.Create(fullPath)
			if err != nil {
				return fmt.Errorf("failed to create file: %v", err)
			}

			if _, err := io.Copy(file, tarReader); err != nil {
				file.Close()
				return fmt.Errorf("failed to copy file content: %v", err)
			}

			file.Close()

			// Set file permissions
			if err := os.Chmod(fullPath, os.FileMode(header.Mode)); err != nil {
				return fmt.Errorf("failed to set file permissions: %v", err)
			}
		}
	}

	return nil
}

// makeScriptsExecutable makes all plugin scripts executable
func (pm *PluginManager) makeScriptsExecutable(pluginDir string) error {
	// Make install script executable
	installPath := filepath.Join(pluginDir, "install")
	if _, err := os.Stat(installPath); err == nil {
		if err := os.Chmod(installPath, 0755); err != nil {
			return fmt.Errorf("failed to make install script executable: %v", err)
		}
	}

	// Make uninstall script executable
	uninstallPath := filepath.Join(pluginDir, "uninstall")
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
