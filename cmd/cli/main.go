package main

import (
	"fmt"
	"os"

	"infra/internal/handlers"
)

const version = "1.0.16"

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(0)
	}

	command := os.Args[1]

	switch command {
	case "apps":
		handlers.HandleApps(os.Args[2:])
	case "config":
		handlers.HandleConfig(os.Args[2:])
	case "run":
		handlers.HandleRun(os.Args[2:])
	case "logs":
		handlers.HandleLogs(os.Args[2:])
	case "status":
		handlers.HandleStatus(os.Args[2:])
	case "restart":
		handlers.HandleRestart(os.Args[2:])
	case "deploy":
		handlers.HandleDeploy(os.Args[2:])
	case "rollback":
		handlers.HandleRollback(os.Args[2:])
	case "ssh":
		handlers.HandleSSH(os.Args[2:])
	case "tool":
		handlers.HandleTool(os.Args[2:])
	case "version", "--version", "-v":
		fmt.Printf("gokku version %s\n", version)
	case "help", "--help", "-h":
		printHelp()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		fmt.Println("Run 'gokku --help' for usage")
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Println(`gokku - Deployment management CLI

Usage:
  gokku <command> [options]

CLIENT COMMANDS (run from local machine):
  server         Manage server connections
  apps           List applications on remote server
  config         Manage environment variables (use --remote)
  run            Run arbitrary commands (use --remote)
  logs           View application logs (use --remote)
  status         Check services status (use --remote)
  restart        Restart services (use --remote)
  deploy         Deploy applications
  rollback       Rollback to previous release
  ssh            SSH to server
  tool           Utility commands for scripts
  version        Show version
  help           Show this help

SERVER COMMANDS (run directly on server):
  config         Manage environment variables locally (--app required)
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

Client Commands (always use --remote):
  gokku config set KEY=VALUE --remote <git-remote>
  gokku config get KEY --remote <git-remote>
  gokku config list --remote <git-remote>
  gokku config unset KEY --remote <git-remote>

  gokku run <command> --remote <git-remote>

  gokku logs <app> <env> [-f] [--remote <git-remote>]
  gokku status [app] [env] [--remote <git-remote>]
  gokku restart --remote <git-remote>

  gokku deploy <app> <env> [--remote <git-remote>]
  gokku rollback <app> <env> [--remote <git-remote>]


Server Commands (run on server only):
  gokku config set KEY=VALUE --app <app> [--env <env>]
  gokku config set KEY=VALUE -a <app> [-e <env>]     (shorthand, env defaults to 'default')
  gokku config get KEY -a <app>                      (uses 'default' env)
  gokku config list -a <app> -e production           (explicit env)
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

  # Client usage - all commands use --remote
  gokku config set PORT=8080 --remote api-production
  gokku config list --remote api-production
  gokku logs api production -f --remote api-production
  gokku status --remote api-production
  gokku deploy api production --remote api-production

  # Server usage - run directly on server (no --remote needed)
  gokku config set PORT=8080 --app api
  gokku config list --app api --env production
  gokku logs api production -f
  gokku status
  gokku restart api

Remote Format:
  --remote <git-remote-name>

  The git remote name (e.g., "api-production", "vad-staging")
  Gokku will run 'git remote get-url <name>' to extract:
  - SSH host (user@ip or user@hostname)
  - App name from path

  Examples of git remotes:
  - api-production → ubuntu@server:api
  - worker-production    → ubuntu@server:/opt/gokku/repos/worker.git

  Environment is extracted from remote name suffix:
  - api-production → app: api, env: production
  - worker-production     → app: worker, env: production

Configuration:
  Config file: ~/.gokku/config.yml`)
}
