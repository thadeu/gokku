package handlers

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"infra/internal/plugins"
)

// IsPluginInstalled checks if a plugin is installed
func IsPluginInstalled(pluginName string) bool {
	pm := plugins.NewPluginManager()
	return pm.PluginExists(pluginName)
}

// handlePlugins manages plugin-related commands
func handlePlugins(args []string) {
	if len(args) == 0 {
		showPluginHelp()
		return
	}

	// Handle both "plugins add" and "plugins:add" formats
	subcommand := args[0]

	// Check if it's in the format "plugins:command"
	if strings.Contains(subcommand, ":") {
		parts := strings.Split(subcommand, ":")
		if len(parts) == 2 && parts[0] == "plugins" {
			subcommand = parts[1]
		}
	}

	switch subcommand {
	case "list", "ls":
		handlePluginsList()
	case "add":
		handlePluginsAdd(args[1:])
	case "update":
		handlePluginsUpdate(args[1:])
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

// handlePluginsAdd adds a new plugin from official repository or Git URL
func handlePluginsAdd(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: gokku plugins:add <plugin-name> [<git-url>]")
		fmt.Println("")
		fmt.Println("Examples:")
		fmt.Println("  gokku plugins:add nginx                              # Official plugin")
		fmt.Println("  gokku plugins:add myplugin https://github.com/user/gokku-myplugin  # Community plugin")
		fmt.Println("")
		fmt.Println("Official plugins are automatically fetched from gokku-vm organization")
		fmt.Println("Community plugins require a git URL")
		os.Exit(1)
	}

	pluginName := args[0]
	var gitURL string
	if len(args) > 1 {
		gitURL = args[1]
	}

	pm := plugins.NewPluginManager()

	// Check if it's a community plugin (git URL provided)
	if gitURL != "" {
		if !pm.IsValidGitURL(gitURL) {
			fmt.Printf("Invalid Git URL: %s\n", gitURL)
			os.Exit(1)
		}

		fmt.Printf("Installing community plugin '%s' from %s...\n", pluginName, gitURL)

		if err := pm.InstallPluginFromGit(gitURL, pluginName); err != nil {
			fmt.Printf("Error installing community plugin: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Plugin '%s' installed successfully\n", pluginName)
		fmt.Printf("Plugin is now available. Create a service with:\n")
		fmt.Printf("  gokku services:create %s --name <service-name>\n", pluginName)
		return
	}

	// Check if it's an official plugin
	officialPlugins := []string{"nginx", "postgres", "redis", "letsencrypt", "cron"}
	isOfficial := false
	for _, official := range officialPlugins {
		if pluginName == official {
			isOfficial = true
			break
		}
	}

	if isOfficial {
		fmt.Printf("Installing official plugin '%s'...\n", pluginName)

		if err := pm.InstallOfficialPlugin(pluginName); err != nil {
			fmt.Printf("Error installing official plugin: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Plugin '%s' installed successfully\n", pluginName)
		fmt.Printf("Plugin is now available. Create a service with:\n")
		fmt.Printf("  gokku services:create %s --name <service-name>\n", pluginName)
		return
	}

	// If not official and no git URL provided, show error
	fmt.Printf("Plugin '%s' is not an official plugin and no Git URL provided\n", pluginName)
	fmt.Println("")
	fmt.Println("Available options:")
	fmt.Println("  - Use official plugin name (e.g., 'nginx')")
	fmt.Println("  - Provide Git URL for community plugins (e.g., 'gokku plugins:add myplugin https://github.com/user/gokku-myplugin')")
	fmt.Println("  - List official plugins: gokku plugins:list-official")
	os.Exit(1)
}

// handlePluginsUpdate updates a plugin from its source repository
func handlePluginsUpdate(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: gokku plugins:update <plugin-name>")
		fmt.Println("")
		fmt.Println("Examples:")
		fmt.Println("  gokku plugins:update redis")
		fmt.Println("  gokku plugins:update nginx")
		os.Exit(1)
	}

	pluginName := args[0]

	pm := plugins.NewPluginManager()

	if !pm.PluginExists(pluginName) {
		fmt.Printf("Plugin '%s' not found\n", pluginName)
		fmt.Printf("Install it with: gokku plugins:add %s\n", pluginName)
		os.Exit(1)
	}

	fmt.Printf("Updating plugin '%s'...\n", pluginName)

	if err := pm.UpdatePlugin(pluginName); err != nil {
		fmt.Printf("Error updating plugin: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Plugin '%s' updated successfully\n", pluginName)
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

	pm := plugins.NewPluginManager()

	// Check if plugin exists
	if !pm.PluginExists(pluginName) {
		fmt.Printf("Plugin '%s' not found\n", pluginName)
		fmt.Printf("Install it with: gokku plugins:add %s\n", pluginName)
		os.Exit(1)
	}

	// Check if command exists
	if !pm.CommandExists(pluginName, command) {
		fmt.Printf("Command '%s' not found for plugin '%s'\n", command, pluginName)
		os.Exit(1)
	}

	// Get the plugin directory from PluginManager
	pluginDir := filepath.Join(pm.GetPluginsDir(), pluginName)
	commandPath := filepath.Join(pluginDir, "commands", command)

	// Build command arguments (pass all remaining args to the plugin command)
	cmdArgs := []string{"bash", commandPath}
	if len(args) > 1 {
		cmdArgs = append(cmdArgs, args[1:]...)
	}

	// Execute the plugin command
	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		// Don't print error if command already printed its own error
		os.Exit(1)
	}
}

// showPluginHelp shows plugin help
func showPluginHelp() {
	fmt.Println("Plugin management commands:")
	fmt.Println("")
	fmt.Println("  gokku plugins:list                    List all installed plugins")
	fmt.Println("  gokku plugins:add <name>              Add official plugin")
	fmt.Println("  gokku plugins:add <name> <git-url>    Add community plugin")
	fmt.Println("  gokku plugins:update <plugin>         Update plugin from source")
	fmt.Println("  gokku plugins:remove <plugin>         Remove plugin")
	fmt.Println("")
	fmt.Println("Plugin commands:")
	fmt.Println("  gokku <plugin>:<command> <service>   Execute plugin command")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  gokku plugins:add nginx                              # Official plugin")
	fmt.Println("  gokku plugins:add aws https://github.com/thadeu/gokku-aws  # Community plugin")
	fmt.Println("  gokku plugins:update redis                           # Update plugin")
}
