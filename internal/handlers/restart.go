package handlers

import (
	"fmt"
	"os"
	"os/exec"

	"infra/internal"
)

// handleRestart restarts services/containers by doing a full rebuild and redeploy
func handleRestart(args []string) {
	remote, remainingArgs := internal.ExtractRemoteFlag(args)

	var app, host string
	var localExecution bool

	if remote != "" {
		remoteInfo, err := internal.GetRemoteInfo(remote)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		app = remoteInfo.App
		host = remoteInfo.Host
	} else if len(remainingArgs) >= 1 {
		// Check if running on server - allow local execution
		if internal.IsRunningOnServer() {
			localExecution = true
			app = remainingArgs[0]
		} else {
			// Client without --remote - show error
			fmt.Println("Error: Local restart commands can only be run on the server")
			fmt.Println("")
			fmt.Println("For client usage, use --remote flag:")
			fmt.Println("  gokku restart <app> --remote <git-remote>")
			fmt.Println("")
			fmt.Println("Or run this command directly on your server.")
			os.Exit(1)
		}
	} else {
		fmt.Println("Usage: gokku restart <app>")
		fmt.Println("   or: gokku restart --remote <git-remote>")
		os.Exit(1)
	}

	fmt.Printf("Restarting %s...\n", app)

	if localExecution {
		// Local execution on server - do a full rebuild and redeploy
		fmt.Printf("-----> Restarting %s (full rebuild and redeploy)...\n", app)

		if err := executeDirectDeployment(app); err != nil {
			fmt.Printf("Error restarting app: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("\n✓ Restart complete!")
	} else {
		// Remote execution via SSH - use gokku restart directly
		sshCmd := fmt.Sprintf("gokku restart %s", app)

		cmd := exec.Command("ssh", host, sshCmd)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			fmt.Printf("Error restarting app: %v\n", err)
			os.Exit(1)
		}
	}
}
