package handlers

import (
	"fmt"
	"os"
	"strings"

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
		fmt.Printf("Connecting to %s (%s)...\n", host, remoteInfo.App)
	}

	output := internal.Bash(fmt.Sprintf("ssh %s %s", host, strings.Join(remainingArgs, " ")))

	if output == "" {
		fmt.Printf("Error running command: %s\n", strings.Join(remainingArgs, " "))
		os.Exit(1)
	}

	fmt.Println(output)
}
