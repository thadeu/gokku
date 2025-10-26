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
		showPSHelp()
		return
	}

	subcommand := args[0]

	switch subcommand {
	case "scale":
		handlePSScale(args[1:])
	case "report", "rs":
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

// handlePSScale handles the ps:scale command
func handlePSScale(args []string) {
	// Parse: gokku ps:scale web=4 worker=2 -a api
	appName := internal.ExtractAppName(args)
	if appName == "" {
		fmt.Println("Error: -a <app> is required")
		os.Exit(1)
	}

	// Parse scale arguments
	scales := make(map[string]int)
	for _, arg := range args {
		if strings.HasPrefix(arg, "-a") {
			continue // Skip app flag
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
			"-p", fmt.Sprintf("%d:8080", hostPort),
			"--env-file", fmt.Sprintf("/opt/gokku/apps/%s/.env", appName),
			"--ulimit", "nofile=65536:65536",
			"--ulimit", "nproc=4096:4096",
			appImage)

		if err := cmd.Run(); err != nil {
			fmt.Printf("       Error creating container %s: %v\n", containerName, err)
			continue
		}

		// Save container info
		info := containers.CreateContainerInfo(appName, processType, containerNum, hostPort, 8080)
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
	appName := internal.ExtractAppName(args)
	if appName == "" {
		fmt.Println("Error: -a <app> is required")
		os.Exit(1)
	}

	// Use docker ps to get running containers
	cmd := exec.Command("docker", "ps", "--format", "table {{.Names}}\t{{.Status}}\t{{.Ports}}")
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("Error running docker ps: %v\n", err)
		os.Exit(1)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) <= 1 {
		fmt.Printf("No processes running for app '%s'\n", appName)
		return
	}

	// Filter containers that belong to this app
	var appContainers []string
	for _, line := range lines[1:] { // Skip header
		if line == "" {
			continue
		}

		// Check if container name starts with app name
		parts := strings.Fields(line)
		if len(parts) > 0 {
			containerName := parts[0]
			if strings.HasPrefix(containerName, appName+"-") || containerName == appName {
				appContainers = append(appContainers, line)
			}
		}
	}

	if len(appContainers) == 0 {
		fmt.Printf("No processes running for app '%s'\n", appName)
		return
	}

	fmt.Printf("Processes for app '%s':\n", appName)
	fmt.Printf("%-20s %-10s %-15s\n", "NAME", "STATUS", "PORT")
	fmt.Printf("%-20s %-10s %-15s\n", "----", "------", "----")

	for _, container := range appContainers {
		parts := strings.Fields(container)
		if len(parts) >= 3 {
			name := parts[0]
			status := parts[1]
			ports := parts[2]
			fmt.Printf("%-20s %-10s %-15s\n", name, status, ports)
		}
	}
}

// handlePSRestart handles the ps:restart command
func handlePSRestart(args []string) {
	appName := internal.ExtractAppName(args)
	if appName == "" {
		fmt.Println("Error: -a <app> is required")
		os.Exit(1)
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
	appName := internal.ExtractAppName(args)
	if appName == "" {
		fmt.Println("Error: -a <app> is required")
		os.Exit(1)
	}

	// Check if a specific process type was provided
	var processType string
	for _, arg := range args {
		if !strings.HasPrefix(arg, "-a") && !strings.Contains(arg, "=") {
			processType = arg
			break
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
	fmt.Println("Process management commands:")
	fmt.Println("")
	fmt.Println("  gokku ps:scale <process=count>... -a <app>    Scale app processes")
	fmt.Println("  gokku ps:report -a <app>                       List running processes")
	fmt.Println("  gokku ps:restart -a <app>                    Restart all processes")
	fmt.Println("  gokku ps:stop [<process>] -a <app>           Stop processes")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  gokku ps:scale web=4 worker=2 -a api")
	fmt.Println("  gokku ps:report -a api")
	fmt.Println("  gokku ps:restart -a api")
	fmt.Println("  gokku ps:stop web -a api")
	fmt.Println("  gokku ps:stop -a api")
}
