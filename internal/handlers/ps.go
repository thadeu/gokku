package handlers

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"infra/internal"
	"infra/internal/containers"
)

// HandlePS handles ps-related commands
func HandlePS(args []string) {
	if len(args) == 0 {
		// List all containers when no arguments provided
		listAllContainers()
		return
	}

	subcommand := args[0]

	// Check if first argument is a flag (-a or --app), treat as list command
	if subcommand == "-a" || subcommand == "--app" {
		appName, _ := internal.ExtractAppFlag(args)
		if appName == "" {
			fmt.Println("Error: App name is required")
			fmt.Println("Usage: gokku ps -a <app>")
			fmt.Println("   or: gokku ps:list -a <app>")
			os.Exit(1)
		}

		// Check if we're in client mode - if so, execute remotely
		isClientMode := internal.IsClientMode()
		if isClientMode {
			// Get remote info to execute command remotely
			remoteInfo, err := internal.GetRemoteInfo(appName)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}

			// Execute ps:list command remotely
			cmd := fmt.Sprintf("gokku ps:list --app %s", remoteInfo.App)
			sshCmd := exec.Command("ssh", remoteInfo.Host, cmd)
			sshCmd.Stdout = os.Stdout
			sshCmd.Stderr = os.Stderr
			sshCmd.Stdin = os.Stdin
			if err := sshCmd.Run(); err != nil {
				os.Exit(1)
			}
			return
		}

		// Server mode - execute locally
		handlePSList([]string{"-a", appName})
		return
	}

	if subcommand == "" {
		appName, _ := internal.ExtractAppFlag(args)
		handlePSList([]string{"-a", appName})
		return
	}

	switch subcommand {
	case "scale":
		handlePSScale(args[1:])
	case "report", "rs", "list":
		handlePSList(args[1:])
	case "restart":
		handlePSRestart(args[1:])
	case "stop":
		handlePSStop(args[1:])
	default:
		fmt.Printf("Unknown ps command: %s\n", subcommand)
		showPSHelp()
		os.Exit(1)
	}
}

// extractAppNameForPS extracts app name from arguments based on server/client mode
// Server mode: accepts app name as direct argument (e.g., gokku ps:list APP_NAME)
// Client mode: requires -a flag (e.g., gokku ps:list -a APP_NAME)
func extractAppNameForPS(args []string) string {
	isServerMode := internal.IsServerMode()

	if isServerMode {
		// Server mode: look for app name as direct argument (not a flag)
		// For scale command, app name is typically the last non-scale argument
		// For other commands, it's the first/only non-flag argument
		// Skip flags and scale arguments (e.g., web=4)
		for i := len(args) - 1; i >= 0; i-- {
			arg := args[i]
			if !strings.HasPrefix(arg, "-") && !strings.Contains(arg, "=") {
				return arg
			}
		}
		return ""
	} else {
		// Client mode: require -a flag
		return internal.ExtractAppName(args)
	}
}

// handlePSScale handles the ps:scale command
func handlePSScale(args []string) {
	// Parse: gokku ps:scale web=4 worker=2 -a api (client)
	//        gokku ps:scale web=4 worker=2 APP_NAME (server)
	appName := extractAppNameForPS(args)
	if appName == "" {
		isServerMode := internal.IsServerMode()
		if isServerMode {
			fmt.Println("Error: App name is required")
			fmt.Println("Usage: gokku ps:scale <process=count>... <app>")
			fmt.Println("")
			fmt.Println("Examples:")
			fmt.Println("  gokku ps:scale web=4 worker=2 api")
			os.Exit(1)
		} else {
			fmt.Println("Error: -a <app> is required")
			fmt.Println("Usage: gokku ps:scale <process=count>... -a <app>")
			fmt.Println("")
			fmt.Println("Examples:")
			fmt.Println("  gokku ps:scale web=4 worker=2 -a api-production")
			os.Exit(1)
		}
	}

	isServerMode := internal.IsServerMode()

	// Parse scale arguments
	scales := make(map[string]int)
	for _, arg := range args {
		if strings.HasPrefix(arg, "-a") {
			continue // Skip app flag (client mode)
		}

		// Skip app name (server mode - it's the last non-scale argument)
		if isServerMode && arg == appName {
			continue
		}

		processType, count, err := containers.ParseScaleArgument(arg)
		if err != nil {
			fmt.Printf("Error parsing scale argument '%s': %v\n", arg, err)
			os.Exit(1)
		}

		scales[processType] = count
	}

	if len(scales) == 0 {
		fmt.Println("Error: No scale arguments provided (e.g., web=4 worker=2)")
		os.Exit(1)
	}

	fmt.Printf("Scaling app '%s'...\n", appName)

	registry := containers.NewContainerRegistry()

	for processType, count := range scales {
		fmt.Printf("-----> Scaling %s to %d instances\n", processType, count)

		// Get current containers
		currentContainers, err := registry.GetContainers(appName, processType)
		if err != nil {
			fmt.Printf("Error getting current containers: %v\n", err)
			continue
		}

		currentCount := len(currentContainers)

		if count > currentCount {
			// Scale up
			scaleUp(appName, processType, count-currentCount, registry)
		} else if count < currentCount {
			// Scale down
			scaleDown(appName, processType, currentCount-count, registry)
		} else {
			fmt.Printf("       %s already at %d instances\n", processType, count)
		}

		// Notify plugins about scale change
		notifyPluginsOfScaleChange(appName, processType)
	}

	fmt.Printf("Scaling complete for app '%s'\n", appName)
}

// scaleUp creates new containers
func scaleUp(appName, processType string, count int, registry *containers.ContainerRegistry) {
	for i := 0; i < count; i++ {
		// Get next container number
		containerNum := registry.GetNextContainerNumber(appName, processType)
		containerName := fmt.Sprintf("%s-%s-%d", appName, processType, containerNum)

		// Get dynamic port
		hostPort, err := containers.GetNextAvailablePort()
		if err != nil {
			fmt.Printf("       Error getting port for %s: %v\n", containerName, err)
			continue
		}

		// Create container
		fmt.Printf("       Creating container: %s (port %d)\n", containerName, hostPort)

		// Check if app image exists
		appImage := fmt.Sprintf("%s:latest", appName)
		if !imageExists(appImage) {
			fmt.Printf("       Error: Image %s not found. Deploy the app first.\n", appImage)
			continue
		}

		// Create container with docker run
		cmd := exec.Command("docker", "run", "-d",
			"--name", containerName,
			"-p", fmt.Sprintf("%d", hostPort),
			"--env-file", fmt.Sprintf("/opt/gokku/apps/%s/.env", appName),
			"--ulimit", "nofile=65536:65536",
			"--ulimit", "nproc=4096:4096",
			appImage)

		if err := cmd.Run(); err != nil {
			fmt.Printf("       Error creating container %s: %v\n", containerName, err)
			continue
		}

		// Save container info
		info := containers.CreateContainerInfo(appName, processType, containerNum, hostPort, hostPort)
		if err := registry.SaveContainerInfo(info); err != nil {
			fmt.Printf("       Warning: Failed to save container info: %v\n", err)
		}
	}
}

// scaleDown removes containers
func scaleDown(appName, processType string, count int, registry *containers.ContainerRegistry) {
	containers, err := registry.GetContainers(appName, processType)
	if err != nil {
		fmt.Printf("       Error getting containers: %v\n", err)
		return
	}

	// Remove last N containers (highest numbers first)
	toRemove := count
	if toRemove > len(containers) {
		toRemove = len(containers)
	}

	for i := len(containers) - toRemove; i < len(containers); i++ {
		container := containers[i]
		containerName := container.Name

		fmt.Printf("       Stopping container: %s\n", containerName)

		// Stop and remove container
		stopCmd := exec.Command("docker", "stop", containerName)
		if err := stopCmd.Run(); err != nil {
			fmt.Printf("       Warning: Failed to stop container %s: %v\n", containerName, err)
		}

		rmCmd := exec.Command("docker", "rm", containerName)
		if err := rmCmd.Run(); err != nil {
			fmt.Printf("       Warning: Failed to remove container %s: %v\n", containerName, err)
		}

		// Remove container info
		if err := registry.RemoveContainerInfo(appName, processType, container.Number); err != nil {
			fmt.Printf("       Warning: Failed to remove container info: %v\n", err)
		}
	}
}

// handlePSList handles the ps:list command
func handlePSList(args []string) {
	// Parse: gokku ps:list -a api (client)
	//        gokku ps:list APP_NAME (server)
	//        gokku ps:list (server, list all)
	appName := extractAppNameForPS(args)
	if appName == "" {
		isServerMode := internal.IsServerMode()
		if isServerMode {
			// In server mode, list all containers when no app name provided
			listAllContainers()
			return
		} else {
			fmt.Println("Error: -a <app> is required")
			fmt.Println("Usage: gokku ps:list -a <app>")
			fmt.Println("")
			fmt.Println("Examples:")
			fmt.Println("  gokku ps:list -a api-production")
			os.Exit(1)
		}
	}

	// Check if we're in client mode - if so, execute remotely
	isClientMode := internal.IsClientMode()
	if isClientMode {
		// Get remote name from args (the -a flag value)
		remoteName, _ := internal.ExtractAppFlag(args)
		if remoteName == "" {
			fmt.Println("Error: -a <app> is required")
			fmt.Println("Usage: gokku ps:list -a <app>")
			os.Exit(1)
		}

		remoteInfo, err := internal.GetRemoteInfo(remoteName)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		// Execute ps:list command remotely using the app name from remote info
		cmd := fmt.Sprintf("gokku ps:list --app %s", remoteInfo.App)
		sshCmd := exec.Command("ssh", remoteInfo.Host, cmd)
		sshCmd.Stdout = os.Stdout
		sshCmd.Stderr = os.Stderr
		sshCmd.Stdin = os.Stdin
		if err := sshCmd.Run(); err != nil {
			os.Exit(1)
		}
		return
	}

	// Use docker ps to get running containers with pipe-separated format for easier parsing
	cmd := exec.Command("docker", "ps", "--format", "{{.Names}}|{{.Status}}|{{.Ports}}")
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("Error running docker ps: %v\n", err)
		os.Exit(1)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 0 || (len(lines) == 1 && lines[0] == "") {
		fmt.Printf("No processes running for app '%s'\n", appName)
		return
	}

	// Filter containers that belong to this app
	var appContainers [][]string
	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.Split(line, "|")
		if len(parts) >= 3 {
			name := strings.TrimSpace(parts[0])
			if strings.HasPrefix(name, appName+"-") || name == appName {
				appContainers = append(appContainers, parts)
			}
		}
	}

	if len(appContainers) == 0 {
		fmt.Printf("No processes running for app '%s'\n", appName)
		return
	}

	fmt.Printf("Processes for app '%s':\n", appName)
	fmt.Printf("%-20s %-15s %s\n", "NAME", "STATUS", "PORT")
	fmt.Printf("%-20s %-15s %s\n", "----", "------", "----")

	for _, parts := range appContainers {
		name := strings.TrimSpace(parts[0])
		status := strings.TrimSpace(parts[1])
		ports := strings.TrimSpace(parts[2])

		fmt.Printf("%-20s %-15s %s\n", name, status, ports)
	}
}

// listAllContainers lists all running containers with gokku format
func listAllContainers() {
	// Use docker ps to get running containers with pipe-separated format for easier parsing
	cmd := exec.Command("docker", "ps", "--format", "{{.Names}}|{{.Status}}|{{.Ports}}")
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("Error running docker ps: %v\n", err)
		os.Exit(1)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 0 || (len(lines) == 1 && lines[0] == "") {
		fmt.Println("No processes running")
		return
	}

	fmt.Printf("%-20s %-15s %s\n", "NAME", "STATUS", "PORT")
	fmt.Printf("%-20s %-15s %s\n", "----", "------", "----")

	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.Split(line, "|")
		if len(parts) >= 3 {
			name := strings.TrimSpace(parts[0])
			status := strings.TrimSpace(parts[1])
			ports := strings.TrimSpace(parts[2])

			fmt.Printf("%-20s %-15s %s\n", name, status, ports)
		}
	}
}

// handlePSRestart handles the ps:restart command
func handlePSRestart(args []string) {
	// Parse: gokku ps:restart -a api (client)
	//        gokku ps:restart APP_NAME (server)
	appName := extractAppNameForPS(args)
	if appName == "" {
		isServerMode := internal.IsServerMode()
		if isServerMode {
			fmt.Println("Error: App name is required")
			fmt.Println("Usage: gokku ps:restart <app>")
			fmt.Println("")
			fmt.Println("Examples:")
			fmt.Println("  gokku ps:restart api")
			os.Exit(1)
		} else {
			fmt.Println("Error: -a <app> is required")
			fmt.Println("Usage: gokku ps:restart -a <app>")
			fmt.Println("")
			fmt.Println("Examples:")
			fmt.Println("  gokku ps:restart -a api-production")
			os.Exit(1)
		}
	}

	registry := containers.NewContainerRegistry()
	allContainers, err := registry.GetAllContainers(appName)
	if err != nil {
		fmt.Printf("Error getting containers: %v\n", err)
		os.Exit(1)
	}

	if len(allContainers) == 0 {
		fmt.Printf("No processes running for app '%s'\n", appName)
		return
	}

	fmt.Printf("Restarting processes for app '%s'...\n", appName)

	for _, container := range allContainers {
		fmt.Printf("-----> Restarting %s\n", container.Name)

		// Restart container
		cmd := exec.Command("docker", "restart", container.Name)
		if err := cmd.Run(); err != nil {
			fmt.Printf("       Error restarting container %s: %v\n", container.Name, err)
			continue
		}

		// Update status
		registry.UpdateContainerStatus(appName, container.ProcessType, container.Number, "running")
	}

	fmt.Printf("Restart complete for app '%s'\n", appName)
}

// handlePSStop handles the ps:stop command
func handlePSStop(args []string) {
	// Parse: gokku ps:stop web -a api (client)
	//        gokku ps:stop web APP_NAME (server)
	//        gokku ps:stop -a api (client, stop all)
	//        gokku ps:stop APP_NAME (server, stop all)
	isServerMode := internal.IsServerMode()

	var appName string
	var processType string

	if isServerMode {
		// Server mode: app name is last argument (or only argument if no process type)
		// Find app name (not a flag, not a process type argument)
		var foundAppName bool
		for i := len(args) - 1; i >= 0; i-- {
			arg := args[i]
			if !strings.HasPrefix(arg, "-") && !strings.Contains(arg, "=") {
				if !foundAppName {
					appName = arg
					foundAppName = true
				} else {
					processType = arg
					break
				}
			}
		}

		if appName == "" {
			fmt.Println("Error: App name is required")
			fmt.Println("Usage: gokku ps:stop [<process>] <app>")
			fmt.Println("")
			fmt.Println("Examples:")
			fmt.Println("  gokku ps:stop web api")
			fmt.Println("  gokku ps:stop api")
			os.Exit(1)
		}
	} else {
		// Client mode: require -a flag
		appName = internal.ExtractAppName(args)
		if appName == "" {
			fmt.Println("Error: -a <app> is required")
			fmt.Println("Usage: gokku ps:stop [<process>] -a <app>")
			fmt.Println("")
			fmt.Println("Examples:")
			fmt.Println("  gokku ps:stop web -a api-production")
			fmt.Println("  gokku ps:stop -a api-production")
			os.Exit(1)
		}

		// Check if a specific process type was provided
		for _, arg := range args {
			if !strings.HasPrefix(arg, "-a") && !strings.Contains(arg, "=") && arg != appName {
				processType = arg
				break
			}
		}
	}

	registry := containers.NewContainerRegistry()

	if processType != "" {
		// Stop specific process type
		containers, err := registry.GetContainers(appName, processType)
		if err != nil {
			fmt.Printf("Error getting containers: %v\n", err)
			os.Exit(1)
		}

		if len(containers) == 0 {
			fmt.Printf("No %s processes running for app '%s'\n", processType, appName)
			return
		}

		fmt.Printf("Stopping %s processes for app '%s'...\n", processType, appName)

		for _, container := range containers {
			fmt.Printf("-----> Stopping %s\n", container.Name)

			// Stop container
			cmd := exec.Command("docker", "stop", container.Name)
			if err := cmd.Run(); err != nil {
				fmt.Printf("       Error stopping container %s: %v\n", container.Name, err)
				continue
			}

			// Update status
			registry.UpdateContainerStatus(appName, container.ProcessType, container.Number, "stopped")
		}
	} else {
		// Stop all processes
		allContainers, err := registry.GetAllContainers(appName)
		if err != nil {
			fmt.Printf("Error getting containers: %v\n", err)
			os.Exit(1)
		}

		if len(allContainers) == 0 {
			fmt.Printf("No processes running for app '%s'\n", appName)
			return
		}

		fmt.Printf("Stopping all processes for app '%s'...\n", appName)

		for _, container := range allContainers {
			fmt.Printf("-----> Stopping %s\n", container.Name)

			// Stop container
			cmd := exec.Command("docker", "stop", container.Name)
			if err := cmd.Run(); err != nil {
				fmt.Printf("       Error stopping container %s: %v\n", container.Name, err)
				continue
			}

			// Update status
			registry.UpdateContainerStatus(appName, container.ProcessType, container.Number, "stopped")
		}
	}

	fmt.Printf("Stop complete for app '%s'\n", appName)
}

// notifyPluginsOfScaleChange notifies all plugins about scale changes
func notifyPluginsOfScaleChange(appName, processType string) {
	// Get all installed plugins
	pluginsDir := "/opt/gokku/plugins"

	plugins, err := os.ReadDir(pluginsDir)
	if err != nil {
		return // No plugins directory, skip
	}

	for _, plugin := range plugins {
		if !plugin.IsDir() {
			continue
		}

		pluginName := plugin.Name()

		// Check if plugin has a scale-change hook
		hookPath := fmt.Sprintf("%s/%s/hooks/scale-change", pluginsDir, pluginName)
		if _, err := os.Stat(hookPath); err == nil {
			// Execute plugin hook
			cmd := exec.Command("bash", hookPath, appName, processType)
			cmd.Run() // Don't fail if plugin hook fails
		}
	}
}

// imageExists checks if a Docker image exists
func imageExists(imageName string) bool {
	cmd := exec.Command("docker", "image", "inspect", imageName)
	return cmd.Run() == nil
}

// showPSHelp shows help for ps commands
func showPSHelp() {
	isServerMode := internal.IsServerMode()

	if isServerMode {
		fmt.Println("Process management commands (server mode):")
		fmt.Println("")
		fmt.Println("  gokku ps                                  List all running containers")
		fmt.Println("  gokku ps:scale <process=count>... <app>    Scale app processes")
		fmt.Println("  gokku ps:report <app>                      List running processes")
		fmt.Println("  gokku ps:list [<app>]                       List running processes (all if no app)")
		fmt.Println("  gokku ps:restart <app>                     Restart all processes")
		fmt.Println("  gokku ps:stop [<process>] <app>            Stop processes")
		fmt.Println("")
		fmt.Println("Examples:")
		fmt.Println("  gokku ps")
		fmt.Println("  gokku ps:scale web=4 worker=2 api")
		fmt.Println("  gokku ps:report api")
		fmt.Println("  gokku ps:list api")
		fmt.Println("  gokku ps:list")
		fmt.Println("  gokku ps:restart api")
		fmt.Println("  gokku ps:stop web api")
		fmt.Println("  gokku ps:stop api")
	} else {
		fmt.Println("Process management commands (client mode):")
		fmt.Println("")
		fmt.Println("  gokku ps:scale <process=count>... -a <app>    Scale app processes")
		fmt.Println("  gokku ps:report -a <app>                       List running processes")
		fmt.Println("  gokku ps:list -a <app>                         List running processes")
		fmt.Println("  gokku ps:restart -a <app>                     Restart all processes")
		fmt.Println("  gokku ps:stop [<process>] -a <app>             Stop processes")
		fmt.Println("")
		fmt.Println("Examples:")
		fmt.Println("  gokku ps:scale web=4 worker=2 -a api-production")
		fmt.Println("  gokku ps:report -a api-production")
		fmt.Println("  gokku ps:restart -a api-production")
		fmt.Println("  gokku ps:stop web -a api-production")
		fmt.Println("  gokku ps:stop -a api-production")
	}
}
