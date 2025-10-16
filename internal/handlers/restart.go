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
		// Local execution on server - Docker only

		// Check for blue or green container first, then fallback to service name
		containerName := serviceName
		dockerCmd := exec.Command("sudo", "docker", "ps", "-a", "--format", "{{.Names}}")
		output, err := dockerCmd.Output()

		if err != nil {
			fmt.Printf("Error checking Docker containers: %v\n", err)
			os.Exit(1)
		}

		containers := string(output)
		if strings.Contains(containers, app+"-blue") {
			containerName = app + "-blue"
		} else if strings.Contains(containers, app+"-green") {
			containerName = app + "-green"
		}

		// Restart the container
		cmd := exec.Command("sudo", "docker", "restart", containerName)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err == nil {
			fmt.Printf("✓ Container restarted: %s\n", containerName)
		} else {
			fmt.Printf("Error restarting container: %v\n", err)
			os.Exit(1)
		}
	} else {
		// Remote execution via SSH - Docker only
		sshCmd := fmt.Sprintf(`
			# Check for blue or green container first
			CONTAINER_NAME="%s"
			CONTAINERS=$(sudo docker ps -a --format "{{.Names}}")

			if echo "$CONTAINERS" | grep -q "%s-blue"; then
				CONTAINER_NAME="%s-blue"
			elif echo "$CONTAINERS" | grep -q "%s-green"; then
				CONTAINER_NAME="%s-green"
			fi

			# Restart the container
			if sudo docker restart "$CONTAINER_NAME" 2>/dev/null; then
				echo "✓ Container restarted: $CONTAINER_NAME"
			else
				echo "Error: Container '$CONTAINER_NAME' not found or failed to restart"
				exit 1
			fi
		`, serviceName, app, app, app, app)

		cmd := exec.Command("ssh", host, sshCmd)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()
	}
}
