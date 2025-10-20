package handlers

import (
	"fmt"
	"os"
	"strings"

	. "infra/internal"
)

// handleServer manages server connections and remotes
func handleServer(args []string) {
	if len(args) < 1 {
		printServerHelp()
		os.Exit(1)
	}

	subcommand := args[0]

	switch subcommand {
	case "add":
		handleServerAdd(args[1:])
	case "list":
		handleServerList()
	case "remove":
		handleServerRemove(args[1:])
	case "help", "--help", "-h":
		printServerHelp()
	default:
		fmt.Printf("Unknown server command: %s\n", subcommand)
		printServerHelp()
		os.Exit(1)
	}
}

// handleServerAdd adds a new server remote
func handleServerAdd(args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: gokku server add <app_name> <user@server_ip>")
		fmt.Println("Example: gokku server add stt ubuntu@54.233.138.116")
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
		fmt.Println("Example: ubuntu@54.233.138.116")
		os.Exit(1)
	}

	// Create git remote URL
	remoteURL := fmt.Sprintf("%s:/opt/gokku/repos/%s.git", serverHost, appName)

	// Add git remote
	output := Bash(fmt.Sprintf("git remote add %s %s", appName, remoteURL))

	if output == "" {
		fmt.Printf("Error adding remote '%s' -> %s\n", appName, remoteURL)
		os.Exit(1)
	}

	fmt.Printf("Added remote '%s' -> %s\n", appName, remoteURL)
	fmt.Printf("You can now use: gokku connect --remote %s\n", appName)
}

// handleServerList lists all configured remotes
func handleServerList() {
	output := Bash("git remote -v")

	if output == "" {
		fmt.Println("No remotes configured")
		fmt.Println("Add a remote with: gokku server add <app_name> <user@server_ip>")
		os.Exit(1)
	}

	fmt.Println("Configured remotes:")
	fmt.Println(output)
}

// handleServerRemove removes a server remote
func handleServerRemove(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: gokku server remove <remote_name>")
		fmt.Println("Example: gokku server remove stt")
		os.Exit(1)
	}

	remoteName := args[0]

	output := Bash(fmt.Sprintf("git remote remove %s", remoteName))

	if output == "" {
		fmt.Printf("Error removing remote '%s'\n", remoteName)
		os.Exit(1)
	}

	fmt.Printf("Removed remote '%s'\n", remoteName)
}

func printServerHelp() {
	fmt.Println(`Server Management Commands:

Usage:
  gokku server <command> [options]

Commands:
  add <app_name> <user@server_ip>    Add a new server remote
  list                               List all configured remotes
  remove <remote_name>               Remove a server remote
  help                               Show this help

Examples:
  gokku server add stt ubuntu@54.233.138.116
  gokku server list
  gokku server remove stt

After adding a remote, you can connect with:
  gokku connect --remote <remote_name>`)
}
