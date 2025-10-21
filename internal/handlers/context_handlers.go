package handlers

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"infra/internal"
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
	// Build docker logs command
	dockerCmd := fmt.Sprintf(`
		if docker ps -a | grep -q %s; then
			docker logs %s %s
		else
			echo "Container '%s' not found"
			exit 1
		fi
	`, serviceName, serviceName, followFlag, serviceName)

	// Execute command
	if err := ctx.ExecuteCommand(dockerCmd); err != nil {
		if !follow {
			os.Exit(1)
		}
	}
}

// handleLogsClientMode handles logs when running from client
func handleLogsClientMode(ctx *internal.ExecutionContext, serviceName, followFlag string, follow bool) {
	// Build docker logs command
	dockerCmd := fmt.Sprintf(`
		if docker ps -a | grep -q %s; then
			docker logs %s %s
		else
			echo "Container '%s' not found"
			exit 1
		fi
	`, serviceName, serviceName, followFlag, serviceName)

	// Execute command with proper signal handling for follow mode
	if follow {
		// For follow mode, use SSH with TTY allocation and proper signal handling
		// Use a more robust approach with proper signal forwarding
		sshCmd := exec.Command("ssh", "-t", "-o", "ServerAliveInterval=30", "-o", "ServerAliveCountMax=3", ctx.Host, dockerCmd)
		sshCmd.Stdout = os.Stdout
		sshCmd.Stderr = os.Stderr
		sshCmd.Stdin = os.Stdin

		if err := sshCmd.Run(); err != nil {
			// Don't exit on signal interruption for follow mode
			if !internal.IsSignalInterruption(err) {
				os.Exit(1)
			}
		}
	} else {
		// For non-follow mode, use regular execution
		if err := ctx.ExecuteCommand(dockerCmd); err != nil {
			os.Exit(1)
		}
	}
}

// handleStatusWithContext shows service/container status using context
func handleStatusWithContext(ctx *internal.ExecutionContext, args []string) {

	// Handle all services status (no specific app)
	if ctx.AppName == "" {
		// Show all services
		dockerCmd := `docker ps -a --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"`

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
		if docker ps -a | grep -q %s; then
			docker ps -a --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" --filter "name=%s"
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

	fmt.Printf("Restarting %s...\n", ctx.GetAppName())

	// Print connection info for remote execution
	ctx.PrintConnectionInfo()

	// Build restart command
	restartCmd := fmt.Sprintf("gokku restart %s", ctx.GetAppName())

	// Execute command
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
		if docker ps -a | grep -q %s; then
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
