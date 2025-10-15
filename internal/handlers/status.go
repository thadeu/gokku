package handlers

import (
	"fmt"
	"os"
	"os/exec"

	"infra/internal"
)

// handleStatus shows service/container status
func handleStatus(args []string) {
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
		// All services
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

		fmt.Printf("Checking status on %s...\n\n", server.Name)

		sshCmd := fmt.Sprintf(`
			echo "==> Systemd Services"
			for svc in $(ls %s/repos/*.git 2>/dev/null | xargs -n1 basename | sed 's/.git//'); do
				sudo systemctl list-units --all | grep $svc- | awk '{print "  " $1, $3, $4}'
			done
			echo ""
			echo "==> Docker Containers"
			docker ps -a --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" 2>/dev/null | grep -E "production|staging|develop" || echo "  No containers found"
		`, server.BaseDir)

		cmd := exec.Command("ssh", server.Host, sshCmd)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()
		return
	}

	serviceName := fmt.Sprintf("%s-%s", app, env)

	// Check systemd or docker
	sshCmd := fmt.Sprintf(`
		if sudo systemctl list-units --all | grep -q %s; then
			sudo systemctl status %s
		elif docker ps -a | grep -q %s; then
			docker ps -a | grep %s
		else
			echo "Service or container '%s' not found"
			exit 1
		fi
	`, serviceName, serviceName, serviceName, serviceName, serviceName)

	cmd := exec.Command("ssh", "-t", host, sshCmd)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}
