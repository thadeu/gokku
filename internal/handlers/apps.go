package handlers

import (
	"fmt"
	"os"
	"os/exec"

	"infra/internal"
)

// handleApps lists applications on the server
func handleApps(args []string) {
	remote, _ := internal.ExtractRemoteFlag(args)

	if remote != "" {
		fmt.Printf("Note: --remote flag ignored for 'apps' command\n\n")
	}

	config, err := internal.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	server := internal.GetDefaultServer(config)
	if server == nil {
		fmt.Println("No servers configured")
		fmt.Println("Add a server: gokku server add production ubuntu@ec2.compute.amazonaws.com")
		os.Exit(1)
	}

	fmt.Printf("Listing apps on %s...\n", server.Name)

	cmd := exec.Command("ssh", server.Host, fmt.Sprintf("ls -1 %s/repos 2>/dev/null | sed 's/.git//'", server.BaseDir))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}
