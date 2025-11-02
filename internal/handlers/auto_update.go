package handlers

import (
	"fmt"
	"os"
	"os/exec"

	"gokku/internal"
)

func handleAutoUpdate(args []string) {
	// Get remote info (or nil if server mode)
	remoteInfo, remainingArgs, err := internal.GetRemoteInfoOrDefault(args)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	command := "curl -fsSL https://gokku-vm.com/install | bash -s --"

	if remoteInfo != nil {
		// Client mode: execute on remote server
		mode := " --server"
		command += mode
		cmd := fmt.Sprintf("bash -c '%s'", command)
		if err := internal.ExecuteRemoteCommand(remoteInfo, cmd); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Server mode: execute locally
	isServerMode := internal.IsServerMode()
	mode := " --server"
	if !isServerMode {
		mode = " --client"
	}

	command += mode
	cmd := exec.Command("bash", "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	_ = remainingArgs // unused for now
}
