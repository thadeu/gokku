package main

import (
	"fmt"
	"os"
	"strings"

	"infra/internal"
	"infra/internal/handlers"
)

const version = "1.0.77"

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

	if contextCommands[command] {
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
		handlers.HandlePS(append([]string{subcommand}, os.Args[2:]...))
		return
	}

	if strings.HasPrefix(command, "plugins:") {
		subcommand := strings.TrimPrefix(command, "plugins:")
		handlers.HandlePlugins(append([]string{subcommand}, os.Args[2:]...))
		return
	}

	if strings.HasPrefix(command, "services:") {
		subcommand := strings.TrimPrefix(command, "services:")
		handlers.HandleServices(append([]string{subcommand}, os.Args[2:]...))
		return
	}

	if strings.HasPrefix(command, "config:") {
		subcommand := strings.TrimPrefix(command, "config:")
		handlers.HandleConfigWithContext(ctx, append([]string{subcommand}, os.Args[2:]...))
		return
	}

	// Execute command with panic recovery
	switch command {
	case "apps":
		handlers.HandleApps(args)
	case "config":
		handlers.HandleConfigWithContext(ctx, args)
	case "run":
		handlers.HandleRunWithContext(ctx, args)
	case "logs":
		handlers.HandleLogsWithContext(ctx, args)
	case "status":
		handlers.HandleStatusWithContext(ctx, args)
	case "restart":
		handlers.HandleRestartWithContext(ctx, args)
	case "deploy":
		handlers.HandleDeploy(args)
	case "rollback":
		handlers.HandleRollbackWithContext(ctx, args)
	case "ssh":
		handlers.HandleSSH(args)
	case "server":
		handlers.HandleServer(args)
	case "tool":
		handlers.HandleTool(os.Args[2:])
	case "plugins":
		handlers.HandlePlugins(os.Args[2:])
	case "services":
		handlers.HandleServices(os.Args[2:])
	case "ps":
		handlers.HandlePS(os.Args[2:])
	case "au", "update", "auto-update":
		handlers.HandleAutoUpdate(os.Args[2:])
	case "autocomplete":
		handlers.HandleAutocomplete(os.Args[2:])
	case "version", "--version", "-v":
		fmt.Printf("gokku version %s\n", version)
	case "help", "--help", "-h":
		printHelp()
	default:
		// Check if it's a plugin command (format: plugin:command)
		if strings.Contains(command, ":") {
			// Pass the full command to plugin handler
			handlers.HandlePlugins(os.Args[1:])
		} else {
			// Check if it's an installed plugin and show help
			if handlers.IsPluginInstalled(command) {
				// Show plugin help by default
				handlers.HandlePlugins([]string{command + ":help"})
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
  server         Manage server connections
  apps           List applications on remote server
  config         Manage environment variables (use -a with git remote)
  run            Run arbitrary commands (use -a)
  logs           View application logs (use -a)
  status         Check services status (use -a)
  restart        Restart services (use -a)
  deploy         Deploy applications
  rollback       Rollback to previous release
  ssh            SSH to server
  tool           Utility commands for scripts
  plugins        Manage plugins
  services       Manage services
  ps             Process management (scale, list, restart, stop)
  autocomplete  Install shell completion (bash, zsh, fish)
  version        Show version
  help           Show this help

SERVER COMMANDS (run directly on server):
  config         Manage environment variables locally (use -a with app name)
  run            Run arbitrary commands locally
  logs           View application logs locally
  status         Check services status locally
  restart        Restart services locally
  rollback       Rollback to previous release locally

Server Management:
  gokku server add <name> <host>           Add a server
  gokku server list                        List servers
  gokku server remove <name>               Remove a server
  gokku server set-default <name>          Set default server

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

  gokku ps:scale web=4 worker=2 -a <git-remote>
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
