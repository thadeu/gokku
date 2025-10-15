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
		app = remainingArgs[0]
		env = remainingArgs[1]
		if len(remainingArgs) > 2 {
			releaseID = remainingArgs[2]
		}

		config, err := internal.LoadConfig()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			os.Exit(1)
		}

		server := internal.GetDefaultServer(config)
		if server == nil {
			fmt.Println("No servers configured")
			os.Exit(1)
		}
		host = server.Host
		baseDir = server.BaseDir
	} else {
		fmt.Println("Usage: gokku rollback <app> <env> [release-id]")
		fmt.Println("   or: gokku rollback --remote <git-remote> [release-id]")
		os.Exit(1)
	}

	appDir := fmt.Sprintf("%s/apps/%s/%s", baseDir, app, env)
	serviceName := fmt.Sprintf("%s-%s", app, env)

	if releaseID == "" {
		// Get previous release
		listCmd := fmt.Sprintf("cd %s/releases && ls -t | sed -n '2p'", appDir)
		cmd := exec.Command("ssh", host, listCmd)
		output, err := cmd.Output()
		if err != nil {
			fmt.Printf("Failed to get releases: %v\n", err)
			os.Exit(1)
		}
		releaseID = strings.TrimSpace(string(output))
	}

	if releaseID == "" {
		fmt.Println("No previous release found")
		os.Exit(1)
	}

	fmt.Printf("Rolling back %s (%s) to release: %s\n", app, env, releaseID)

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
