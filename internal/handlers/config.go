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
		fmt.Println("  -a, --app <app>           App name")
		fmt.Println("")
		fmt.Println("Examples:")
		fmt.Println("  # Client mode (from local machine)")
		fmt.Println("  gokku config set PORT=8080 -a api-production")
		fmt.Println("  gokku config list -a api-production")
		fmt.Println("")
		fmt.Println("  # Server mode (on server)")
		fmt.Println("  gokku config set PORT=8080 -a api")
		fmt.Println("  gokku config list -a api")
		os.Exit(1)
	}

	app, remainingArgs := internal.ExtractAppFlag(args)

	// Check if we're in client mode or server mode
	if internal.IsClientMode() {
		// Client mode: -a flag requires git remote for SSH execution
		if app != "" {
			remoteInfo, err := internal.GetRemoteInfo(app)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}

			if len(remainingArgs) < 1 {
				fmt.Println("Usage: gokku config <set|get|list|unset> [args...] -a <app>")
				os.Exit(1)
			}

			subcommand := remainingArgs[0]

			// Build command to run on server
			var sshCmd string
			switch subcommand {
			case "set":
				if len(remainingArgs) < 2 {
					fmt.Println("Usage: gokku config set KEY=VALUE [KEY2=VALUE2...] -a <app>")
					os.Exit(1)
				}
				pairs := strings.Join(remainingArgs[1:], " ")
				sshCmd = fmt.Sprintf("gokku config set %s --app %s", pairs, remoteInfo.App)
			case "get":
				if len(remainingArgs) < 2 {
					fmt.Println("Usage: gokku config get KEY -a <app>")
					os.Exit(1)
				}
				key := remainingArgs[1]
				sshCmd = fmt.Sprintf("gokku config get %s --app %s", key, remoteInfo.App)
			case "list":
				sshCmd = fmt.Sprintf("gokku config list --app %s", remoteInfo.App)
			case "unset":
				if len(remainingArgs) < 2 {
					fmt.Println("Usage: gokku config unset KEY [KEY2...] -a <app>")
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
					fmt.Printf("Warning: Failed to restart container. Run 'gokku restart -a %s' manually.\n", app)
				} else {
					fmt.Printf("✓ Container restarted with new configuration\n")
				}
			}
			return
		} else {
			// Client mode without -a flag
			fmt.Println("Error: Client mode requires -a flag to specify app")
			fmt.Println("")
			fmt.Println("Usage: gokku config <command> [args...] -a <app>")
			fmt.Println("")
			fmt.Println("Examples:")
			fmt.Println("  gokku config set PORT=8080 -a api-production")
			fmt.Println("  gokku config list -a api-production")
			os.Exit(1)
		}
	} else {
		// Server mode: -a flag uses app name directly, no git remote needed
		if app == "" {
			fmt.Println("Error: Server mode requires -a flag to specify app")
			fmt.Println("")
			fmt.Println("Usage: gokku config <command> [args...] -a <app>")
			fmt.Println("")
			fmt.Println("Examples:")
			fmt.Println("  gokku config set PORT=8080 -a api")
			fmt.Println("  gokku config list -a api")
			os.Exit(1)
		}
	}

	// Server mode execution - use app name directly
	appName := app
	var finalArgs []string

	// Parse remaining args for env flags
	for i := 0; i < len(remainingArgs); i++ {
		if (remainingArgs[i] == "--env" || remainingArgs[i] == "-e") && i+1 < len(remainingArgs) {
			// Skip env flag and value for now (could be used in future)
			i++
		} else {
			finalArgs = append(finalArgs, remainingArgs[i])
		}
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
		if len(finalArgs) < 1 {
			fmt.Println("Usage: gokku config set KEY=VALUE [KEY2=VALUE2...] --app <app>")
			os.Exit(1)
		}
		internal.EnvSet(envFile, finalArgs[1:])
	case "get":
		if len(finalArgs) < 1 {
			fmt.Println("Usage: gokku config get KEY --app <app>")
			os.Exit(1)
		}
		internal.EnvGet(envFile, finalArgs[1])
	case "list":
		internal.EnvList(envFile)
	case "unset":
		if len(finalArgs) < 1 {
			fmt.Println("Usage: gokku config unset KEY [KEY2...] --app <app>")
			os.Exit(1)
		}
		internal.EnvUnset(envFile, finalArgs[1:])
	default:
		fmt.Printf("Unknown command: %s\n", command)
		fmt.Println("Available: set, get, list, unset")
		os.Exit(1)
	}
}
