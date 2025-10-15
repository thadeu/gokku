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
			sshCmd = fmt.Sprintf("gokku config set %s --app %s --env %s", pairs, remoteInfo.App, remoteInfo.Env)
		case "get":
			if len(remainingArgs) < 2 {
				fmt.Println("Usage: gokku config get KEY --remote <git-remote>")
				os.Exit(1)
			}
			key := remainingArgs[1]
			sshCmd = fmt.Sprintf("gokku config get %s --app %s --env %s", key, remoteInfo.App, remoteInfo.Env)
		case "list":
			sshCmd = fmt.Sprintf("gokku config list --app %s --env %s", remoteInfo.App, remoteInfo.Env)
		case "unset":
			if len(remainingArgs) < 2 {
				fmt.Println("Usage: gokku config unset KEY [KEY2...] --remote <git-remote>")
				os.Exit(1)
			}
			keys := strings.Join(remainingArgs[1:], " ")
			sshCmd = fmt.Sprintf("gokku config unset %s --app %s --env %s", keys, remoteInfo.App, remoteInfo.Env)
		default:
			fmt.Printf("Unknown subcommand: %s\n", subcommand)
			os.Exit(1)
		}

		fmt.Printf("→ %s/%s (%s)\n", remoteInfo.App, remoteInfo.Env, remoteInfo.Host)

		cmd := exec.Command("ssh", remoteInfo.Host, sshCmd)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		cmd.Run()
		return
	}

	// Local execution - parse --app and --env flags
	var appName, envName string
	var finalArgs []string

	for i := 0; i < len(remainingArgs); i++ {
		if (remainingArgs[i] == "--app" || remainingArgs[i] == "-a") && i+1 < len(remainingArgs) {
			appName = remainingArgs[i+1]
			i++
		} else if (remainingArgs[i] == "--env" || remainingArgs[i] == "-e") && i+1 < len(remainingArgs) {
			envName = remainingArgs[i+1]
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

	// Default environment if not specified
	if envName == "" {
		envName = "default"
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

	envFile := filepath.Join(baseDir, "apps", appName, envName, ".env")

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
