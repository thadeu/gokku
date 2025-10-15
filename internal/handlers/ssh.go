package handlers

import (
	"fmt"
	"os"
	"os/exec"

	"infra/internal"
)

// handleSSH establishes SSH connections to servers
func handleSSH(args []string) {
	remote, remainingArgs := internal.ExtractRemoteFlag(args)

	var host string

	if remote != "" {
		remoteInfo, err := internal.GetRemoteInfo(remote)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		host = remoteInfo.Host
		fmt.Printf("Connecting to %s (%s/%s)...\n", host, remoteInfo.App, remoteInfo.Env)
	} else {
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
		fmt.Printf("Connecting to %s...\n", server.Name)
	}

	cmd := exec.Command("ssh", append([]string{"-t", host}, remainingArgs...)...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Run()
}
