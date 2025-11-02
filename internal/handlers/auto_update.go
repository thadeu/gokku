package handlers

import (
	"fmt"
	"os"
	"os/exec"

	"gokku/internal"
)

func handleAutoUpdate(args []string) {
	command := "curl -fsSL https://gokku-vm.com/install | bash -s --"

	// Extract --remote flag directly (no fallback to "gokku" remote)
	// If --remote is not present, execute locally (update client)
	remote, remainingArgs := internal.ExtractRemoteFlag(args)

	if remote != "" {
		// --remote flag is present: execute on remote server
		remoteInfo, err := internal.GetRemoteInfo(remote)
		if err != nil {
			fmt.Printf("Error: remote '%s' not found: %v. Add it with: gokku remote add %s user@host\n", remote, err, remote)
			os.Exit(1)
		}

		mode := " --server"
		command += mode
		cmd := fmt.Sprintf("bash -c '%s'", command)
		if err := internal.ExecuteRemoteCommand(remoteInfo, cmd); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// No --remote flag: execute locally (update client)
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
