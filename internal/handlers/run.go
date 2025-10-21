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

	// Check if we're in client mode or server mode
	if internal.IsClientMode() {
		// Client mode: -a flag requires git remote for SSH execution
		if app != "" {
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
			// Client mode without -a flag
			fmt.Println("Error: Client mode requires -a flag to specify app")
			fmt.Println("")
			fmt.Println("Usage: gokku run <command> -a <app>")
			fmt.Println("")
			fmt.Println("Examples:")
			fmt.Println("  gokku run bash -a api-production")
			fmt.Println("  gokku run rails console -a api-production")
			os.Exit(1)
		}
	} else {
		// Server mode: -a flag uses app name directly, no git remote needed
		if app == "" {
			fmt.Println("Error: Server mode requires -a flag to specify app")
			fmt.Println("")
			fmt.Println("Usage: gokku run <command> -a <app>")
			fmt.Println("")
			fmt.Println("Examples:")
			fmt.Println("  gokku run bash -a api")
			fmt.Println("  gokku run rails console -a api")
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
		containerName := app
		dockerCommand := fmt.Sprintf("docker exec -it %s %s", containerName, command)

		fmt.Printf("$ %s\n\n", command)

		cmd := exec.Command("bash", "-c", dockerCommand)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		cmd.Run()
	}
}
