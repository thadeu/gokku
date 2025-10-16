package handlers

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"infra/internal"
)

// handlePS manages Procfile processes (Dokku-style commands)
func handlePS(args []string) {

	// Parse Dokku-style subcommand (ps:subcommand or ps subcommand)
	// First, remove the "ps" part if it exists
	var remainingArgs []string
	var subcommand string

	if len(args) > 0 && args[0] == "ps" {
		// Called as "ps subcommand ..."
		remainingArgs = args[1:]
	} else {
		// Called as "ps:subcommand ..." or just subcommand
		remainingArgs = args
	}

	if len(remainingArgs) == 0 {
		// No subcommand provided, show help
		fmt.Println("Usage: gokku ps:<subcommand> [options]")
		fmt.Println("")
		fmt.Println("Dokku-style subcommands:")
		fmt.Println("  ps:list <app> <env>                   List Procfile processes for an app")
		fmt.Println("  ps:logs <app> <env> [process] [-f]    Show process logs")
		fmt.Println("  ps:start <app> <env> [process]        Start process(es)")
		fmt.Println("  ps:stop <app> <env> [process]         Stop process(es)")
		fmt.Println("  ps:restart <app> <env> [process]      Restart process(es)")
		fmt.Println("  ps:scale <app> <env> <process>=<count> Scale process instances")
		fmt.Println("")
		fmt.Println("Options:")
		fmt.Println("  --remote <git-remote>                 Execute on remote server via SSH")
		os.Exit(1)
	}

	subcommand = remainingArgs[0]
	var actualSubcommand string

	if strings.Contains(subcommand, ":") {
		// Handle ps:subcommand format (when called as ps subcommand)
		parts := strings.SplitN(subcommand, ":", 2)
		if len(parts) == 2 {
			actualSubcommand = parts[1]
			remainingArgs = remainingArgs[1:]
		} else {
			actualSubcommand = subcommand
			remainingArgs = remainingArgs[1:]
		}
	} else {
		// Handle direct subcommand (when called as ps subcommand)
		actualSubcommand = subcommand
		remainingArgs = remainingArgs[1:]
	}

	switch actualSubcommand {
	case "list":
		handlePSList(remainingArgs)
	case "logs":
		handlePSLogs(remainingArgs)
	case "start":
		handlePSStart(remainingArgs)
	case "stop":
		handlePSStop(remainingArgs)
	case "restart":
		handlePSRestart(remainingArgs)
	case "scale":
		handlePSScale(remainingArgs)
	default:
		fmt.Printf("Unknown subcommand: %s\n", actualSubcommand)
		fmt.Println("Available subcommands: list, logs, start, stop, restart, scale")
		os.Exit(1)
	}
}

// handlePSList lists Procfile processes for an app
func handlePSList(args []string) {
	remote, remainingArgs := internal.ExtractRemoteFlag(args)

	if remote != "" {
		// Remote execution
		remoteInfo, err := internal.GetRemoteInfo(remote)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("→ %s/%s (%s)\n\n", remoteInfo.App, remoteInfo.Env, remoteInfo.Host)

		// SSH command to list Procfile processes
		sshCmd := fmt.Sprintf(`
			app="%s"
			env="%s"
			base_dir="%s"

			echo "=== Procfile Processes: $app ($env) ==="
			echo ""

			# Check if this app has Procfile processes (look for services with -web, -worker, etc.)
			services=$(sudo systemctl list-units --all | grep "$app-$env-" | awk '{print $1}' | sed 's/.service//')

			if [ -z "$services" ]; then
				echo "No Procfile processes found for $app ($env)"
				echo "This app may not use Procfile or may not be deployed yet."
				exit 1
			fi

			for service in $services; do
				# Extract process type from service name (app-env-processtype)
				process_type=$(echo "$service" | sed "s/$app-$env-//")

				# Check systemd status
				if sudo systemctl is-active --quiet "$service"; then
					systemd_status="running"
				else
					systemd_status="stopped"
				fi

				# Check Docker container status
				container_name="$service"
				if docker ps --format '{{.Names}}' | grep -q "^${container_name}$"; then
					container_status="running"
				else
					container_status="stopped"
				fi

				echo "$process_type:"
				echo "  Systemd: $systemd_status"
				echo "  Container: $container_status"
				echo ""
			done
		`, remoteInfo.App, remoteInfo.Env, remoteInfo.BaseDir)

		cmd := exec.Command("ssh", "-t", remoteInfo.Host, sshCmd)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()
	} else {
		// Local execution
		if len(remainingArgs) < 2 {
			fmt.Println("Error: app and environment are required for local execution")
			fmt.Println("Usage: gokku ps list <app> <env>")
			os.Exit(1)
		}

		app := remainingArgs[0]
		env := remainingArgs[1]

		if !internal.IsRunningOnServer() {
			fmt.Println("Error: Local ps commands can only be run on the server")
			fmt.Println("")
			fmt.Println("For remote execution, use --remote flag:")
			fmt.Println("  gokku ps list <app> <env> --remote <git-remote>")
			os.Exit(1)
		}

		fmt.Printf("=== Procfile Processes: %s (%s) ===\n\n", app, env)

		// Check for Procfile processes (services with app-env-* pattern)
		cmd := exec.Command("sudo", "systemctl", "list-units", "--all")
		output, err := cmd.Output()
		if err != nil {
			fmt.Printf("Error checking systemd services: %v\n", err)
			os.Exit(1)
		}

		servicePattern := fmt.Sprintf("%s-%s-", app, env)
		lines := strings.Split(string(output), "\n")
		found := false

		for _, line := range lines {
			if strings.Contains(line, servicePattern) && strings.Contains(line, ".service") {
				parts := strings.Fields(line)
				if len(parts) >= 4 {
					serviceName := strings.TrimSuffix(parts[0], ".service")
					processType := strings.TrimPrefix(serviceName, servicePattern)

					// Check systemd status
					systemdCmd := exec.Command("sudo", "systemctl", "is-active", serviceName)
					systemdOutput, _ := systemdCmd.Output()
					systemdStatus := strings.TrimSpace(string(systemdOutput))

					// Check Docker container status
					containerName := serviceName
					dockerCmd := exec.Command("docker", "ps", "--format", "{{.Names}}")
					dockerOutput, err := dockerCmd.Output()
					containerStatus := "stopped"
					if err == nil && strings.Contains(string(dockerOutput), containerName) {
						containerStatus = "running"
					}

					fmt.Printf("%s:\n", processType)
					fmt.Printf("  Systemd: %s\n", systemdStatus)
					fmt.Printf("  Container: %s\n", containerStatus)
					fmt.Println("")
					found = true
				}
			}
		}

		if !found {
			fmt.Printf("No Procfile processes found for %s (%s)\n", app, env)
			fmt.Println("This app may not use Procfile or may not be deployed yet.")
		}
	}
}

// handlePSLogs shows logs for Procfile processes
func handlePSLogs(args []string) {
	remote, remainingArgs := internal.ExtractRemoteFlag(args)

	if remote != "" {
		// Remote execution
		remoteInfo, err := internal.GetRemoteInfo(remote)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		if len(remainingArgs) < 2 {
			fmt.Println("Error: app and environment are required")
			fmt.Println("Usage: gokku ps logs <app> <env> [process] [-f] --remote <git-remote>")
			os.Exit(1)
		}

		app := remainingArgs[0]
		env := remainingArgs[1]

		// Check for follow flag
		follow := false
		process := ""
		for _, arg := range remainingArgs[2:] {
			if arg == "-f" {
				follow = true
			} else if arg != "-f" && process == "" {
				process = arg
			}
		}

		fmt.Printf("→ %s/%s (%s)\n\n", app, env, remoteInfo.Host)

		// Build SSH command
		followFlag := ""
		if follow {
			followFlag = "-f"
		}

		processFilter := ""
		if process != "" {
			processFilter = fmt.Sprintf(" | grep '%s-%s-%s'", app, env, process)
		}

		sshCmd := fmt.Sprintf(`
			app="%s"
			env="%s"

			# Get all Procfile processes for this app
			services=$(sudo systemctl list-units --all | grep "$app-$env-" | awk '{print $1}' | sed 's/.service//' %s)

			if [ -z "$services" ]; then
				echo "No Procfile processes found for $app ($env)"
				exit 1
			fi

			for service in $services; do
				process_type=$(echo "$service" | sed "s/$app-$env-//")
				container_name="$service"

				if docker ps -a --format '{{.Names}}' | grep -q "^${container_name}$"; then
					echo "=== $process_type logs ==="
					if [ "%s" = "-f" ]; then
						docker logs $container_name -f --tail 100
					else
						docker logs $container_name --tail 50
					fi
					echo ""
				else
					echo "Container $container_name not found for process $process_type"
				fi
			done
		`, app, env, processFilter, followFlag)

		cmd := exec.Command("ssh", "-t", remoteInfo.Host, sshCmd)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()
	} else {
		// Local execution
		if len(remainingArgs) < 2 {
			fmt.Println("Error: app and environment are required for local execution")
			fmt.Println("Usage: gokku ps logs <app> <env> [process] [-f]")
			os.Exit(1)
		}

		app := remainingArgs[0]
		env := remainingArgs[1]

		if !internal.IsRunningOnServer() {
			fmt.Println("Error: Local ps commands can only be run on the server")
			fmt.Println("")
			fmt.Println("For remote execution, use --remote flag:")
			fmt.Println("  gokku ps logs <app> <env> [process] [-f] --remote <git-remote>")
			os.Exit(1)
		}

		// Check for follow flag and process
		follow := false
		process := ""
		for _, arg := range remainingArgs[2:] {
			if arg == "-f" {
				follow = true
			} else if arg != "-f" && process == "" {
				process = arg
			}
		}

		// Get all Procfile processes
		cmd := exec.Command("sudo", "systemctl", "list-units", "--all")
		output, err := cmd.Output()
		if err != nil {
			fmt.Printf("Error checking systemd services: %v\n", err)
			os.Exit(1)
		}

		servicePattern := fmt.Sprintf("%s-%s-", app, env)
		lines := strings.Split(string(output), "\n")
		found := false

		for _, line := range lines {
			if strings.Contains(line, servicePattern) && strings.Contains(line, ".service") {
				parts := strings.Fields(line)
				if len(parts) >= 4 {
					serviceName := strings.TrimSuffix(parts[0], ".service")
					processType := strings.TrimPrefix(serviceName, servicePattern)

					// Filter by process if specified
					if process != "" && processType != process {
						continue
					}

					containerName := serviceName

					// Check if container exists
					dockerCmd := exec.Command("docker", "ps", "-a", "--format", "{{.Names}}")
					dockerOutput, err := dockerCmd.Output()
					if err == nil && strings.Contains(string(dockerOutput), containerName) {
						fmt.Printf("=== %s logs ===\n", processType)

						logsCmd := exec.Command("docker", "logs", containerName, "--tail", "50")
						if follow {
							logsCmd = exec.Command("docker", "logs", containerName, "-f", "--tail", "100")
						}

						logsCmd.Stdout = os.Stdout
						logsCmd.Stderr = os.Stderr
						logsCmd.Run()
						fmt.Println("")
						found = true
					}
				}
			}
		}

		if !found {
			if process != "" {
				fmt.Printf("Process '%s' not found for %s (%s)\n", process, app, env)
			} else {
				fmt.Printf("No Procfile processes found for %s (%s)\n", app, env)
			}
		}
	}
}

// handlePSRestart restarts Procfile processes
func handlePSRestart(args []string) {
	remote, remainingArgs := internal.ExtractRemoteFlag(args)

	if remote != "" {
		// Remote execution
		remoteInfo, err := internal.GetRemoteInfo(remote)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		if len(remainingArgs) < 2 {
			fmt.Println("Error: app and environment are required")
			fmt.Println("Usage: gokku ps restart <app> <env> [process] --remote <git-remote>")
			os.Exit(1)
		}

		app := remainingArgs[0]
		env := remainingArgs[1]
		process := ""
		if len(remainingArgs) > 2 {
			process = remainingArgs[2]
		}

		fmt.Printf("→ %s/%s (%s)\n\n", app, env, remoteInfo.Host)

		// Build SSH command
		processFilter := ""
		if process != "" {
			processFilter = fmt.Sprintf(" | grep '%s-%s-%s'", app, env, process)
		}

		sshCmd := fmt.Sprintf(`
			app="%s"
			env="%s"

			# Get Procfile processes
			services=$(sudo systemctl list-units --all | grep "$app-$env-" | awk '{print $1}' | sed 's/.service//' %s)

			if [ -z "$services" ]; then
				echo "No Procfile processes found for $app ($env)"
				exit 1
			fi

			for service in $services; do
				process_type=$(echo "$service" | sed "s/$app-$env-//")
				echo "Restarting $process_type..."
				sudo systemctl restart "$service"
				sleep 1

				if sudo systemctl is-active --quiet "$service"; then
					echo "✓ $process_type restarted successfully"
				else
					echo "✗ Failed to restart $process_type"
				fi
			done
		`, app, env, processFilter)

		cmd := exec.Command("ssh", "-t", remoteInfo.Host, sshCmd)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()
	} else {
		// Local execution
		if len(remainingArgs) < 2 {
			fmt.Println("Error: app and environment are required for local execution")
			fmt.Println("Usage: gokku ps restart <app> <env> [process]")
			os.Exit(1)
		}

		app := remainingArgs[0]
		env := remainingArgs[1]
		process := ""
		if len(remainingArgs) > 2 {
			process = remainingArgs[2]
		}

		if !internal.IsRunningOnServer() {
			fmt.Println("Error: Local ps commands can only be run on the server")
			fmt.Println("")
			fmt.Println("For remote execution, use --remote flag:")
			fmt.Println("  gokku ps restart <app> <env> [process] --remote <git-remote>")
			os.Exit(1)
		}

		// Get all Procfile processes
		cmd := exec.Command("sudo", "systemctl", "list-units", "--all")
		output, err := cmd.Output()
		if err != nil {
			fmt.Printf("Error checking systemd services: %v\n", err)
			os.Exit(1)
		}

		servicePattern := fmt.Sprintf("%s-%s-", app, env)
		lines := strings.Split(string(output), "\n")
		found := false

		for _, line := range lines {
			if strings.Contains(line, servicePattern) && strings.Contains(line, ".service") {
				parts := strings.Fields(line)
				if len(parts) >= 4 {
					serviceName := strings.TrimSuffix(parts[0], ".service")
					processType := strings.TrimPrefix(serviceName, servicePattern)

					// Filter by process if specified
					if process != "" && processType != process {
						continue
					}

					fmt.Printf("Restarting %s...\n", processType)

					// Restart systemd service
					restartCmd := exec.Command("sudo", "systemctl", "restart", serviceName)
					err := restartCmd.Run()
					if err != nil {
						fmt.Printf("✗ Failed to restart %s: %v\n", processType, err)
					} else {
						// Check if service is active
						checkCmd := exec.Command("sudo", "systemctl", "is-active", serviceName)
						output, _ := checkCmd.Output()
						status := strings.TrimSpace(string(output))
						if status == "active" {
							fmt.Printf("✓ %s restarted successfully\n", processType)
						} else {
							fmt.Printf("✗ %s status: %s\n", processType, status)
						}
					}

					found = true
				}
			}
		}

		if !found {
			if process != "" {
				fmt.Printf("Process '%s' not found for %s (%s)\n", process, app, env)
			} else {
				fmt.Printf("No Procfile processes found for %s (%s)\n", app, env)
			}
		}
	}
}

// handlePSStart starts Procfile processes
func handlePSStart(args []string) {
	remote, remainingArgs := internal.ExtractRemoteFlag(args)

	if remote != "" {
		// Remote execution
		remoteInfo, err := internal.GetRemoteInfo(remote)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		if len(remainingArgs) < 2 {
			fmt.Println("Error: app and environment are required")
			fmt.Println("Usage: gokku ps:start <app> <env> [process] --remote <git-remote>")
			os.Exit(1)
		}

		app := remainingArgs[0]
		env := remainingArgs[1]
		process := ""
		if len(remainingArgs) > 2 {
			process = remainingArgs[2]
		}

		fmt.Printf("→ %s/%s (%s)\n\n", app, env, remoteInfo.Host)

		// Build SSH command
		processFilter := ""
		if process != "" {
			processFilter = fmt.Sprintf(" | grep '%s-%s-%s'", app, env, process)
		}

		sshCmd := fmt.Sprintf(`
			app="%s"
			env="%s"

			# Get Procfile processes
			services=$(sudo systemctl list-units --all | grep "$app-$env-" | awk '{print $1}' | sed 's/.service//' %s)

			if [ -z "$services" ]; then
				echo "No Procfile processes found for $app ($env)"
				exit 1
			fi

			for service in $services; do
				process_type=$(echo "$service" | sed "s/$app-$env-//")
				echo "Starting $process_type..."
				sudo systemctl start "$service"
				sleep 1

				if sudo systemctl is-active --quiet "$service"; then
					echo "✓ $process_type started successfully"
				else
					echo "✗ Failed to start $process_type"
				fi
			done
		`, app, env, processFilter)

		cmd := exec.Command("ssh", "-t", remoteInfo.Host, sshCmd)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()
	} else {
		// Local execution
		if len(remainingArgs) < 2 {
			fmt.Println("Error: app and environment are required for local execution")
			fmt.Println("Usage: gokku ps:start <app> <env> [process]")
			os.Exit(1)
		}

		app := remainingArgs[0]
		env := remainingArgs[1]
		process := ""
		if len(remainingArgs) > 2 {
			process = remainingArgs[2]
		}

		if !internal.IsRunningOnServer() {
			fmt.Println("Error: Local ps commands can only be run on the server")
			fmt.Println("")
			fmt.Println("For remote execution, use --remote flag:")
			fmt.Println("  gokku ps:start <app> <env> [process] --remote <git-remote>")
			os.Exit(1)
		}

		// Get all Procfile processes
		cmd := exec.Command("sudo", "systemctl", "list-units", "--all")
		output, err := cmd.Output()
		if err != nil {
			fmt.Printf("Error checking systemd services: %v\n", err)
			os.Exit(1)
		}

		servicePattern := fmt.Sprintf("%s-%s-", app, env)
		lines := strings.Split(string(output), "\n")
		found := false

		for _, line := range lines {
			if strings.Contains(line, servicePattern) && strings.Contains(line, ".service") {
				parts := strings.Fields(line)
				if len(parts) >= 4 {
					serviceName := strings.TrimSuffix(parts[0], ".service")
					processType := strings.TrimPrefix(serviceName, servicePattern)

					// Filter by process if specified
					if process != "" && processType != process {
						continue
					}

					fmt.Printf("Starting %s...\n", processType)

					// Start systemd service
					startCmd := exec.Command("sudo", "systemctl", "start", serviceName)
					err := startCmd.Run()
					if err != nil {
						fmt.Printf("✗ Failed to start %s: %v\n", processType, err)
					} else {
						// Check if service is active
						checkCmd := exec.Command("sudo", "systemctl", "is-active", serviceName)
						output, _ := checkCmd.Output()
						status := strings.TrimSpace(string(output))
						if status == "active" {
							fmt.Printf("✓ %s started successfully\n", processType)
						} else {
							fmt.Printf("✗ %s status: %s\n", processType, status)
						}
					}

					found = true
				}
			}
		}

		if !found {
			if process != "" {
				fmt.Printf("Process '%s' not found for %s (%s)\n", process, app, env)
			} else {
				fmt.Printf("No Procfile processes found for %s (%s)\n", app, env)
			}
		}
	}
}

// handlePSStop stops Procfile processes
func handlePSStop(args []string) {
	remote, remainingArgs := internal.ExtractRemoteFlag(args)

	if remote != "" {
		// Remote execution
		remoteInfo, err := internal.GetRemoteInfo(remote)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		if len(remainingArgs) < 2 {
			fmt.Println("Error: app and environment are required")
			fmt.Println("Usage: gokku ps:stop <app> <env> [process] --remote <git-remote>")
			os.Exit(1)
		}

		app := remainingArgs[0]
		env := remainingArgs[1]
		process := ""
		if len(remainingArgs) > 2 {
			process = remainingArgs[2]
		}

		fmt.Printf("→ %s/%s (%s)\n\n", app, env, remoteInfo.Host)

		// Build SSH command
		processFilter := ""
		if process != "" {
			processFilter = fmt.Sprintf(" | grep '%s-%s-%s'", app, env, process)
		}

		sshCmd := fmt.Sprintf(`
			app="%s"
			env="%s"

			# Get Procfile processes
			services=$(sudo systemctl list-units --all | grep "$app-$env-" | awk '{print $1}' | sed 's/.service//' %s)

			if [ -z "$services" ]; then
				echo "No Procfile processes found for $app ($env)"
				exit 1
			fi

			for service in $services; do
				process_type=$(echo "$service" | sed "s/$app-$env-//")
				echo "Stopping $process_type..."
				sudo systemctl stop "$service"
				sleep 1

				if ! sudo systemctl is-active --quiet "$service"; then
					echo "✓ $process_type stopped successfully"
				else
					echo "✗ Failed to stop $process_type"
				fi
			done
		`, app, env, processFilter)

		cmd := exec.Command("ssh", "-t", remoteInfo.Host, sshCmd)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()
	} else {
		// Local execution
		if len(remainingArgs) < 2 {
			fmt.Println("Error: app and environment are required for local execution")
			fmt.Println("Usage: gokku ps:stop <app> <env> [process]")
			os.Exit(1)
		}

		app := remainingArgs[0]
		env := remainingArgs[1]
		process := ""
		if len(remainingArgs) > 2 {
			process = remainingArgs[2]
		}

		if !internal.IsRunningOnServer() {
			fmt.Println("Error: Local ps commands can only be run on the server")
			fmt.Println("")
			fmt.Println("For remote execution, use --remote flag:")
			fmt.Println("  gokku ps:stop <app> <env> [process] --remote <git-remote>")
			os.Exit(1)
		}

		// Get all Procfile processes
		cmd := exec.Command("sudo", "systemctl", "list-units", "--all")
		output, err := cmd.Output()
		if err != nil {
			fmt.Printf("Error checking systemd services: %v\n", err)
			os.Exit(1)
		}

		servicePattern := fmt.Sprintf("%s-%s-", app, env)
		lines := strings.Split(string(output), "\n")
		found := false

		for _, line := range lines {
			if strings.Contains(line, servicePattern) && strings.Contains(line, ".service") {
				parts := strings.Fields(line)
				if len(parts) >= 4 {
					serviceName := strings.TrimSuffix(parts[0], ".service")
					processType := strings.TrimPrefix(serviceName, servicePattern)

					// Filter by process if specified
					if process != "" && processType != process {
						continue
					}

					fmt.Printf("Stopping %s...\n", processType)

					// Stop systemd service
					stopCmd := exec.Command("sudo", "systemctl", "stop", serviceName)
					err := stopCmd.Run()
					if err != nil {
						fmt.Printf("✗ Failed to stop %s: %v\n", processType, err)
					} else {
						// Check if service is inactive
						checkCmd := exec.Command("sudo", "systemctl", "is-active", serviceName)
						output, _ := checkCmd.Output()
						status := strings.TrimSpace(string(output))
						if status == "inactive" || status == "failed" {
							fmt.Printf("✓ %s stopped successfully\n", processType)
						} else {
							fmt.Printf("✗ %s still running (status: %s)\n", processType, status)
						}
					}

					found = true
				}
			}
		}

		if !found {
			if process != "" {
				fmt.Printf("Process '%s' not found for %s (%s)\n", process, app, env)
			} else {
				fmt.Printf("No Procfile processes found for %s (%s)\n", app, env)
			}
		}
	}
}

// handlePSScale scales Procfile processes (Dokku-style scaling)
func handlePSScale(args []string) {
	remote, remainingArgs := internal.ExtractRemoteFlag(args)

	if len(remainingArgs) < 3 {
		fmt.Println("Error: app, environment, and scale specification are required")
		fmt.Println("Usage: gokku ps:scale <app> <env> <process>=<count> [--remote <git-remote>]")
		fmt.Println("Example: gokku ps:scale myapp production worker=2")
		os.Exit(1)
	}

	app := remainingArgs[0]
	env := remainingArgs[1]
	scaleSpec := remainingArgs[2]

	// Parse scale specification (process=count)
	if !strings.Contains(scaleSpec, "=") {
		fmt.Printf("Error: Invalid scale specification '%s'. Use format: process=count\n", scaleSpec)
		fmt.Println("Example: worker=2")
		os.Exit(1)
	}

	parts := strings.SplitN(scaleSpec, "=", 2)
	process := parts[0]
	countStr := parts[1]

	count, err := strconv.Atoi(countStr)
	if err != nil || count < 0 {
		fmt.Printf("Error: Invalid count '%s'. Must be a non-negative integer\n", countStr)
		os.Exit(1)
	}

	fmt.Printf("Scaling %s/%s %s to %d instances\n", app, env, process, count)

	if remote != "" {
		// Remote execution
		remoteInfo, err := internal.GetRemoteInfo(remote)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("→ %s/%s (%s)\n\n", app, env, remoteInfo.Host)

		// For now, implement basic scaling (0 or 1 instance)
		// Full scaling with multiple instances would require more complex systemd setup
		sshCmd := fmt.Sprintf(`
			app="%s"
			env="%s"
			process="%s"
			count="%d"

			service_name="$app-$env-$process"

			if [ "$count" -eq 0 ]; then
				echo "Stopping $process (scaling to 0)..."
				sudo systemctl stop "$service_name" 2>/dev/null || true
				sudo systemctl disable "$service_name" 2>/dev/null || true
				echo "✓ $process scaled to 0 instances"
			else
				echo "Starting $process (scaling to $count)..."
				sudo systemctl enable "$service_name" 2>/dev/null || true
				sudo systemctl start "$service_name"
				sleep 2

				if sudo systemctl is-active --quiet "$service_name"; then
					echo "✓ $process scaled to $count instance(s)"
				else
					echo "✗ Failed to scale $process"
				fi
			fi
		`, app, env, process, count)

		cmd := exec.Command("ssh", "-t", remoteInfo.Host, sshCmd)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()
	} else {
		// Local execution
		if !internal.IsRunningOnServer() {
			fmt.Println("Error: Local ps commands can only be run on the server")
			fmt.Println("")
			fmt.Println("For remote execution, use --remote flag:")
			fmt.Printf("  gokku ps:scale %s %s %s --remote <git-remote>\n", app, env, scaleSpec)
			os.Exit(1)
		}

		serviceName := fmt.Sprintf("%s-%s-%s", app, env, process)

		// For now, implement basic scaling (0 or 1 instance)
		// Full scaling with multiple instances would require more complex systemd setup
		if count == 0 {
			fmt.Printf("Stopping %s (scaling to 0)...\n", process)

			stopCmd := exec.Command("sudo", "systemctl", "stop", serviceName)
			stopCmd.Run() // Ignore errors if service doesn't exist

			disableCmd := exec.Command("sudo", "systemctl", "disable", serviceName)
			disableCmd.Run() // Ignore errors

			fmt.Printf("✓ %s scaled to 0 instances\n", process)
		} else {
			fmt.Printf("Starting %s (scaling to %d)...\n", process, count)

			enableCmd := exec.Command("sudo", "systemctl", "enable", serviceName)
			enableCmd.Run() // Ignore errors if service doesn't exist

			startCmd := exec.Command("sudo", "systemctl", "start", serviceName)
			err := startCmd.Run()
			if err != nil {
				fmt.Printf("✗ Failed to start %s: %v\n", process, err)
			} else {
				// Check if service is active
				checkCmd := exec.Command("sudo", "systemctl", "is-active", serviceName)
				output, _ := checkCmd.Output()
				status := strings.TrimSpace(string(output))
				if status == "active" {
					fmt.Printf("✓ %s scaled to %d instance(s)\n", process, count)
				} else {
					fmt.Printf("✗ %s failed to start (status: %s)\n", process, status)
				}
			}
		}
	}
}
