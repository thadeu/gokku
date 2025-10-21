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
	appName, remainingArgs := internal.ExtractAppFlag(args)

	var app, host, baseDir string
	var releaseID string
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
			baseDir = remoteInfo.BaseDir

			if len(remainingArgs) > 0 {
				releaseID = remainingArgs[0]
			}
		} else {
			// Client mode without -a flag
			fmt.Println("Error: Client mode requires -a flag to specify app")
			fmt.Println("")
			fmt.Println("Usage: gokku rollback -a <app> [release-id]")
			fmt.Println("")
			fmt.Println("Examples:")
			fmt.Println("  gokku rollback -a api-production")
			fmt.Println("  gokku rollback -a api-production 20231201-123456")
			os.Exit(1)
		}
	} else {
		// Server mode: -a flag uses app name directly, no git remote needed
		if appName != "" {
			localExecution = true
			app = appName
			baseDir = "/opt/gokku"

			if len(remainingArgs) > 0 {
				releaseID = remainingArgs[0]
			}
		} else if len(remainingArgs) >= 1 {
			// Server mode with positional args
			localExecution = true
			app = remainingArgs[0]
			baseDir = "/opt/gokku"

			if len(remainingArgs) > 1 {
				releaseID = remainingArgs[1]
			}
		} else {
			// Server mode without args
			fmt.Println("Error: Server mode requires -a flag to specify app")
			fmt.Println("")
			fmt.Println("Usage: gokku rollback -a <app> [release-id]")
			fmt.Println("")
			fmt.Println("Examples:")
			fmt.Println("  gokku rollback -a api")
			fmt.Println("  gokku rollback -a api 20231201-123456")
			os.Exit(1)
		}
	}

	appDir := fmt.Sprintf("%s/apps/%s", baseDir, app)
	serviceName := app

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

	fmt.Printf("Rolling back %s to release: %s\n", app, releaseID)

	if localExecution {
		// Local execution on server

		// Rollback command
		rollbackCmd := fmt.Sprintf(`
			cd %s && \
			docker stop %s && \
			docker rm -f %s && \
			docker run -d --name %s --env-file %s/shared/.env %s:release-%s && \
			docker start %s && \
			echo "✓ Rollback complete"
		`, appDir, serviceName, serviceName, serviceName, appDir, app, releaseID, serviceName)

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
			if docker ps -a | grep -q %s; then
				docker stop %s && \
				docker rm -f %s && \
				docker run -d --name %s --env-file %s/shared/.env %s:release-%s && \
				docker start %s && \
				echo "✓ Rollback complete"
			else
				echo "Error: Service or container not found"
				exit 1
			fi
		`, appDir, serviceName, serviceName, serviceName, serviceName, appDir, app, releaseID, serviceName)

		cmd := exec.Command("ssh", host, rollbackCmd)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			fmt.Printf("Rollback failed: %v\n", err)
			os.Exit(1)
		}
	}
}
