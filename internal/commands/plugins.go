package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gokku/internal"
	"gokku/internal/plugins"
	"gokku/tui"
)

func usePlugins(args []string) {
	if len(args) == 0 {
		showHelp()
		return
	}

	// Extract --remote flag first (if present)
	remoteInfo, remainingArgs, err := internal.GetRemoteInfoOrDefault(args)

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if len(remainingArgs) == 0 {
		// Only --remote provided, show help or list plugins
		if remoteInfo != nil {
			cmd := "gokku plugins list"

			if err := internal.ExecuteRemoteCommand(remoteInfo, cmd); err != nil {
				os.Exit(1)
			}

			return
		}

		showHelp()
		return
	}

	// Handle both "plugins add" and "plugins:add" formats
	subcommand := remainingArgs[0]

	// Check if it's in the format "plugins:command"
	if strings.Contains(subcommand, ":") {
		parts := strings.Split(subcommand, ":")
		if len(parts) == 2 && parts[0] == "plugins" {
			subcommand = parts[1]
		}
	}

	// Check if subcommand is a flag (like --remote that wasn't caught)
	if strings.HasPrefix(subcommand, "--") && subcommand != "--help" && subcommand != "--remote" {
		fmt.Printf("Unknown plugin command: %s\n", subcommand)
		showHelp()
		os.Exit(1)
	}

	switch subcommand {
	case "list", "ls":
		list(remainingArgs[1:], remoteInfo)
	case "add", "install":
		add(remainingArgs[1:], remoteInfo)
	case "update":
		update(remainingArgs[1:], remoteInfo)
	case "remove":
		remove(remainingArgs[1:], remoteInfo)
	default:
		if remoteInfo != nil {
			// Execute plugin command remotely
			cmd := fmt.Sprintf("gokku plugins %s", strings.Join(remainingArgs, " "))

			if err := internal.ExecuteRemoteCommand(remoteInfo, cmd); err != nil {
				os.Exit(1)
			}

			return
		}

		wildcard(remainingArgs)
	}
}

// lists all installed plugins
func list(args []string, remoteInfo *internal.RemoteInfo) {
	if remoteInfo != nil {
		// Client mode: execute remotely
		cmd := "gokku plugins list"
		if err := internal.ExecuteRemoteCommand(remoteInfo, cmd); err != nil {
			fmt.Printf("Error listing plugins: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Server mode: execute locally
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

	table := tui.NewTable(tui.ASCII)
	table.AppendHeaders([]string{"NAME"})
	table.AppendSeparator()

	for _, plugin := range pluginList {
		table.AppendRow([]string{plugin}, true)
	}

	fmt.Print(table.Render())
	_ = args // unused for now
}

// adds a new plugin from official repository or Git URL
func add(args []string, remoteInfo *internal.RemoteInfo) {
	// Use args directly if remoteInfo is already set (from parent)
	// The --remote flag was already extracted by handlePlugins
	cleanArgs := args

	if len(cleanArgs) < 1 {
		fmt.Println("Usage: gokku plugins:add <plugin-name> [<git-url>] [--remote]")
		fmt.Println("")
		fmt.Println("Examples:")
		fmt.Println("  gokku plugins:add nginx                              # Official plugin")
		fmt.Println("  gokku plugins:add myplugin https://github.com/user/gokku-myplugin  # Community plugin")
		fmt.Println("  gokku plugins:add nginx --remote                    # Install on remote server")
		fmt.Println("")
		fmt.Println("Official plugins are automatically fetched from gokku-vm organization")
		fmt.Println("Community plugins require a git URL")
		os.Exit(1)
	}

	pluginName := cleanArgs[0]
	var gitURL string
	if len(cleanArgs) > 1 {
		// Check if next arg is a flag, not a URL
		nextArg := cleanArgs[1]
		if !strings.HasPrefix(nextArg, "-") {
			gitURL = nextArg
		}
	}

	// If remote mode, execute remotely
	if remoteInfo != nil {
		cmdParts := []string{"gokku plugins:add", pluginName}
		if gitURL != "" {
			cmdParts = append(cmdParts, gitURL)
		}
		cmd := strings.Join(cmdParts, " ")
		if err := internal.ExecuteRemoteCommand(remoteInfo, cmd); err != nil {
			fmt.Printf("Error installing plugin: %v\n", err)
			os.Exit(1)
		}
		return
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

// updates a plugin from its source repository
func update(args []string, remoteInfo *internal.RemoteInfo) {
	// Use args directly if remoteInfo is already set (from parent)
	cleanArgs := args
	if remoteInfo == nil {
		// Extract --remote flag if present
		_, extractedArgs, err := internal.GetRemoteInfoOrDefault(args)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		cleanArgs = extractedArgs
	}

	if len(cleanArgs) < 1 {
		fmt.Println("Usage: gokku plugins:update <plugin-name> [--remote]")
		fmt.Println("")
		fmt.Println("Examples:")
		fmt.Println("  gokku plugins:update redis")
		fmt.Println("  gokku plugins:update redis --remote")
		os.Exit(1)
	}

	pluginName := cleanArgs[0]

	// If remote mode, execute remotely
	if remoteInfo != nil {
		cmd := fmt.Sprintf("gokku plugins:update %s", pluginName)
		if err := internal.ExecuteRemoteCommand(remoteInfo, cmd); err != nil {
			fmt.Printf("Error updating plugin: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Server mode: execute locally
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

// removes a plugin
func remove(args []string, remoteInfo *internal.RemoteInfo) {
	// Use args directly if remoteInfo is already set (from parent)
	cleanArgs := args
	if remoteInfo == nil {
		// Extract --remote flag if present
		_, extractedArgs, err := internal.GetRemoteInfoOrDefault(args)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		cleanArgs = extractedArgs
	}

	if len(cleanArgs) < 1 {
		fmt.Println("Usage: gokku plugins:remove <plugin-name> [--remote]")
		fmt.Println("")
		fmt.Println("Examples:")
		fmt.Println("  gokku plugins:remove nginx")
		fmt.Println("  gokku plugins:remove nginx --remote")
		os.Exit(1)
	}

	pluginName := cleanArgs[0]

	// If remote mode, execute remotely
	if remoteInfo != nil {
		cmd := fmt.Sprintf("gokku plugins:remove %s", pluginName)
		if err := internal.ExecuteRemoteCommand(remoteInfo, cmd); err != nil {
			fmt.Printf("Error removing plugin: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Server mode: execute locally
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

// runs a dynamic plugin command and subcommand
func wildcard(args []string) {
	// Parse: gokku postgres:export postgres-api
	parts := strings.Split(args[0], ":")

	if len(parts) != 2 {
		fmt.Printf("Unknown command: %s\n", args[0])
		showHelp()

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

	internal.TryCatch(func() {
		// Get the plugin directory from PluginManager
		pluginDir := filepath.Join(pm.GetPluginsDir(), pluginName)

		var commandPath string

		if pm.BinExists(pluginName, command) {
			commandPath = filepath.Join(pluginDir, "bin", command)
		} else {
			commandPath = filepath.Join(pluginDir, "commands", command)
		}

		if commandPath == "" {
			fmt.Printf("Command '%s' not found for plugin '%s'\n", command, pluginName)
			os.Exit(1)
		}

		// Build command arguments (pass all remaining args to the plugin command)
		cmdArgs := []string{"-c", commandPath}

		if len(args) > 1 {
			cmdArgs = append(cmdArgs, args[1:]...)
		}

		// Execute the plugin command
		cmd := exec.Command("bash", cmdArgs[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin

		if err := cmd.Run(); err != nil {
			// Don't print error if command already printed its own error
			os.Exit(1)
		}
	})
}

func IsPluginInstalled(pluginName string) bool {
	pm := plugins.NewPluginManager()
	return pm.PluginExists(pluginName)
}

// showPluginHelp shows plugin help
func showHelp() {
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
