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

	// Check if --remote flag is present (new pattern)
	remoteInfo, remainingArgs, err := internal.GetRemoteInfoOrDefault(args)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// If only --remote flag provided (no subcommand), list all containers remotely
	if len(remainingArgs) == 0 && remoteInfo != nil {
		cmd := "gokku ps"
		if err := internal.ExecuteRemoteCommand(remoteInfo, cmd); err != nil {
			os.Exit(1)
		}
		return
	}

	// Determine subcommand from remaining args (skip flags like --remote)
	var subcommand string
	var subcommandIndex int
	for i, arg := range remainingArgs {
		if !strings.HasPrefix(arg, "-") || arg == "-a" || arg == "--app" {
			subcommand = arg
			subcommandIndex = i
			break
		}
	}

	// Check if first argument is a flag (-a or --app), treat as list command
	if len(remainingArgs) > 0 && (remainingArgs[0] == "-a" || remainingArgs[0] == "--app") {
		appName, _ := internal.ExtractAppFlag(remainingArgs)
		if appName == "" {
			fmt.Println("Error: App name is required")
			fmt.Println("Usage: gokku ps -a <app> [--remote]")
			fmt.Println("   or: gokku ps:list -a <app> [--remote]")
			os.Exit(1)
		}

		if remoteInfo != nil {
			// Client mode - execute remotely
			cmd := fmt.Sprintf("gokku ps:list --app %s", appName)
			if err := internal.ExecuteRemoteCommand(remoteInfo, cmd); err != nil {
				os.Exit(1)
			}
			return
		}

		// Server mode - execute locally
		handlePSList([]string{"-a", appName})
		return
	}

	// If no subcommand and no -a flag, just list all containers
	if subcommand == "" {
		if remoteInfo != nil {
			// Client mode - execute remotely (should have been caught above)
			cmd := "gokku ps"
			if err := internal.ExecuteRemoteCommand(remoteInfo, cmd); err != nil {
				os.Exit(1)
			}
			return
		}
		listAllContainers()
		return
	}

	// Process remaining args for subcommands (skip the subcommand itself)
	subcommandArgs := remainingArgs[subcommandIndex+1:]

	switch subcommand {
	case "report", "rs", "list":
		if remoteInfo != nil {
			cmdArgs := []string{subcommand}
			cmdArgs = append(cmdArgs, subcommandArgs...)
			cmd := fmt.Sprintf("gokku ps:%s %s", subcommand, strings.Join(cmdArgs[1:], " "))
			if err := internal.ExecuteRemoteCommand(remoteInfo, cmd); err != nil {
				os.Exit(1)
			}
			return
		}
		handlePSList(subcommandArgs)
	case "restart":
		if remoteInfo != nil {
			cmdArgs := []string{subcommand}
			cmdArgs = append(cmdArgs, subcommandArgs...)
			cmd := fmt.Sprintf("gokku ps:%s %s", subcommand, strings.Join(cmdArgs[1:], " "))
			if err := internal.ExecuteRemoteCommand(remoteInfo, cmd); err != nil {
				os.Exit(1)
			}
			return
		}
		handlePSRestart(subcommandArgs)
	case "stop":
		if remoteInfo != nil {
			cmdArgs := []string{subcommand}
			cmdArgs = append(cmdArgs, subcommandArgs...)
			cmd := fmt.Sprintf("gokku ps:%s %s", subcommand, strings.Join(cmdArgs[1:], " "))
			if err := internal.ExecuteRemoteCommand(remoteInfo, cmd); err != nil {
				os.Exit(1)
			}
			return
		}
		handlePSStop(subcommandArgs)
	default:
		// Check if it's a flag we should ignore
		if strings.HasPrefix(subcommand, "--") {
			// It's a flag, not a subcommand - list all
			if remoteInfo != nil {
				cmd := "gokku ps"
				if err := internal.ExecuteRemoteCommand(remoteInfo, cmd); err != nil {
					os.Exit(1)
				}
				return
			}
			listAllContainers()
			return
		}
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

	baseDir := "/opt/gokku"
	containerService := services.NewContainerService(baseDir)

	if len(allContainers) == 0 {
		// Fallback: try to find containers directly by name for backwards compatibility
		// This handles containers created before labels or not registered in the registry
		filter := services.ContainerFilter{
			AppName: appName,
			All:     true,
		}

		containers, err := containerService.ListContainers(filter)
		if err == nil && len(containers) > 0 {
			fmt.Printf("Restarting processes for app '%s' (found %d container(s))...\n", appName, len(containers))
			for _, container := range containers {
				fmt.Printf("-----> Restarting %s\n", container.Names)
				containerService.RestartContainer(container.Names) // Ignore errors for backwards compatibility
			}
			fmt.Printf("Restart complete for app '%s'\n", appName)
			return
		}

		fmt.Printf("No processes running for app '%s'\n", appName)
		return
	}

	fmt.Printf("Restarting processes for app '%s'...\n", appName)

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

// showPSHelp shows help for ps commands
func showPSHelp() {
	isServerMode := internal.IsServerMode()

	if isServerMode {
		fmt.Println("Process management commands (server mode):")
		fmt.Println("")
		fmt.Println("  gokku ps                                  List all running containers")
		fmt.Println("  gokku ps:report <app>                      List running processes")
		fmt.Println("  gokku ps:list [<app>]                       List running processes (all if no app)")
		fmt.Println("  gokku ps:restart <app>                     Restart all processes")
		fmt.Println("  gokku ps:stop [<process>] <app>            Stop processes")
		fmt.Println("")
		fmt.Println("Examples:")
		fmt.Println("  gokku ps")
		fmt.Println("  gokku ps:report api")
		fmt.Println("  gokku ps:list api")
		fmt.Println("  gokku ps:list")
		fmt.Println("  gokku ps:restart api")
		fmt.Println("  gokku ps:stop web api")
		fmt.Println("  gokku ps:stop api")
	} else {
		fmt.Println("Process management commands (client mode):")
		fmt.Println("")
		fmt.Println("  gokku ps:report -a <app>                       List running processes")
		fmt.Println("  gokku ps:list -a <app>                         List running processes")
		fmt.Println("  gokku ps:restart -a <app>                     Restart all processes")
		fmt.Println("  gokku ps:stop [<process>] -a <app>             Stop processes")
		fmt.Println("")
		fmt.Println("Examples:")
		fmt.Println("  gokku ps:report -a api-production")
		fmt.Println("  gokku ps:restart -a api-production")
		fmt.Println("  gokku ps:stop web -a api-production")
		fmt.Println("  gokku ps:stop -a api-production")
	}
}
