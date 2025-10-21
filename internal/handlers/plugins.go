package handlers

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"infra/internal/plugins"
)

// handlePlugins manages plugin-related commands
func handlePlugins(args []string) {
	if len(args) == 0 {
		showPluginHelp()
		return
	}

	subcommand := args[0]

	switch subcommand {
	case "list", "ls":
		handlePluginsList()
	case "add":
		handlePluginsAdd(args[1:])
	case "remove":
		handlePluginsRemove(args[1:])
	default:
		// Try to execute as plugin command
		handlePluginCommand(args)
	}
}

// handlePluginsList lists all installed plugins
func handlePluginsList() {
	pm := plugins.NewPluginManager()

	pluginList, err := pm.ListPlugins()
	if err != nil {
		fmt.Printf("Error listing plugins: %v\n", err)
		os.Exit(1)
	}

	if len(pluginList) == 0 {
		fmt.Println("No plugins installed")
		return
	}

	fmt.Println("Installed plugins:")
	for _, plugin := range pluginList {
		fmt.Printf("  %s\n", plugin)
	}
}

// handlePluginsAdd adds a new plugin from GitHub
func handlePluginsAdd(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: gokku plugins:add <owner/repo>")
		fmt.Println("")
		fmt.Println("Examples:")
		fmt.Println("  gokku plugins:add thadeu/gokku-postgres")
		fmt.Println("  gokku plugins:add thadeu/gokku-redis")
		os.Exit(1)
	}

	repo := args[0]

	// Parse owner/repo
	parts := strings.Split(repo, "/")
	if len(parts) != 2 {
		fmt.Printf("Invalid repository format: %s\n", repo)
		fmt.Println("Expected format: owner/repo")
		os.Exit(1)
	}

	owner := parts[0]
	repoName := parts[1]

	// Extract plugin name from repo (remove gokku- prefix if present)
	pluginName := strings.TrimPrefix(repoName, "gokku-")

	fmt.Printf("Adding plugin '%s' from %s/%s...\n", pluginName, owner, repoName)

	pm := plugins.NewPluginManager()

	// Download plugin from GitHub
	if err := pm.DownloadPlugin(owner, repoName, pluginName); err != nil {
		fmt.Printf("Error downloading plugin: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Plugin '%s' downloaded successfully\n", pluginName)
	fmt.Printf("Plugin is now available. Create a service with:\n")
	fmt.Printf("  gokku services:create %s --name <service-name>\n", pluginName)
}

// handlePluginsRemove removes a plugin
func handlePluginsRemove(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: gokku plugins:remove <plugin>")
		os.Exit(1)
	}

	pluginName := args[0]

	pm := plugins.NewPluginManager()

	if !pm.PluginExists(pluginName) {
		fmt.Printf("Plugin '%s' not found\n", pluginName)
		os.Exit(1)
	}

	fmt.Printf("Removing plugin '%s'...\n", pluginName)

	if err := pm.RemovePlugin(pluginName); err != nil {
		fmt.Printf("Error removing plugin: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Plugin '%s' removed successfully\n", pluginName)
}

// handlePluginCommand executes a plugin command
func handlePluginCommand(args []string) {
	// Parse: gokku postgres:export postgres-api
	parts := strings.Split(args[0], ":")
	if len(parts) != 2 {
		fmt.Printf("Unknown command: %s\n", args[0])
		showPluginHelp()
		os.Exit(1)
	}

	pluginName := parts[0]
	command := parts[1]

	// Get service name from args
	if len(args) < 2 {
		fmt.Printf("Usage: gokku %s:%s <service>\n", pluginName, command)
		os.Exit(1)
	}

	serviceName := args[1]

	// Execute plugin command
	commandPath := filepath.Join("/opt/gokku/plugins", pluginName, "commands", command)

	if _, err := os.Stat(commandPath); err != nil {
		fmt.Printf("Command '%s' not found for plugin '%s'\n", command, pluginName)
		os.Exit(1)
	}

	// Execute the plugin command
	cmd := exec.Command("bash", commandPath, serviceName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("Error executing command: %v\n", err)
		os.Exit(1)
	}
}

// showPluginHelp shows plugin help
func showPluginHelp() {
	fmt.Println("Plugin management commands:")
	fmt.Println("")
	fmt.Println("  gokku plugins:list                    List all installed plugins")
	fmt.Println("  gokku plugins:add <owner/repo>       Add plugin from GitHub")
	fmt.Println("  gokku plugins:remove <plugin>         Remove plugin")
	fmt.Println("")
	fmt.Println("Plugin commands:")
	fmt.Println("  gokku <plugin>:<command> <service>   Execute plugin command")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  gokku plugins:add thadeu/gokku-postgres")
	fmt.Println("  gokku postgres:info postgres-api")
	fmt.Println("  gokku postgres:export postgres-api > backup.sql")
}
