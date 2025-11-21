package main

import (
	"fmt"
	"os"
	"strings"

	"gokku/internal"
	"gokku/internal/commands"
)

const version = "1.0.112"

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(0)
	}

	command := os.Args[1]
	args := os.Args[2:]

	// Create execution context for commands that need it
	var ctx *internal.ExecutionContext
	var err error

	// Commands that need context (use -a flag)
	contextCommands := map[string]bool{
		"config": true, "run": true, "logs": true,
		"status": true, "restart": true, "rollback": true,
		"ps": true,
	}

	// Check if command needs context (exact match or prefix match)
	needsContext := contextCommands[command] ||
		strings.HasPrefix(command, "config:") ||
		strings.HasPrefix(command, "ps:")

	if needsContext {
		// Extract app flag to create context
		appName, _ := internal.ExtractAppFlag(args)

		ctx, err = internal.NewExecutionContext(appName)

		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	}

	if strings.HasPrefix(command, "ps:") {
		subcommand := strings.TrimPrefix(command, "ps:")
		commands.Processes(append([]string{subcommand}, os.Args[2:]...))
		return
	}

	if strings.HasPrefix(command, "plugin:") {
		subcommand := strings.TrimPrefix(command, "plugin:")
		commands.Plugins(append([]string{subcommand}, os.Args[2:]...))
		return
	}

	if strings.HasPrefix(command, "plugins:") {
		subcommand := strings.TrimPrefix(command, "plugins:")
		commands.Plugins(append([]string{subcommand}, os.Args[2:]...))
		return
	}

	if strings.HasPrefix(command, "service:") {
		subcommand := strings.TrimPrefix(command, "service:")
		commands.Services(append([]string{subcommand}, os.Args[2:]...))
		return
	}

	if strings.HasPrefix(command, "services:") {
		subcommand := strings.TrimPrefix(command, "services:")
		commands.Services(append([]string{subcommand}, os.Args[2:]...))
		return
	}

	if strings.HasPrefix(command, "config:") {
		subcommand := strings.TrimPrefix(command, "config:")
		commands.ConfigWithContext(ctx, append([]string{subcommand}, os.Args[2:]...))
		return
	}

	if strings.HasPrefix(command, "apps:") {
		subcommand := strings.TrimPrefix(command, "apps:")
		commands.Apps(append([]string{subcommand}, os.Args[2:]...))
		return
	}

	if strings.HasPrefix(command, "remote:") {
		subcommand := strings.TrimPrefix(command, "remote:")
		commands.Remote(append([]string{subcommand}, os.Args[2:]...))
		return
	}

	// Execute command with panic recovery
	switch command {
	case "apps":
		commands.Apps(args)
	case "config":
		commands.ConfigWithContext(ctx, args)
	case "run":
		commands.RunWithContext(ctx, args)
	case "logs":
		commands.LogsWithContext(ctx, args)
	case "status":
		commands.StatusWithContext(ctx, args)
	case "restart":
		commands.RestartWithContext(ctx, args)
	case "deploy":
		commands.Deploy(args)
	case "rollback":
		commands.RollbackWithContext(ctx, args)
	case "remote":
		commands.Remote(args)
	case "tool":
		commands.Tool(os.Args[2:])
	case "plugins":
		commands.Plugins(os.Args[2:])
	case "services":
		commands.Services(os.Args[2:])
	case "ps":
		commands.Processes(os.Args[2:])
	case "au", "update", "auto-update":
		commands.AutoUpdate(os.Args[2:])
	case "uninstall":
		commands.Uninstall(os.Args[2:])
	case "version", "--version", "-v":
		fmt.Printf("gokku version %s\n", version)
	case "help", "--help", "-h":
		printHelp()
	default:
		// Check if --remote flag is present
		remoteInfo, remainingArgs, err := internal.GetRemoteInfoOrDefault(args)

		if err == nil && remoteInfo != nil {
			// If --remote is present, execute the command remotely
			// This handles cases like "gokku --remote nginx" or "gokku --remote nginx:reload nginx-lb"
			// Build command: "gokku" + command + remaining args (without --remote)
			remoteCmd := []string{"gokku", command}
			remoteCmd = append(remoteCmd, remainingArgs...)
			cmd := strings.Join(remoteCmd, " ")

			if err := internal.ExecuteRemoteCommand(remoteInfo, cmd); err != nil {
				os.Exit(1)
			}

			return
		}

		// Check if it's a plugin command (format: plugin:command)
		if strings.Contains(command, ":") {
			// Pass the full command to plugin handler
			commands.Plugins(os.Args[1:])
		} else {
			// Check if it's an installed plugin and show help
			if commands.IsPluginInstalled(command) {
				// Show plugin help by default
				commands.Plugins([]string{command + ":help"})
			} else {
				fmt.Printf("Unknown command: %s\n", command)
				fmt.Println("Run 'gokku --help' for usage")
				os.Exit(1)
			}
		}
	}
}

func printHelp() {
	fmt.Println(`gokku - Deployment management CLI

Usage:
  gokku <command> [options]

CLIENT COMMANDS (run from local machine):
  remote        Manage git remotes (add, list, remove, setup)
  apps          List applications on remote server
  config         Manage environment variables (use -a with git remote)
  run            Run arbitrary commands (use -a)
  logs           View application logs (use -a)
  status         Check services status (use -a)
  restart        Restart services (use -a)
  deploy         Deploy applications
  rollback       Rollback to previous release
  tool           Utility commands for scripts
  plugins        Manage plugins
  services       Manage services
  ps             Process management (list, restart, stop)
  uninstall      Remove Gokku installation
  version        Show version
  help           Show this help

SERVER COMMANDS (run directly on server):
  config         Manage environment variables locally (use -a with app name)
  run            Run arbitrary commands locally
  logs           View application logs locally
  status         Check services status locally
  restart        Restart services locally
  rollback       Rollback to previous release locally

Remote Management:
  gokku remote add <app_name> <user@host>                      Add a git remote
  gokku remote list                                             List git remotes
  gokku remote remove <name>                                    Remove a git remote
  gokku remote setup <user@host> [-i|--identity <pem_file>]      One-time server setup

Client Commands (always use -a with git remote):
  gokku config set KEY=VALUE -a <git-remote>
  gokku config get KEY -a <git-remote>
  gokku config list -a <git-remote>
  gokku config unset KEY -a <git-remote>

  gokku run <command> -a <git-remote>

  gokku logs -a <git-remote> [-f]
  gokku status -a <git-remote>
  gokku restart -a <git-remote>

  gokku deploy -a <git-remote>
  gokku rollback -a <git-remote>

  gokku ps:list -a <git-remote>
  gokku ps:restart -a <git-remote>
  gokku ps:stop -a <git-remote>

Server Commands (run on server only, use -a with app name):
  gokku run <command>                                (run locally)
  gokku logs <app> <env> [-f]                        (view logs locally)
  gokku status [app]                                 (check status locally)
  gokku restart <app>                                (restart locally)
  gokku rollback <app> <env> [release-id]            (rollback locally)

  Server Mode (-a with app name):
  -a, --app <app-name>`)
}
