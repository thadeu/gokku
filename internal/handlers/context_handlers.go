package handlers

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"gokku/internal"
)

// handleLogsWithContext shows application logs using context
func handleLogsWithContext(ctx *internal.ExecutionContext, args []string) {
	// Validate that app is required
	if err := ctx.ValidateAppRequired(); err != nil {
		ctx.PrintUsageError("logs", err.Error())
	}

	// Extract remaining args (without -a flag)
	_, remainingArgs := internal.ExtractAppFlag(args)

	// Check for -f flag
	follow := false
	for _, arg := range remainingArgs {
		if arg == "-f" {
			follow = true
			break
		}
	}

	serviceName := ctx.GetAppName()
	followFlag := ""

	if follow {
		followFlag = "-f"
	}

	// Print connection info for remote execution
	ctx.PrintConnectionInfo()

	// Check if we're running locally on server or remotely
	if ctx.ServerExecution {
		// Server mode - execute docker logs directly
		handleLogsServerMode(ctx, serviceName, followFlag, follow)
	} else {
		// Client mode - execute via SSH with proper signal handling
		handleLogsClientMode(ctx, serviceName, followFlag, follow)
	}
}

// handleLogsServerMode handles logs when running on server
func handleLogsServerMode(ctx *internal.ExecutionContext, serviceName, followFlag string, follow bool) {
	// Check if container exists
	checkCmd := exec.Command("docker", "ps", "-a", "--format", "{{.Names}}")
	output, err := checkCmd.Output()
	if err != nil {
		fmt.Printf("Error checking containers: %v\n", err)
		os.Exit(1)
	}

	containerExists := false
	for _, name := range strings.Split(string(output), "\n") {
		if strings.TrimSpace(name) == serviceName {
			containerExists = true
			break
		}
	}

	if !containerExists {
		fmt.Printf("Container '%s' not found\n", serviceName)
		os.Exit(1)
	}

	// Execute docker logs directly on server
	var cmd *exec.Cmd
	if follow {
		cmd = exec.Command("docker", "logs", "-f", "--tail", "500", serviceName)
	} else {
		cmd = exec.Command("docker", "logs", "--tail", "500", serviceName)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	// Run the command and check for errors
	if err := cmd.Run(); err != nil {
		// Check if it's a signal interruption (Ctrl+C), which is normal
		if exitErr, ok := err.(*exec.ExitError); ok {
			// Exit code 130 = SIGINT (Ctrl+C)
			// Exit code 143 = SIGTERM
			if exitErr.ExitCode() == 130 || exitErr.ExitCode() == 143 {
				os.Exit(0)
			}
		}
		fmt.Printf("Error executing docker logs: %v\n", err)
		os.Exit(1)
	}
}

// handleLogsClientMode handles logs when running from client
func handleLogsClientMode(ctx *internal.ExecutionContext, serviceName, followFlag string, follow bool) {
	// Execute via SSH
	sshCmd := fmt.Sprintf(`
  if docker ps | grep -q %s; then
    docker logs %s %s
  else
    echo "Container '%s' not found"
    exit 1
  fi
`, serviceName, serviceName, followFlag, serviceName)

	cmd := exec.Command("ssh", "-t", ctx.Host, sshCmd)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Run()
}

// handleStatusWithContext shows service/container status using context
func handleStatusWithContext(ctx *internal.ExecutionContext, args []string) {
	// Handle all services status (no specific app)
	if ctx.AppName == "" {
		// Show all services
		dockerCmd := `docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"`

		if ctx.ServerExecution {
			fmt.Println("==> Docker Containers")
		} else {
			ctx.PrintConnectionInfo()
		}

		if err := ctx.ExecuteCommand(dockerCmd); err != nil {
			os.Exit(1)
		}
		return
	}

	// Handle specific app status
	serviceName := ctx.GetAppName()

	// Print connection info for remote execution
	ctx.PrintConnectionInfo()

	// Build docker status command
	dockerCmd := fmt.Sprintf(`
		if docker ps | grep -q %s; then
			docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" --filter "name=%s"
		else
			echo "Service or container '%s' not found"
			exit 1
		fi
	`, serviceName, serviceName, serviceName)

	// Execute command
	if err := ctx.ExecuteCommand(dockerCmd); err != nil {
		os.Exit(1)
	}
}

// handleRestartWithContext restarts services/containers using context
func handleRestartWithContext(ctx *internal.ExecutionContext, args []string) {
	// Validate that app is required
	if err := ctx.ValidateAppRequired(); err != nil {
		ctx.PrintUsageError("restart", err.Error())
	}

	appName := ctx.GetAppName()
	fmt.Printf("Restarting %s...\n", appName)

	// Print connection info for remote execution
	ctx.PrintConnectionInfo()

	// Check if we're running locally on server or remotely
	if ctx.ServerExecution {
		// Server mode - recreate container with updated env file
		executeRestartServerMode(ctx, appName)
	} else {
		// Client mode - execute via SSH
		executeRestartClientMode(ctx, appName)
	}
}

// executeRestartServerMode recreates container on server with updated environment
func executeRestartServerMode(ctx *internal.ExecutionContext, appName string) {
	envFile := fmt.Sprintf("%s/apps/%s/shared/.env", ctx.BaseDir, appName)
	appDir := fmt.Sprintf("%s/apps/%s/current", ctx.BaseDir, appName)

	// Use RecreateActiveContainer to properly reload env vars
	if err := internal.RecreateActiveContainer(appName, envFile, appDir); err != nil {
		fmt.Printf("Error restarting app: %v\n", err)
		os.Exit(1)
	}
}

// executeRestartClientMode executes restart via SSH
func executeRestartClientMode(ctx *internal.ExecutionContext, appName string) {
	// Call gokku restart on the server (which will use RecreateActiveContainer)
	restartCmd := fmt.Sprintf("gokku restart -a %s", appName)

	if err := ctx.ExecuteCommand(restartCmd); err != nil {
		fmt.Printf("Error restarting app: %v\n", err)
		os.Exit(1)
	}
}

// handleRollbackWithContext rolls back to a previous release using context
func handleRollbackWithContext(ctx *internal.ExecutionContext, args []string) {
	// Validate that app is required
	if err := ctx.ValidateAppRequired(); err != nil {
		ctx.PrintUsageError("rollback", err.Error())
	}

	// Extract remaining args (without -a flag)
	_, remainingArgs := internal.ExtractAppFlag(args)

	var releaseID string
	if len(remainingArgs) > 0 {
		releaseID = remainingArgs[0]
	}

	appDir := fmt.Sprintf("%s/apps/%s", ctx.BaseDir, ctx.GetAppName())
	serviceName := ctx.GetAppName()

	if releaseID == "" {
		// Get previous release
		listCmd := fmt.Sprintf("cd %s/releases && ls -t | sed -n '2p'", appDir)
		output, err := ctx.ExecuteCommandWithOutput(listCmd)
		if err != nil {
			fmt.Printf("Failed to get releases: %v\n", err)
			os.Exit(1)
		}
		releaseID = strings.TrimSpace(output)
	}

	if releaseID == "" {
		fmt.Println("No previous release found")
		os.Exit(1)
	}

	fmt.Printf("Rolling back %s to release: %s\n", ctx.GetAppName(), releaseID)

	// Print connection info for remote execution
	ctx.PrintConnectionInfo()

	// Build rollback command
	rollbackCmd := fmt.Sprintf(`
		cd %s && \
		if docker ps | grep -q %s; then
			docker stop %s && \
			docker rm -f %s && \
			docker run -d --name %s --env-file %s/shared/.env %s:release-%s && \
			docker start %s && \
			echo "✓ Rollback complete"
		else
			echo "Error: Service or container not found"
			exit 1
		fi
	`, appDir, serviceName, serviceName, serviceName, serviceName, appDir, ctx.GetAppName(), releaseID, serviceName)

	// Execute command
	if err := ctx.ExecuteCommand(rollbackCmd); err != nil {
		fmt.Printf("Rollback failed: %v\n", err)
		os.Exit(1)
	}
}

// handleRunWithContext executes arbitrary commands using context
func handleRunWithContext(ctx *internal.ExecutionContext, args []string) {
	if err := ctx.ValidateAppRequired(); err != nil {
		ctx.PrintUsageError("run", err.Error())
	}

	_, remainingArgs := internal.ExtractAppFlag(args)

	if len(remainingArgs) < 1 {
		fmt.Println("Error: command is required")
		fmt.Println("Usage: gokku run <command> -a <app>")
		os.Exit(1)
	}

	command := strings.Join(remainingArgs, " ")

	containerName := ctx.GetAppName()
	dockerCommand := fmt.Sprintf("docker exec -it %s %s", containerName, command)

	ctx.PrintConnectionInfo()
	fmt.Printf("$ %s\n\n", command)

	// Execute command
	if err := ctx.ExecuteCommand(dockerCommand); err != nil {
		os.Exit(1)
	}
}
