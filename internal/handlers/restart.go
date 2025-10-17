package handlers

import (
	"fmt"
	"os"
	"os/exec"

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
		// Local execution on server - recreate container with new env using Go
		baseDir := "/opt/gokku"
		if envVar := os.Getenv("GOKKU_BASE_DIR"); envVar != "" {
			baseDir = envVar
		}

		envFile := fmt.Sprintf("%s/apps/%s/%s/shared/.env", baseDir, app, env)
		appDir := fmt.Sprintf("%s/apps/%s/%s", baseDir, app, env)

		// Use Go Docker client to recreate container
		dc, err := internal.NewDockerClient()
		if err != nil {
			fmt.Printf("Error creating Docker client: %v\n", err)
			os.Exit(1)
		}

		if err := dc.RecreateActiveContainer(app, envFile, appDir); err != nil {
			fmt.Printf("Error recreating container: %v\n", err)
			os.Exit(1)
		}
	} else {
		// Remote execution via SSH - use gokku restart directly
		sshCmd := fmt.Sprintf("gokku restart %s %s", app, env)

		cmd := exec.Command("ssh", host, sshCmd)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Printf("Error recreating container: %v\n", err)
			os.Exit(1)
		}
	}
}
