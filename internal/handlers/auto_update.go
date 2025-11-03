package handlers

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"gokku/internal"
)

func handleAutoUpdate(args []string) {
	command := "curl -fsSL https://gokku-vm.com/install | bash -s --"

	remoteInfo, remainingArgs, err := internal.GetRemoteInfoOrDefault(args)

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	IsClientMode := internal.IsClientMode()
	hasRemoteFlag := strings.Contains(strings.Join(remainingArgs, " "), "--remote")

	if IsClientMode && hasRemoteFlag {
		command += " --server"

		cmd := fmt.Sprintf("bash -c '%s'", command)
		fmt.Println("Executing command auto-update on server: ", cmd)

		if err := internal.ExecuteRemoteCommand(remoteInfo, cmd); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		return
	}

	command += " --client"
	fmt.Println("Executing command auto-update on client: ", command)

	cmd := exec.Command("bash", "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	_ = remainingArgs // unused for now
}
