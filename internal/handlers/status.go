package handlers

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"infra/internal"
)

// handleStatus shows service/container status
func handleStatus(args []string) {
	remote, remainingArgs := internal.ExtractRemoteFlag(args)

	var app, env, host, baseDir string
	var localExecution bool

	if remote != "" {
		remoteInfo, err := internal.GetRemoteInfo(remote)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		app = remoteInfo.App
		env = remoteInfo.Env
		host = remoteInfo.Host
		baseDir = remoteInfo.BaseDir
	} else if len(remainingArgs) >= 2 {
		// Check if running on server - allow local execution
		if internal.IsRunningOnServer() {
			localExecution = true
			app = remainingArgs[0]
			env = remainingArgs[1]
			baseDir = "/opt/gokku"
		} else {
			// Client without --remote - show error
			fmt.Println("Error: Local status commands can only be run on the server")
			fmt.Println("")
			fmt.Println("For client usage, use --remote flag:")
			fmt.Println("  gokku status [app] [env] --remote <git-remote>")
			fmt.Println("")
			fmt.Println("Or run this command directly on your server.")
			os.Exit(1)
		}
	} else {
		// All services - works from both client and server
		if internal.IsRunningOnServer() {
			localExecution = true
			baseDir = "/opt/gokku"
			fmt.Printf("Checking status on local server...\n\n")
		} else {
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
			host = server.Host
			baseDir = server.BaseDir
		}

		if localExecution {
			// Local execution for all services
			fmt.Println("==> Docker Containers")
			dockerCmd := exec.Command("docker", "ps", "-a", "--format", "table {{.Names}}\t{{.Status}}\t{{.Ports}}")
			output, err := dockerCmd.Output()
			if err == nil {
				lines := strings.Split(string(output), "\n")
				for _, line := range lines {
					if strings.Contains(line, "-production") || strings.Contains(line, "-staging") || strings.Contains(line, "-develop") {
						parts := strings.Fields(line)
						if len(parts) >= 4 {
							fmt.Printf("  %s %s %s\n", parts[0], parts[2], parts[3])
						}
					}
				}
			}
			fmt.Println("")
			fmt.Println("==> Docker Containers")
			dockerCmd = exec.Command("docker", "ps", "-a", "--format", "table {{.Names}}\t{{.Status}}\t{{.Ports}}")
			output, err = dockerCmd.Output()
			if err == nil {
				lines := strings.Split(string(output), "\n")
				found := false
				for _, line := range lines {
					if strings.Contains(line, "production") || strings.Contains(line, "staging") || strings.Contains(line, "develop") {
						fmt.Println(" ", line)
						found = true
					}
				}
				if !found {
					fmt.Println("  No containers found")
				}
			} else {
				fmt.Println("  Docker not available")
			}
		} else {
			// Remote execution for all services
			sshCmd := fmt.Sprintf(`
				echo "==> Docker Containers"
				for svc in $(ls %s/repos/*.git 2>/dev/null | xargs -n1 basename | sed 's/.git//'); do
					docker ps -a --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" 2>/dev/null | grep $svc- | awk '{print "  " $1, $3, $4}'
				done
				echo ""
			`, baseDir)

			cmd := exec.Command("ssh", host, sshCmd)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Run()
		}
		return
	}

	serviceName := fmt.Sprintf("%s-%s", app, env)

	if localExecution {
		// Local execution for specific app/env
		var cmd *exec.Cmd

		dockerCmd := exec.Command("docker", "ps", "-a")
		output, err := dockerCmd.Output()
		if err == nil && strings.Contains(string(output), serviceName) {
			cmd = exec.Command("docker", "ps", "-a", "--format", "table {{.Names}}\t{{.Status}}\t{{.Ports}}", "--filter", fmt.Sprintf("name=%s", serviceName))
		} else {
			// Try docker
			fmt.Printf("Service or container '%s' not found\n", serviceName)
			os.Exit(1)
		}

		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		cmd.Run()
	} else {
		// Remote execution for specific app/env
		sshCmd := fmt.Sprintf(`
			if docker ps -a | grep -q %s; then
				docker ps -a --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" --filter "name=%s"
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
}
