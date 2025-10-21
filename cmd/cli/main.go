package main

import (
	"fmt"
	"os"

	"infra/internal"
	"infra/internal/handlers"
)

const version = "1.0.38"

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

	// Execute command with panic recovery
	internal.TryCatch(func() {
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
			handlers.HandleTool(args)
		case "version", "--version", "-v":
			fmt.Printf("gokku version %s\n", version)
		case "help", "--help", "-h":
			printHelp()
		default:
			fmt.Printf("Unknown command: %s\n", command)
			fmt.Println("Run 'gokku --help' for usage")
			os.Exit(1)
		}
	})
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


Server Commands (run on server only, use -a with app name):
  gokku config set KEY=VALUE -a <app>
  gokku config get KEY -a <app>
  gokku config list -a <app>
  gokku config unset KEY -a <app>

  gokku run <command>                                (run locally)
  gokku logs <app> <env> [-f]                        (view logs locally)
  gokku status [app] [env]                           (check status locally)
  gokku restart <app>                                (restart locally)
  gokku rollback <app> <env> [release-id]            (rollback locally)


Examples:
  # Setup server connection (client)
  gokku server add prod ubuntu@ec2.compute.amazonaws.com

  # Setup git remote (standard git)
  git remote add api-production ubuntu@server:api

  # Client usage - all commands use -a with git remote
  gokku config set PORT=8080 -a api-production
  gokku config list -a api-production
  gokku logs -a api-production -f
  gokku status -a api-production
  gokku deploy -a api-production

  # Server usage - run directly on server with app name
  gokku config set PORT=8080 -a api
  gokku config list -a api
  gokku logs api production -f
  gokku status
  gokku restart api

App Format:
  Client Mode (-a with git remote):
  -a, --app <git-remote-name>

  The git remote name (e.g., "api-production", "worker-staging")
  Gokku will run 'git remote get-url <name>' to extract:
  - SSH host (user@ip or user@hostname)
  - App name from path

  Examples of git remotes:
  - api-production → ubuntu@server:api
  - worker-production    → ubuntu@server:/opt/gokku/repos/worker.git

  Server Mode (-a with app name):
  -a, --app <app-name>

  The app name directly (e.g., "api", "worker")
  No git remote needed - uses app name directly

Configuration:
  Config file: ~/.gokku/config.yml`)
}
