package handlers

import (
	"fmt"
	"os"
	"os/exec"

	"infra/internal"
)

// handleRestart restarts services/containers
func handleRestart(args []string) {
	remote, remainingArgs := internal.ExtractRemoteFlag(args)

	var app, env, host string

	if remote != "" {
		remoteInfo, err := internal.GetRemoteInfo(remote)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		app = remoteInfo.App
		env = remoteInfo.Env
		host = remoteInfo.Host
	} else if len(remainingArgs) >= 2 {
		app = remainingArgs[0]
		env = remainingArgs[1]

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
	} else {
		fmt.Println("Usage: gokku restart <app> <env>")
		fmt.Println("   or: gokku restart --remote <git-remote>")
		os.Exit(1)
	}

	serviceName := fmt.Sprintf("%s-%s", app, env)
	fmt.Printf("Restarting %s...\n", serviceName)

	// Check systemd or docker and restart accordingly
	sshCmd := fmt.Sprintf(`
		if sudo systemctl list-units --all | grep -q %s; then
			sudo systemctl restart %s && echo "✓ Service restarted"
		elif docker ps -a | grep -q %s; then
			docker restart %s && echo "✓ Container restarted"
		else
			echo "Error: Service or container '%s' not found"
			exit 1
		fi
	`, serviceName, serviceName, serviceName, serviceName, serviceName)

	cmd := exec.Command("ssh", host, sshCmd)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}
