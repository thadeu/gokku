package handlers

import (
	"fmt"
	"os"
	"os/exec"

	"infra/internal"
)

// handleLogs shows application logs
func handleLogs(args []string) {
	remote, remainingArgs := internal.ExtractRemoteFlag(args)

	var app, env, host string
	var follow bool

	if remote != "" {
		remoteInfo, err := internal.GetRemoteInfo(remote)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		app = remoteInfo.App
		env = remoteInfo.Env
		host = remoteInfo.Host

		// Check for -f flag
		for _, arg := range remainingArgs {
			if arg == "-f" {
				follow = true
				break
			}
		}
	} else {
		// Legacy: parse from positional args
		if len(remainingArgs) < 2 {
			fmt.Println("Usage: gokku logs <app> <env> [-f]")
			fmt.Println("   or: gokku logs --remote <git-remote> [-f]")
			os.Exit(1)
		}
		app = remainingArgs[0]
		env = remainingArgs[1]
		follow = len(remainingArgs) > 2 && remainingArgs[2] == "-f"

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
	}

	serviceName := fmt.Sprintf("%s-%s", app, env)
	followFlag := ""
	if follow {
		followFlag = "-f"
	}

	// Try systemd logs first, fallback to docker logs
	sshCmd := fmt.Sprintf(`
		if sudo systemctl list-units --all | grep -q %s; then
			sudo journalctl -u %s %s -n 100
		elif docker ps -a | grep -q %s; then
			docker logs %s %s
		else
			echo "Service or container '%s' not found"
			exit 1
		fi
	`, serviceName, serviceName, followFlag, serviceName, followFlag, serviceName, serviceName)

	cmd := exec.Command("ssh", "-t", host, sshCmd)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Run()
}
