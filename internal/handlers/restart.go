package handlers

import (
	"fmt"
	"os"
	"os/exec"

	"infra/internal"
)

// handleRestart restarts services/containers by doing a full rebuild and redeploy
func handleRestart(args []string) {
	appName, remainingArgs := internal.ExtractAppFlag(args)

	var app, host string
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
		} else {
			// Client mode without -a flag
			fmt.Println("Error: Client mode requires -a flag to specify app")
			fmt.Println("")
			fmt.Println("Usage: gokku restart -a <app>")
			fmt.Println("")
			fmt.Println("Examples:")
			fmt.Println("  gokku restart -a api-production")
			os.Exit(1)
		}
	} else {
		// Server mode: -a flag uses app name directly, no git remote needed
		if appName != "" {
			localExecution = true
			app = appName
		} else if len(remainingArgs) >= 1 {
			// Server mode with positional args
			localExecution = true
			app = remainingArgs[0]
		} else {
			// Server mode without args
			fmt.Println("Error: Server mode requires -a flag to specify app")
			fmt.Println("")
			fmt.Println("Usage: gokku restart -a <app>")
			fmt.Println("")
			fmt.Println("Examples:")
			fmt.Println("  gokku restart -a api")
			os.Exit(1)
		}
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
