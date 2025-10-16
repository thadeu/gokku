package handlers

import (
	"fmt"
	"os"
	"os/exec"

	"infra/internal"
)

// handleLogs shows application logs
func handleLogs(args []string) {
	remote, remainingArgs := internal.ExtractRemoteFlag(args)

	var app, env, host string
	var follow bool
	var localExecution bool

	if remote != "" {
		remoteInfo, err := internal.GetRemoteInfo(remote)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		app = remoteInfo.App
		env = remoteInfo.Env
		host = remoteInfo.Host

		// Check for -f flag
		for _, arg := range remainingArgs {
			if arg == "-f" {
				follow = true
				break
			}
		}
	} else {
		// Check if running on server - allow local execution
		if internal.IsRunningOnServer() {
			localExecution = true
			if len(remainingArgs) < 2 {
				fmt.Println("Usage: gokku logs <app> <env> [-f]")
				fmt.Println("   or: gokku logs --remote <git-remote> [-f]")
				os.Exit(1)
			}
			app = remainingArgs[0]
			env = remainingArgs[1]
			follow = len(remainingArgs) > 2 && remainingArgs[2] == "-f"
		} else {
			// Client without --remote - show error
			fmt.Println("Error: Local logs commands can only be run on the server")
			fmt.Println("")
			fmt.Println("For client usage, use --remote flag:")
			fmt.Println("  gokku logs <app> <env> [-f] --remote <git-remote>")
			fmt.Println("")
			fmt.Println("Or run this command directly on your server.")
			os.Exit(1)
		}
	}

	serviceName := fmt.Sprintf("%s-%s", app, env)
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
			if sudo systemctl list-units --all | grep -q %s; then
				sudo journalctl -u %s %s -n 100
			elif docker ps -a | grep -q %s; then
				docker logs %s %s
			else
				echo "Service or container '%s' not found"
				exit 1
			fi
		`, serviceName, serviceName, followFlag, serviceName, followFlag, serviceName, serviceName)

		cmd := exec.Command("ssh", "-t", host, sshCmd)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		cmd.Run()
	}
}
