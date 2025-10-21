package handlers

import (
	"fmt"
	"os"
	"os/exec"

	"infra/internal"
)

// handleLogs shows application logs
func handleLogs(args []string) {
	appName, remainingArgs := internal.ExtractAppFlag(args)

	var app, host string
	var follow bool
	var localExecution bool

	// Check if we're in client mode or server mode
	if internal.IsClientMode() {
		// Client mode: -a flag requires git remote for SSH execution
		if appName != "" {
			remoteInfo, err := internal.GetRemoteInfo(appName)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
			app = remoteInfo.App
			host = remoteInfo.Host

			// Check for -f flag
			for _, arg := range remainingArgs {
				if arg == "-f" {
					follow = true
					break
				}
			}
		} else {
			// Client mode without -a flag
			fmt.Println("Error: Client mode requires -a flag to specify app")
			fmt.Println("")
			fmt.Println("Usage: gokku logs -a <app> [-f]")
			fmt.Println("")
			fmt.Println("Examples:")
			fmt.Println("  gokku logs -a api-production")
			fmt.Println("  gokku logs -a api-production -f")
			os.Exit(1)
		}
	} else {
		// Server mode: -a flag uses app name directly, no git remote needed
		if appName == "" {
			fmt.Println("Error: Server mode requires -a flag to specify app")
			fmt.Println("")
			fmt.Println("Usage: gokku logs -a <app> [-f]")
			fmt.Println("")
			fmt.Println("Examples:")
			fmt.Println("  gokku logs -a api")
			fmt.Println("  gokku logs -a api -f")
			os.Exit(1)
		}

		localExecution = true
		app = appName

		// Check for -f flag
		for _, arg := range remainingArgs {
			if arg == "-f" {
				follow = true
				break
			}
		}
	}

	serviceName := app
	followFlag := ""
	if follow {
		followFlag = "-f"
	}

	if localExecution {
		// Execute locally on server
		var cmd *exec.Cmd
		if follow {
			followFlag = "-f"
		}

		if follow {
			// For follow mode, use journalctl/docker logs directly
			cmd = exec.Command("docker", "logs", followFlag, serviceName)
		} else {
			// Try docker
			cmd = exec.Command("docker", "logs", serviceName)
		}

		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		cmd.Run()
	} else {
		// Execute via SSH
		sshCmd := fmt.Sprintf(`
			if docker ps -a | grep -q %s; then
				docker logs %s %s
			else
				echo "Container '%s' not found"
				exit 1
			fi
		`, serviceName, serviceName, followFlag, serviceName)

		cmd := exec.Command("ssh", "-t", host, sshCmd)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		cmd.Run()
	}
}
