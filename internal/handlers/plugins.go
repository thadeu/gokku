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
	case "list-official":
		handlePluginsListOfficial()
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

// handlePluginsListOfficial lists available official plugins
func handlePluginsListOfficial() {
	contribDir := filepath.Join("contrib", "plugins")

	if _, err := os.Stat(contribDir); os.IsNotExist(err) {
		fmt.Println("No official plugins available")
		return
	}

	entries, err := os.ReadDir(contribDir)
	if err != nil {
		fmt.Printf("Error reading contrib/plugins directory: %v\n", err)
		return
	}

	if len(entries) == 0 {
		fmt.Println("No official plugins available")
		return
	}

	fmt.Println("Available official plugins:")
	fmt.Println("")

	for _, entry := range entries {
		if entry.IsDir() {
			fmt.Printf("  %s\n", entry.Name())
		}
	}

	fmt.Println("")
	fmt.Println("Install with: gokku plugins:add <plugin-name>")
	fmt.Println("Example: gokku plugins:add nginx")
}

// handlePluginsAdd adds a new plugin from local contrib or Git repository
func handlePluginsAdd(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: gokku plugins:add <plugin-name>|<git-url>")
		fmt.Println("")
		fmt.Println("Examples:")
		fmt.Println("  gokku plugins:add nginx                              # Official plugin from contrib/plugins")
		fmt.Println("  gokku plugins:add https://github.com/thadeu/gokku-nginx  # Plugin from GitHub")
		fmt.Println("  gokku plugins:add https://gitlab.com/user/gokku-redis   # Plugin from GitLab")
		os.Exit(1)
	}

	pluginArg := args[0]

	pm := plugins.NewPluginManager()

	// Check if it's a URL
	if pm.IsValidGitURL(pluginArg) {
		// Extract plugin name from URL
		pluginName := pm.ExtractPluginNameFromURL(pluginArg)

		fmt.Printf("Adding plugin '%s' from %s...\n", pluginName, pluginArg)

		if err := pm.InstallPluginFromGit(pluginArg, pluginName); err != nil {
			fmt.Printf("Error installing plugin from Git: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Plugin '%s' installed successfully\n", pluginName)
		fmt.Printf("Plugin is now available. Create a service with:\n")
		fmt.Printf("  gokku services:create %s --name <service-name>\n", pluginName)
		return
	}

	// If not a URL, try to find in local contrib/plugins directory
	localPluginPath := filepath.Join("contrib", "plugins", pluginArg)
	if _, err := os.Stat(localPluginPath); err == nil {
		fmt.Printf("Adding official plugin '%s' from contrib/plugins...\n", pluginArg)

		if err := pm.InstallLocalPlugin(pluginArg, localPluginPath); err != nil {
			fmt.Printf("Error installing local plugin: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Plugin '%s' installed successfully\n", pluginArg)
		fmt.Printf("Plugin is now available. Create a service with:\n")
		fmt.Printf("  gokku services:create %s --name <service-name>\n", pluginArg)
		return
	}

	// If not found locally and not a URL, show error
	fmt.Printf("Plugin '%s' not found in contrib/plugins and is not a valid Git URL\n", pluginArg)
	fmt.Println("")
	fmt.Println("Available options:")
	fmt.Println("  - Use official plugin name (e.g., 'nginx')")
	fmt.Println("  - Use Git URL (e.g., 'https://github.com/user/gokku-plugin')")
	fmt.Println("  - List official plugins: gokku plugins:list-official")
	os.Exit(1)
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
	fmt.Println("  gokku plugins:list-official           List available official plugins")
	fmt.Println("  gokku plugins:add <plugin-name>       Add official plugin")
	fmt.Println("  gokku plugins:add <git-url>           Add plugin from Git repository")
	fmt.Println("  gokku plugins:remove <plugin>         Remove plugin")
	fmt.Println("")
	fmt.Println("Plugin commands:")
	fmt.Println("  gokku <plugin>:<command> <service>   Execute plugin command")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  gokku plugins:add nginx                              # Official plugin")
	fmt.Println("  gokku plugins:add https://github.com/user/gokku-nginx  # GitHub plugin")
	fmt.Println("  gokku plugins:add https://gitlab.com/user/gokku-redis  # GitLab plugin")
	fmt.Println("  gokku postgres:info postgres-api")
	fmt.Println("  gokku postgres:export postgres-api > backup.sql")
}
