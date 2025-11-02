package handlers

import (
	"fmt"
	"os"
	"strings"

	"gokku/internal"
	"gokku/internal/services"
)

// handleRemote manages git remote commands
func handleRemote(args []string) {
	if len(args) < 1 {
		printRemoteHelp()
		os.Exit(1)
	}

	subcommand := args[0]

	if strings.Contains(subcommand, ":") {
		parts := strings.Split(subcommand, ":")

		if len(parts) == 2 && parts[0] == "remote" {
			subcommand = parts[1]
		}
	}

	switch subcommand {
	case "add":
		handleRemoteAdd(args[1:])
	case "list", "ls":
		handleRemoteList()
	case "remove", "rm":
		handleRemoteRemove(args[1:])
	case "setup":
		handleRemoteSetup(args[1:])
	case "help", "--help", "-h":
		printRemoteHelp()
	default:
		fmt.Printf("Unknown remote command: %s\n", subcommand)
		printRemoteHelp()
		os.Exit(1)
	}
}

// handleRemoteAdd adds a new git remote
func handleRemoteAdd(args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: gokku remote add <app_name> <user@server_ip>")
		fmt.Println("Example: gokku remote add api user@hostname")
		os.Exit(1)
	}

	appName := args[0]
	serverHost := args[1]

	// Validate app name
	if appName == "" {
		fmt.Println("Error: app name cannot be empty")
		os.Exit(1)
	}

	// Validate server host format
	if !strings.Contains(serverHost, "@") {
		fmt.Println("Error: server host must be in format user@host")
		fmt.Println("Example: user@hostname")
		os.Exit(1)
	}

	// Create git remote URL
	remoteURL := fmt.Sprintf("%s:/opt/gokku/repos/%s.git", serverHost, appName)

	// Check if remote already exists
	checkCmd := fmt.Sprintf("git remote get-url %s 2>/dev/null", appName)
	output, _ := internal.Bash(checkCmd)
	if output != "" {
		fmt.Printf("Error: remote '%s' already exists -> %s\n", appName, strings.TrimSpace(output))
		fmt.Printf("Remove it first with: gokku remote remove %s\n", appName)
		os.Exit(1)
	}

	// Add git remote
	_, err := internal.Bash(fmt.Sprintf("git remote add %s %s", appName, remoteURL))

	if err != nil {
		fmt.Printf("Error adding remote '%s' -> %s\n", appName, remoteURL)
		os.Exit(1)
	}

	fmt.Printf("Added remote '%s' -> %s\n", appName, remoteURL)
}

// handleRemoteList lists all configured git remotes
func handleRemoteList() {
	output, err := internal.Bash("git remote -v")

	if err != nil {
		fmt.Println("Error listing remotes")
		os.Exit(1)
	}

	if strings.TrimSpace(output) == "" {
		fmt.Println("No remotes configured")
		fmt.Println("Add one with: gokku remote add <app_name> <user@server_ip>")
		return
	}

	fmt.Println("Configured remotes:")
	fmt.Println(output)
}

// handleRemoteRemove removes a git remote
func handleRemoteRemove(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: gokku remote remove <remote_name>")
		fmt.Println("Example: gokku remote remove api")
		os.Exit(1)
	}

	remoteName := args[0]

	// Check if remote exists
	checkCmd := fmt.Sprintf("git remote get-url %s 2>/dev/null", remoteName)
	output, _ := internal.Bash(checkCmd)
	if output == "" {
		fmt.Printf("Error: remote '%s' not found\n", remoteName)
		os.Exit(1)
	}

	_, err := internal.Bash(fmt.Sprintf("git remote remove %s", remoteName))

	if err != nil {
		fmt.Printf("Error removing remote '%s'\n", remoteName)
		os.Exit(1)
	}

	fmt.Printf("Removed remote '%s'\n", remoteName)
}

// handleRemoteSetup performs one-time server setup
func handleRemoteSetup(args []string) {
	// Extract identity flag (-i or --identity)
	identityFile, remainingArgs := internal.ExtractIdentityFlag(args)

	// Get server host from remaining args
	if len(remainingArgs) < 1 {
		fmt.Println("Usage: gokku remote setup <user@host> [-i|--identity <pem_file>]")
		fmt.Println("Example: gokku remote setup ubuntu@192.168.105.3")
		fmt.Println("Example: gokku remote setup ubuntu@ec2.example.com -i ~/.ssh/my-key.pem")
		os.Exit(1)
	}

	serverHost := remainingArgs[0]

	// Validate server host format
	if !strings.Contains(serverHost, "@") {
		fmt.Println("Error: server host must be in format user@host")
		fmt.Println("Example: ubuntu@192.168.105.3")
		os.Exit(1)
	}

	// Validate identity file if provided
	if identityFile != "" {
		if _, err := os.Stat(identityFile); os.IsNotExist(err) {
			fmt.Printf("Error: identity file not found: %s\n", identityFile)
			os.Exit(1)
		}
	}

	// Create setup service and execute
	setup := services.NewServerSetup(serverHost, identityFile)
	if err := setup.Execute(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func printRemoteHelp() {
	fmt.Println(`Remote Management Commands:

Usage:
  gokku remote <command> [options]

Commands:
  add <app_name> <user@server_ip>    Add a new git remote
  list, ls                           List all configured remotes
  remove, rm <remote_name>           Remove a git remote
  setup <user@host> [-i|--identity <pem_file>]    Perform one-time server setup
  help                               Show this help

Examples:
  gokku remote add api ubuntu@192.168.105.3
  gokku remote list
  gokku remote remove api
  gokku remote setup ubuntu@192.168.105.3
  gokku remote setup ubuntu@ec2.example.com -i ~/.ssh/my-key.pem

The setup command will:
  - Install Gokku on the server
  - Install essential plugins (nginx, letsencrypt, cron, postgres, redis)
  - Configure SSH keys
  - Verify the installation

After adding a remote, you can:
  git push <remote_name> main`)
}
