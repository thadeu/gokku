package handlers

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"infra/internal"
)

// handleRollback rolls back to a previous release
func handleRollback(args []string) {
	remote, remainingArgs := internal.ExtractRemoteFlag(args)

	var app, env, host, baseDir string
	var releaseID string
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
		baseDir = remoteInfo.BaseDir

		if len(remainingArgs) > 0 {
			releaseID = remainingArgs[0]
		}
	} else if len(remainingArgs) >= 2 {
		// Check if running on server - allow local execution
		if internal.IsRunningOnServer() {
			localExecution = true
			app = remainingArgs[0]
			env = remainingArgs[1]
			if len(remainingArgs) > 2 {
				releaseID = remainingArgs[2]
			}
			baseDir = "/opt/gokku"
		} else {
			// Client without --remote - show error
			fmt.Println("Error: Local rollback commands can only be run on the server")
			fmt.Println("")
			fmt.Println("For client usage, use --remote flag:")
			fmt.Println("  gokku rollback <app> <env> [release-id] --remote <git-remote>")
			fmt.Println("")
			fmt.Println("Or run this command directly on your server.")
			os.Exit(1)
		}
	} else {
		fmt.Println("Usage: gokku rollback <app> <env> [release-id]")
		fmt.Println("   or: gokku rollback --remote <git-remote> [release-id]")
		os.Exit(1)
	}

	appDir := fmt.Sprintf("%s/apps/%s/%s", baseDir, app, env)
	serviceName := fmt.Sprintf("%s-%s", app, env)

	if releaseID == "" {
		// Get previous release
		if localExecution {
			cmd := exec.Command("bash", "-c", fmt.Sprintf("cd %s/releases && ls -t | sed -n '2p'", appDir))
			output, err := cmd.Output()
			if err != nil {
				fmt.Printf("Failed to get releases: %v\n", err)
				os.Exit(1)
			}
			releaseID = strings.TrimSpace(string(output))
		} else {
			listCmd := fmt.Sprintf("cd %s/releases && ls -t | sed -n '2p'", appDir)
			cmd := exec.Command("ssh", host, listCmd)
			output, err := cmd.Output()
			if err != nil {
				fmt.Printf("Failed to get releases: %v\n", err)
				os.Exit(1)
			}
			releaseID = strings.TrimSpace(string(output))
		}
	}

	if releaseID == "" {
		fmt.Println("No previous release found")
		os.Exit(1)
	}

	fmt.Printf("Rolling back %s (%s) to release: %s\n", app, env, releaseID)

	if localExecution {
		// Local execution on server
		var rollbackCmd string

		// Check if systemd service exists
		systemdCmd := exec.Command("sudo", "systemctl", "list-units", "--all")
		output, err := systemdCmd.Output()
		if err == nil && strings.Contains(string(output), serviceName) {
			rollbackCmd = fmt.Sprintf(`
				cd %s && \
				sudo systemctl stop %s && \
				ln -sfn releases/%s current && \
				sudo systemctl start %s && \
				echo "✓ Rollback complete"
			`, appDir, serviceName, releaseID, serviceName)
		} else {
			// Try docker
			dockerCmd := exec.Command("docker", "ps", "-a")
			output, err := dockerCmd.Output()
			if err == nil && strings.Contains(string(output), serviceName) {
				rollbackCmd = fmt.Sprintf(`
					cd %s && \
					docker stop %s && \
					docker rm %s && \
					docker run -d --name %s --env-file .env -p 8080:8080 %s:release-%s && \
					echo "✓ Rollback complete"
				`, appDir, serviceName, serviceName, serviceName, app, releaseID)
			} else {
				fmt.Printf("Error: Service or container '%s' not found\n", serviceName)
				os.Exit(1)
			}
		}

		cmd := exec.Command("bash", "-c", rollbackCmd)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Printf("Rollback failed: %v\n", err)
			os.Exit(1)
		}
	} else {
		// Remote execution via SSH
		rollbackCmd := fmt.Sprintf(`
			cd %s && \
			if sudo systemctl list-units --all | grep -q %s; then
				sudo systemctl stop %s && \
				ln -sfn %s/releases/%s current && \
				sudo systemctl start %s && \
				echo "✓ Rollback complete"
			elif docker ps -a | grep -q %s; then
				docker stop %s && \
				docker rm %s && \
				docker run -d --name %s --env-file %s/.env -p 8080:8080 %s:release-%s && \
				echo "✓ Rollback complete"
			else
				echo "Error: Service or container not found"
				exit 1
			fi
		`, appDir, serviceName, serviceName, appDir, releaseID, serviceName, serviceName, serviceName, serviceName, serviceName, appDir, app, releaseID)

		cmd := exec.Command("ssh", host, rollbackCmd)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			fmt.Printf("Rollback failed: %v\n", err)
			os.Exit(1)
		}
	}
}
