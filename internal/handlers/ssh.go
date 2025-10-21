package handlers

import (
	"fmt"
	"os"
	"os/exec"

	"infra/internal"
)

// handleSSH establishes SSH connections to servers
func handleSSH(args []string) {
	app, remainingArgs := internal.ExtractAppFlag(args)

	var host string

	if app != "" {
		remoteInfo, err := internal.GetRemoteInfo(app)

		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		host = remoteInfo.Host
		fmt.Printf("Connecting to %s (%s)...\n", host, remoteInfo.App)
	}

	// Build SSH command with proper arguments
	sshArgs := []string{"-t", host}
	sshArgs = append(sshArgs, remainingArgs...)

	cmd := exec.Command("ssh", sshArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	// Run SSH and let it take over the terminal
	err := cmd.Run()

	if err != nil {
		// SSH returns non-zero exit codes for various reasons
		// Don't treat this as a fatal error unless it's a connection issue
		os.Exit(1)
	}
}
