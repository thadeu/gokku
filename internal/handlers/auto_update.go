package handlers

import (
	"fmt"
	"os"
	"os/exec"

	"gokku/internal"
)

func handleAutoUpdate(args []string) {
	isServerMode := internal.IsServerMode()

	command := "curl -fsSL https://gokku-vm.com/install | bash -s --"
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
}
