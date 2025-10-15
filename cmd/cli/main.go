package main

import (
	"fmt"
	"os"

	"infra/internal/handlers"
)

const version = "1.0.0"

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(0)
	}

	command := os.Args[1]

	switch command {
	case "server":
		handlers.HandleServer(os.Args[2:])
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

Commands:
  server         Manage servers
  apps           List applications
  config         Manage environment variables
  run            Run arbitrary command
  logs           View application logs
  status         Check services status
  restart        Restart service
  deploy         Deploy applications
  rollback       Rollback to previous release
  ssh            SSH to server
  version        Show version
  help           Show this help

Server Management:
  gokku server add <name> <host>           Add a server
  gokku server list                        List servers
  gokku server remove <name>               Remove a server
  gokku server set-default <name>          Set default server

Configuration Management:
  # Remote execution (from local machine or CI/CD)
  gokku config set KEY=VALUE --remote <git-remote>
  gokku config get KEY --remote <git-remote>
  gokku config list --remote <git-remote>
  gokku config unset KEY --remote <git-remote>

  # Local execution (on server)
  gokku config set KEY=VALUE --app <app> [--env <env>]
  gokku config set KEY=VALUE -a <app> [-e <env>]     (shorthand, env defaults to 'default')
  gokku config get KEY -a <app>                      (uses 'default' env)
  gokku config list -a <app> -e production           (explicit env)
  gokku config unset KEY -a <app>

Run Commands:
  gokku run <command> --remote <git-remote>          Run on remote server
  gokku run <command> -a <app> -e <env>              Run locally

Logs & Status:
  gokku logs <app> <env> [-f] [--remote <git-remote>]
  gokku status [app] [env] [--remote <git-remote>]
  gokku restart <app> <env> [--remote <git-remote>]

Deployment:
  gokku deploy <app> <env> [--remote <git-remote>]
  gokku rollback <app> <env> [--remote <git-remote>]

Examples:
  # Setup server
  gokku server add prod ubuntu@ec2.compute.amazonaws.com

  # Setup git remote (standard git)
  git remote add api-production ubuntu@server:/opt/gokku/repos/api.git
  git remote add vad-staging ubuntu@server:/opt/gokku/repos/vad.git

  # Configuration
  gokku config set PORT=8080 --remote api-production
  gokku config set DATABASE_URL=postgres://... --remote api-production
  gokku config list --remote api-production
  gokku config get PORT --remote vad-staging

  # Run commands
  gokku run "systemctl status api-production" --remote api-production
  gokku run "docker ps" --remote vad-production
  gokku run "bundle exec bin/console" --remote app-production

  # Logs and status
  gokku logs api production -f
  gokku logs --remote api-production -f
  gokku status --remote api-production
  gokku restart --remote vad-staging

  # Deploy
  gokku deploy api production
  gokku deploy --remote api-production

Remote Format:
  --remote <git-remote-name>

  The git remote name (e.g., "api-production", "vad-staging")
  Gokku will run 'git remote get-url <name>' to extract:
  - SSH host (user@ip or user@hostname)
  - App name from path (/opt/gokku/repos/<app>.git)

  Examples of git remotes:
  - api-production → ubuntu@server:/opt/gokku/repos/api.git
  - vad-staging    → ubuntu@server:/opt/gokku/repos/vad.git

  Environment is extracted from remote name suffix:
  - api-production → app: api, env: production
  - vad-staging    → app: vad, env: staging
  - worker-dev     → app: worker, env: dev

Configuration:
  Config file: ~/.gokku/config.yml`)
}
