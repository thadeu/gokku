package handlers

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"infra/internal"
)

// handleRestart restarts services/containers
func handleRestart(args []string) {
	remote, remainingArgs := internal.ExtractRemoteFlag(args)

	var app, env, host string
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
	} else if len(remainingArgs) >= 2 {
		// Check if running on server - allow local execution
		if internal.IsRunningOnServer() {
			localExecution = true
			app = remainingArgs[0]
			env = remainingArgs[1]
		} else {
			// Client without --remote - show error
			fmt.Println("Error: Local restart commands can only be run on the server")
			fmt.Println("")
			fmt.Println("For client usage, use --remote flag:")
			fmt.Println("  gokku restart <app> <env> --remote <git-remote>")
			fmt.Println("")
			fmt.Println("Or run this command directly on your server.")
			os.Exit(1)
		}
	} else {
		fmt.Println("Usage: gokku restart <app> <env>")
		fmt.Println("   or: gokku restart --remote <git-remote>")
		os.Exit(1)
	}

	serviceName := fmt.Sprintf("%s-%s", app, env)
	fmt.Printf("Restarting %s...\n", serviceName)

	if localExecution {
		// Local execution on server
		var cmd *exec.Cmd

		// Check systemd first
		systemdCmd := exec.Command("sudo", "systemctl", "list-units", "--all")
		output, err := systemdCmd.Output()
		if err == nil && strings.Contains(string(output), serviceName) {
			cmd = exec.Command("sudo", "systemctl", "restart", serviceName)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err == nil {
				fmt.Println("✓ Service restarted")
			} else {
				fmt.Printf("Error restarting service: %v\n", err)
				os.Exit(1)
			}
		} else {
			// Try docker
			dockerCmd := exec.Command("docker", "ps", "-a")
			output, err := dockerCmd.Output()
			if err == nil && strings.Contains(string(output), serviceName) {
				cmd = exec.Command("docker", "restart", serviceName)
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				if err := cmd.Run(); err == nil {
					fmt.Println("✓ Container restarted")
				} else {
					fmt.Printf("Error restarting container: %v\n", err)
					os.Exit(1)
				}
			} else {
				fmt.Printf("Error: Service or container '%s' not found\n", serviceName)
				os.Exit(1)
			}
		}
	} else {
		// Remote execution via SSH
		sshCmd := fmt.Sprintf(`
			if sudo systemctl list-units --all | grep -q %s; then
				sudo systemctl restart %s && echo "✓ Service restarted"
			elif docker ps -a | grep -q %s; then
				docker restart %s && echo "✓ Container restarted"
			else
				echo "Error: Service or container '%s' not found"
				exit 1
			fi
		`, serviceName, serviceName, serviceName, serviceName, serviceName)

		cmd := exec.Command("ssh", host, sshCmd)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()
	}
}
