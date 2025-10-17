package handlers

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"infra/internal"
)

// handleConfig manages environment variable configuration
func handleConfig(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: gokku config <set|get|list|unset> [KEY[=VALUE]] [options]")
		fmt.Println("")
		fmt.Println("Options:")
		fmt.Println("  --remote <git-remote>     Execute on remote server via SSH")
		fmt.Println("  --app, -a <app>           App name (required for local execution)")
		fmt.Println("  --env, -e <env>           Environment name (optional, defaults to 'default')")
		fmt.Println("")
		fmt.Println("Examples:")
		fmt.Println("  # Remote execution (from local machine)")
		fmt.Println("  gokku config set PORT=8080 --remote api-production")
		fmt.Println("  gokku config list --remote api-production")
		fmt.Println("")
		fmt.Println("  # Local execution (on server)")
		fmt.Println("  gokku config set PORT=8080 --app api")
		fmt.Println("  gokku config set PORT=8080 -a api                     (uses 'default' env)")
		fmt.Println("  gokku config set PORT=8080 -a api -e production       (explicit env)")
		fmt.Println("  gokku config list -a api                              (uses 'default' env)")
		os.Exit(1)
	}

	remote, remainingArgs := internal.ExtractRemoteFlag(args)

	// If --remote is provided, execute via SSH
	if remote != "" {
		remoteInfo, err := internal.GetRemoteInfo(remote)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		if len(remainingArgs) < 1 {
			fmt.Println("Usage: gokku config <set|get|list|unset> [args...] --remote <git-remote>")
			os.Exit(1)
		}

		subcommand := remainingArgs[0]

		// Build command to run on server
		var sshCmd string
		switch subcommand {
		case "set":
			if len(remainingArgs) < 2 {
				fmt.Println("Usage: gokku config set KEY=VALUE [KEY2=VALUE2...] --remote <git-remote>")
				os.Exit(1)
			}
			pairs := strings.Join(remainingArgs[1:], " ")
			sshCmd = fmt.Sprintf("gokku config set %s --app %s", pairs, remoteInfo.App)
		case "get":
			if len(remainingArgs) < 2 {
				fmt.Println("Usage: gokku config get KEY --remote <git-remote>")
				os.Exit(1)
			}
			key := remainingArgs[1]
			sshCmd = fmt.Sprintf("gokku config get %s --app %s", key, remoteInfo.App)
		case "list":
			sshCmd = fmt.Sprintf("gokku config list --app %s", remoteInfo.App)
		case "unset":
			if len(remainingArgs) < 2 {
				fmt.Println("Usage: gokku config unset KEY [KEY2...] --remote <git-remote>")
				os.Exit(1)
			}
			keys := strings.Join(remainingArgs[1:], " ")
			sshCmd = fmt.Sprintf("gokku config unset %s --app %s", keys, remoteInfo.App)
		default:
			fmt.Printf("Unknown subcommand: %s\n", subcommand)
			os.Exit(1)
		}

		fmt.Printf("→ %s (%s)\n", remoteInfo.App, remoteInfo.Host)

		cmd := exec.Command("ssh", remoteInfo.Host, sshCmd)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		if err := cmd.Run(); err != nil {
			os.Exit(1)
		}

		// Auto-restart container after set/unset to apply changes
		if subcommand == "set" || subcommand == "unset" {
			fmt.Printf("\n-----> Restarting container to apply changes...\n")
			restartCmd := exec.Command("ssh", remoteInfo.Host, fmt.Sprintf("gokku restart %s", remoteInfo.App))
			restartCmd.Stdout = os.Stdout
			restartCmd.Stderr = os.Stderr
			if err := restartCmd.Run(); err != nil {
				fmt.Printf("Warning: Failed to restart container. Run 'gokku restart --remote %s' manually.\n", remote)
			} else {
				fmt.Printf("✓ Container restarted with new configuration\n")
			}
		}
		return
	}

	// Check if we're running on server - local execution only works on server
	if !internal.IsRunningOnServer() {
		fmt.Println("Error: Local config commands can only be run on the server")
		fmt.Println("")
		fmt.Println("For client usage, use --remote flag:")
		fmt.Println("  gokku config set KEY=VALUE --remote <git-remote>")
		fmt.Println("  gokku config get KEY --remote <git-remote>")
		fmt.Println("  gokku config list --remote <git-remote>")
		fmt.Println("  gokku config unset KEY --remote <git-remote>")
		fmt.Println("")
		fmt.Println("Or run this command directly on your server.")
		os.Exit(1)
	}

	// Local execution - parse --app and --env flags
	var appName string
	var finalArgs []string

	for i := 0; i < len(remainingArgs); i++ {
		if (remainingArgs[i] == "--app" || remainingArgs[i] == "-a") && i+1 < len(remainingArgs) {
			appName = remainingArgs[i+1]
			i++
		} else {
			finalArgs = append(finalArgs, remainingArgs[i])
		}
	}

	// If no app specified, error
	if appName == "" {
		fmt.Println("Error: --app is required for local execution")
		fmt.Println("")
		fmt.Println("Usage: gokku config <command> [args...] --app <app> [--env <env>]")
		fmt.Println("   or: gokku config <command> [args...] -a <app> [-e <env>]")
		fmt.Println("")
		fmt.Println("Examples:")
		fmt.Println("  gokku config set PORT=8080 --app api                    (uses 'default' env)")
		fmt.Println("  gokku config set PORT=8080 -a api                        (uses 'default' env)")
		fmt.Println("  gokku config set PORT=8080 -a api -e production          (explicit env)")
		fmt.Println("  gokku config list -a api                                 (uses 'default' env)")
		os.Exit(1)
	}

	if len(finalArgs) < 1 {
		fmt.Println("Error: command is required (set, get, list, unset)")
		os.Exit(1)
	}

	command := finalArgs[0]

	// Determine env file path
	baseDir := "/opt/gokku"
	if envVar := os.Getenv("GOKKU_BASE_DIR"); envVar != "" {
		baseDir = envVar
	}

	envFile := filepath.Join(baseDir, "apps", appName, "shared", ".env")

	// Ensure directory exists
	envDir := filepath.Dir(envFile)
	if err := os.MkdirAll(envDir, 0755); err != nil {
		fmt.Printf("Error creating directory: %v\n", err)
		os.Exit(1)
	}

	switch command {
	case "set":
		if len(finalArgs) < 2 {
			fmt.Println("Usage: gokku config set KEY=VALUE [KEY2=VALUE2...] --app <app> --env <env>")
			os.Exit(1)
		}
		internal.EnvSet(envFile, finalArgs[1:])
	case "get":
		if len(finalArgs) < 2 {
			fmt.Println("Usage: gokku config get KEY --app <app> --env <env>")
			os.Exit(1)
		}
		internal.EnvGet(envFile, finalArgs[1])
	case "list":
		internal.EnvList(envFile)
	case "unset":
		if len(finalArgs) < 2 {
			fmt.Println("Usage: gokku config unset KEY [KEY2...] --app <app> --env <env>")
			os.Exit(1)
		}
		internal.EnvUnset(envFile, finalArgs[1:])
	default:
		fmt.Printf("Unknown command: %s\n", command)
		fmt.Println("Available: set, get, list, unset")
		os.Exit(1)
	}
}
