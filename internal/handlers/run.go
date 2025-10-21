package handlers

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"infra/internal"
)

// handleRun executes arbitrary commands on remote servers or locally on server
func handleRun(args []string) {
	app, remainingArgs := internal.ExtractAppFlag(args)

	if app != "" {
		// Remote execution
		remoteInfo, err := internal.GetRemoteInfo(app)

		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		if len(remainingArgs) < 1 {
			fmt.Println("Error: command is required")
			fmt.Println("Usage: gokku run <command> -a <app>")
			os.Exit(1)
		}

		// Join all remaining args as the command
		command := strings.Join(remainingArgs, " ")

		// Build the docker exec command to run in the active container
		containerName := remoteInfo.App
		dockerCommand := fmt.Sprintf("docker exec -it %s %s", containerName, command)

		fmt.Printf("→ %s (%s)\n", remoteInfo.App, remoteInfo.Host)
		fmt.Printf("$ %s\n\n", command)

		cmd := exec.Command("ssh", "-t", remoteInfo.Host, dockerCommand)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		cmd.Run()
	} else {
		// Local execution - check if running on server
		if !internal.IsRunningOnServer() {
			fmt.Println("Error: Local run commands can only be run on the server")
			fmt.Println("")
			fmt.Println("For client usage, use -a flag:")
			fmt.Println("  gokku run <command> -a <app>")
			fmt.Println("")
			fmt.Println("Or run this command directly on your server.")
			os.Exit(1)
		}

		if len(remainingArgs) < 2 {
			fmt.Println("Error: app name and command are required")
			fmt.Println("Usage: gokku run <command> --app <app>")
			os.Exit(1)
		}

		// Join all remaining args as the command
		command := strings.Join(remainingArgs[1:], " ")
		appName := remainingArgs[0]

		// Build the docker exec command to run in the active container
		containerName := appName
		dockerCommand := fmt.Sprintf("docker exec -it %s %s", containerName, command)

		fmt.Printf("$ %s\n\n", command)

		cmd := exec.Command("bash", "-c", dockerCommand)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		cmd.Run()
	}
}
