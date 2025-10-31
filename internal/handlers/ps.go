package handlers

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"gokku/internal"
	"gokku/internal/containers"
	"gokku/internal/services"
	"gokku/tui"
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

	baseDir := "/opt/gokku"
	if !isServerMode {
		baseDir = "/opt/gokku" // Will be executed on server via SSH
	}

	containerService := services.NewContainerService(baseDir)

	// Get current counts for display
	registry := containers.NewContainerRegistry()
	for processType, count := range scales {
		currentContainers, err := registry.GetContainers(appName, processType)
		if err != nil {
			fmt.Printf("Error getting current containers: %v\n", err)
			continue
		}

		currentCount := len(currentContainers)
		fmt.Printf("-----> Scaling %s from %d to %d instances\n", processType, currentCount, count)

		if count == currentCount {
			fmt.Printf("       %s already at %d instances\n", processType, count)
			continue
		}
	}

	// Perform scaling
	if err := containerService.ScaleProcesses(appName, scales); err != nil {
		fmt.Printf("Error scaling app: %v\n", err)
		os.Exit(1)
	}

	// Notify plugins about scale change
	for processType := range scales {
		notifyPluginsOfScaleChange(appName, processType)
	}

	fmt.Printf("Scaling complete for app '%s'\n", appName)
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

	// Server mode - use ContainerService
	baseDir := "/opt/gokku"
	containerService := services.NewContainerService(baseDir)

	filter := services.ContainerFilter{
		AppName: appName,
		All:     false,
	}

	containers, err := containerService.ListContainers(filter)
	if err != nil {
		fmt.Printf("Error listing containers: %v\n", err)
		os.Exit(1)
	}

	if len(containers) == 0 {
		fmt.Printf("No processes running for app '%s'\n", appName)
		return
	}

	fmt.Printf("Processes for app '%s':\n", appName)

	table := tui.NewTable(tui.ASCII)
	table.AppendHeaders([]string{"NAME", "STATUS", "PORT"})
	table.AppendSeparator()

	for _, container := range containers {
		table.AppendRow([]string{
			container.Names,
			container.Status,
			container.Ports,
		})
	}

	fmt.Print(table.Render())
}

// listAllContainers lists all running containers with gokku format
func listAllContainers() {
	baseDir := "/opt/gokku"
	containerService := services.NewContainerService(baseDir)

	filter := services.ContainerFilter{
		All: false,
	}

	containers, err := containerService.ListContainers(filter)
	if err != nil {
		fmt.Printf("Error listing containers: %v\n", err)
		os.Exit(1)
	}

	if len(containers) == 0 {
		fmt.Println("No processes running")
		return
	}

	table := tui.NewTable(tui.ASCII)
	table.AppendHeaders([]string{"NAME", "STATUS", "PORT"})

	for _, container := range containers {
		table.AppendRow([]string{
			container.Names,
			container.Status,
			container.Ports,
		})
		table.AppendSeparator()
	}

	fmt.Print(table.Render())
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

	baseDir := "/opt/gokku"
	containerService := services.NewContainerService(baseDir)

	for _, container := range allContainers {
		fmt.Printf("-----> Restarting %s\n", container.Name)

		// Restart container using service
		if err := containerService.RestartContainer(container.Name); err != nil {
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

		baseDir := "/opt/gokku"
		containerService := services.NewContainerService(baseDir)

		for _, container := range containers {
			fmt.Printf("-----> Stopping %s\n", container.Name)

			// Stop container using service
			if err := containerService.StopContainer(container.Name); err != nil {
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
			// Fallback: try to find containers directly by name for backwards compatibility
			// This handles containers created before labels or not registered in the registry
			baseDir := "/opt/gokku"
			containerService := services.NewContainerService(baseDir)

			filter := services.ContainerFilter{
				AppName: appName,
				All:     true,
			}

			containers, err := containerService.ListContainers(filter)
			if err == nil && len(containers) > 0 {
				fmt.Printf("Stopping all processes for app '%s' (found %d container(s))...\n", appName, len(containers))
				for _, container := range containers {
					fmt.Printf("-----> Stopping %s\n", container.Names)
					containerService.StopContainer(container.Names) // Ignore errors for backwards compatibility
				}
				fmt.Printf("Stop complete for app '%s'\n", appName)
				return
			}

			fmt.Printf("No processes running for app '%s'\n", appName)
			return
		}

		fmt.Printf("Stopping all processes for app '%s'...\n", appName)

		baseDir := "/opt/gokku"
		containerService := services.NewContainerService(baseDir)

		for _, container := range allContainers {
			fmt.Printf("-----> Stopping %s\n", container.Name)

			// Stop container using service
			if err := containerService.StopContainer(container.Name); err != nil {
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
