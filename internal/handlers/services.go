package handlers

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gokku/internal"
	"gokku/internal/services"
	"gokku/tui"
)

// handleServices manages service-related commands
func handleServices(args []string) {
	if len(args) == 0 {
		showServicesHelp()
		return
	}

	// Extract --remote flag first (if present)
	remoteInfo, remainingArgs, err := internal.GetRemoteInfoOrDefault(args)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if len(remainingArgs) == 0 {
		// Only --remote provided, show help or list services
		if remoteInfo != nil {
			cmd := "gokku services list"
			if err := internal.ExecuteRemoteCommand(remoteInfo, cmd); err != nil {
				os.Exit(1)
			}
			return
		}
		showServicesHelp()
		return
	}

	subcommand := remainingArgs[0]

	// Check if subcommand is a flag (like --remote that wasn't caught)
	if strings.HasPrefix(subcommand, "--") && subcommand != "--help" && subcommand != "--remote" {
		fmt.Printf("Unknown service command: %s\n", subcommand)
		showServicesHelp()
		os.Exit(1)
	}

	switch subcommand {
	case "list", "ls":
		handleServicesList(remainingArgs[1:], remoteInfo)
	case "create":
		handleServicesCreate(remainingArgs[1:], remoteInfo)
	case "link":
		handleServicesLink(remainingArgs[1:], remoteInfo)
	case "unlink":
		handleServicesUnlink(remainingArgs[1:], remoteInfo)
	case "destroy":
		handleServicesDestroy(remainingArgs[1:], remoteInfo)
	case "info":
		handleServicesInfo(remainingArgs[1:], remoteInfo)
	case "logs":
		handleServicesLogs(remainingArgs[1:], remoteInfo)
	default:
		// Try to execute as service command
		if remoteInfo != nil {
			// Execute service command remotely
			cmd := fmt.Sprintf("gokku services %s", strings.Join(remainingArgs, " "))
			if err := internal.ExecuteRemoteCommand(remoteInfo, cmd); err != nil {
				os.Exit(1)
			}
			return
		}
		handleServiceCommand(remainingArgs)
	}
}

// handleServicesList lists all services
func handleServicesList(args []string, remoteInfo *internal.RemoteInfo) {
	if remoteInfo != nil {
		// Client mode: execute remotely
		cmd := "gokku services list"
		if err := internal.ExecuteRemoteCommand(remoteInfo, cmd); err != nil {
			fmt.Printf("Error listing services: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Server mode: execute locally
	sm := services.NewServiceManager()

	serviceList, err := sm.ListServices()
	if err != nil {
		fmt.Printf("Error listing services: %v\n", err)
		os.Exit(1)
	}

	if len(serviceList) == 0 {
		fmt.Println("No services created")
		return
	}

	fmt.Print("===== Services\n")

	table := tui.NewTable(tui.ASCII)
	table.AppendHeaders([]string{"NAME", "PLUGIN", "STATUS"})

	for _, service := range serviceList {
		status := "stopped"

		if service.Running {
			status = "running"
		}

		// fmt.Printf("  %s (%s) - %s\n", service.Name, service.Plugin, status)
		table.AppendRow([]string{service.Name, service.Plugin, status}, len(service.Name) > 110)

		if len(service.LinkedApps) > 0 {
			// fmt.Printf("    Linked to: %s\n", strings.Join(service.LinkedApps, ", "))
			table.AppendRow([]string{service.Name, service.Plugin, status, strings.Join(service.LinkedApps, ", ")}, len(service.Name) > 110)
		}
	}

	fmt.Print(table.Render())
	_ = args // unused for now
}

// handleServicesCreate creates a new service
func handleServicesCreate(args []string, remoteInfo *internal.RemoteInfo) {
	// Use args directly if remoteInfo is already set (from parent)
	cleanArgs := args

	if len(cleanArgs) < 2 {
		fmt.Println("Usage: gokku services:create <plugin>[:<version>] --name <service-name> [--remote]")
		fmt.Println("")
		fmt.Println("Examples:")
		fmt.Println("  gokku services:create postgres --name postgres-api")
		fmt.Println("  gokku services:create postgres:14 --name postgres-api")
		fmt.Println("  gokku services:create redis:7 --name redis-cache")
		fmt.Println("  gokku services:create postgres --name postgres-api --remote")
		os.Exit(1)
	}

	// If remote mode, execute remotely
	if remoteInfo != nil {
		cmd := fmt.Sprintf("gokku services:create %s", strings.Join(cleanArgs, " "))
		if err := internal.ExecuteRemoteCommand(remoteInfo, cmd); err != nil {
			fmt.Printf("Error creating service: %v\n", err)
			os.Exit(1)
		}
		return
	}

	pluginWithVersion := cleanArgs[0]
	serviceName := internal.ExtractFlagValue(cleanArgs, "--name")

	if serviceName == "" {
		fmt.Println("Error: --name is required")
		os.Exit(1)
	}

	// Split plugin and version by ':'
	parts := strings.Split(pluginWithVersion, ":")
	pluginName := parts[0]
	version := ""
	if len(parts) > 1 {
		version = parts[1]
	}

	sm := services.NewServiceManager()

	if version != "" {
		fmt.Printf("Creating service '%s' from plugin '%s:%s'...\n", serviceName, pluginName, version)
	} else {
		fmt.Printf("Creating service '%s' from plugin '%s'...\n", serviceName, pluginName)
	}

	if err := sm.CreateService(pluginName, serviceName, version); err != nil {
		fmt.Printf("Error creating service: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Service '%s' created successfully\n", serviceName)
}

// handleServicesLink links a service to an app
func handleServicesLink(args []string, remoteInfo *internal.RemoteInfo) {
	// Use args directly if remoteInfo is already set (from parent)
	cleanArgs := args

	if len(cleanArgs) < 1 {
		fmt.Println("Usage: gokku services:link <service> -a <app> [--remote]")
		fmt.Println("")
		fmt.Println("Examples:")
		fmt.Println("  gokku services:link postgres-api -a api-production")
		fmt.Println("  gokku services:link redis-cache -a api-production")
		fmt.Println("  gokku services:link postgres-api -a api-production --remote")
		os.Exit(1)
	}

	// If remote mode, execute remotely
	if remoteInfo != nil {
		cmd := fmt.Sprintf("gokku services:link %s", strings.Join(cleanArgs, " "))
		if err := internal.ExecuteRemoteCommand(remoteInfo, cmd); err != nil {
			fmt.Printf("Error linking service: %v\n", err)
			os.Exit(1)
		}
		return
	}

	serviceName := cleanArgs[0]
	appName := internal.ExtractAppName(cleanArgs[1:])

	sm := services.NewServiceManager()

	fmt.Printf("Linking service '%s' to app '%s'...\n", serviceName, appName)

	if err := sm.LinkService(serviceName, appName, "default"); err != nil {
		fmt.Printf("Error linking service: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Service '%s' linked to '%s' successfully\n", serviceName, appName)
	fmt.Printf("Environment variables updated. Restart your app to use the service.\n")
}

// handleServicesUnlink unlinks a service from an app
func handleServicesUnlink(args []string, remoteInfo *internal.RemoteInfo) {
	// Use args directly if remoteInfo is already set (from parent)
	cleanArgs := args

	if len(cleanArgs) < 1 {
		fmt.Println("Usage: gokku services:unlink <service> -a <app> [--remote]")
		fmt.Println("")
		fmt.Println("Examples:")
		fmt.Println("  gokku services:unlink postgres-api -a api-production")
		fmt.Println("  gokku services:unlink postgres-api -a api-production --remote")
		os.Exit(1)
	}

	// If remote mode, execute remotely
	if remoteInfo != nil {
		cmd := fmt.Sprintf("gokku services:unlink %s", strings.Join(cleanArgs, " "))
		if err := internal.ExecuteRemoteCommand(remoteInfo, cmd); err != nil {
			fmt.Printf("Error unlinking service: %v\n", err)
			os.Exit(1)
		}
		return
	}

	serviceName := cleanArgs[0]
	appName := internal.ExtractAppName(cleanArgs[1:])

	sm := services.NewServiceManager()

	fmt.Printf("Unlinking service '%s' from app '%s'...\n", serviceName, appName)

	if err := sm.UnlinkService(serviceName, appName, "default"); err != nil {
		fmt.Printf("Error unlinking service: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Service '%s' unlinked from '%s' successfully\n", serviceName, appName)
}

// handleServicesDestroy destroys a service
func handleServicesDestroy(args []string, remoteInfo *internal.RemoteInfo) {
	// Use args directly if remoteInfo is already set (from parent)
	cleanArgs := args

	if len(cleanArgs) < 1 {
		fmt.Println("Usage: gokku services:destroy <service> [--remote]")
		fmt.Println("")
		fmt.Println("Examples:")
		fmt.Println("  gokku services:destroy postgres-api")
		fmt.Println("  gokku services:destroy redis-cache")
		fmt.Println("  gokku services:destroy postgres-api --remote")
		os.Exit(1)
	}

	serviceName := cleanArgs[0]

	// Handle help flag
	if serviceName == "--help" || serviceName == "-h" {
		fmt.Println("Usage: gokku services:destroy <service> [--remote]")
		fmt.Println("")
		fmt.Println("Examples:")
		fmt.Println("  gokku services:destroy postgres-api")
		fmt.Println("  gokku services:destroy redis-cache")
		fmt.Println("  gokku services:destroy postgres-api --remote")
		return
	}

	// If remote mode, execute remotely
	if remoteInfo != nil {
		cmd := fmt.Sprintf("gokku services:destroy %s", serviceName)
		if err := internal.ExecuteRemoteCommand(remoteInfo, cmd); err != nil {
			fmt.Printf("Error destroying service: %v\n", err)
			os.Exit(1)
		}
		return
	}

	sm := services.NewServiceManager()

	fmt.Printf("Destroying service '%s'...\n", serviceName)

	if err := sm.DestroyService(serviceName); err != nil {
		fmt.Printf("Error destroying service: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Service '%s' destroyed successfully\n", serviceName)
}

// handleServicesInfo shows service information
func handleServicesInfo(args []string, remoteInfo *internal.RemoteInfo) {
	// Use args directly if remoteInfo is already set (from parent)
	cleanArgs := args

	if len(cleanArgs) < 1 {
		fmt.Println("Usage: gokku services:info <service> [--remote]")
		fmt.Println("")
		fmt.Println("Examples:")
		fmt.Println("  gokku services:info postgres-api")
		fmt.Println("  gokku services:info postgres-api --remote")
		os.Exit(1)
	}

	// If remote mode, execute remotely
	if remoteInfo != nil {
		cmd := fmt.Sprintf("gokku services:info %s", strings.Join(cleanArgs, " "))
		if err := internal.ExecuteRemoteCommand(remoteInfo, cmd); err != nil {
			fmt.Printf("Error getting service info: %v\n", err)
			os.Exit(1)
		}
		return
	}

	serviceName := cleanArgs[0]

	sm := services.NewServiceManager()

	// Check if service exists
	service, err := sm.GetService(serviceName)
	if err != nil {
		fmt.Printf("Service '%s' not found\n", serviceName)
		os.Exit(1)
	}

	// Execute plugin's info command
	infoCommand := filepath.Join("/opt/gokku/plugins", service.Plugin, "commands", "info")

	if _, err := os.Stat(infoCommand); err != nil {
		fmt.Printf("Info command not available for plugin '%s'\n", service.Plugin)
		os.Exit(1)
	}

	// Execute the plugin's info command
	cmd := exec.Command("bash", infoCommand, serviceName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("Error executing info command: %v\n", err)
		os.Exit(1)
	}
}

// handleServicesLogs shows service logs
func handleServicesLogs(args []string, remoteInfo *internal.RemoteInfo) {
	// Use args directly if remoteInfo is already set (from parent)
	cleanArgs := args

	if len(cleanArgs) < 1 {
		fmt.Println("Usage: gokku services:logs <service> [-f|--follow] [--remote]")
		fmt.Println("")
		fmt.Println("Examples:")
		fmt.Println("  gokku services:logs postgres-api")
		fmt.Println("  gokku services:logs postgres-api -f")
		fmt.Println("  gokku services:logs postgres-api --remote")
		os.Exit(1)
	}

	// If remote mode, execute remotely
	if remoteInfo != nil {
		cmd := fmt.Sprintf("gokku services:logs %s", strings.Join(cleanArgs, " "))
		if err := internal.ExecuteRemoteCommand(remoteInfo, cmd); err != nil {
			fmt.Printf("Error getting service logs: %v\n", err)
			os.Exit(1)
		}
		return
	}

	serviceName := cleanArgs[0]

	sm := services.NewServiceManager()

	// Check if service exists
	service, err := sm.GetService(serviceName)
	if err != nil {
		fmt.Printf("Service '%s' not found\n", serviceName)
		os.Exit(1)
	}

	// Execute plugin's logs command
	logsCommand := filepath.Join("/opt/gokku/plugins", service.Plugin, "commands", "logs")

	if _, err := os.Stat(logsCommand); err != nil {
		fmt.Printf("Logs command not available for plugin '%s'\n", service.Plugin)
		os.Exit(1)
	}

	// Execute the plugin's logs command
	cmd := exec.Command("bash", logsCommand, serviceName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("Error executing logs command: %v\n", err)
		os.Exit(1)
	}
}

// handleServiceCommand executes a service command
func handleServiceCommand(args []string) {
	// Parse: gokku postgres:export postgres-api
	parts := strings.Split(args[0], ":")
	if len(parts) != 2 {
		fmt.Printf("Unknown command: %s\n", args[0])
		showServicesHelp()
		os.Exit(1)
	}

	pluginName := parts[0]
	command := parts[1]

	// Get service name from args
	if len(args) < 2 {
		fmt.Printf("Usage: gokku %s:%s <service>\n", pluginName, command)
		os.Exit(1)
	}

	serviceName := args[1]

	// Execute plugin command
	commandPath := filepath.Join("/opt/gokku/plugins", pluginName, "commands", command)

	if _, err := os.Stat(commandPath); err != nil {
		fmt.Printf("Command '%s' not found for plugin '%s'\n", command, pluginName)
		os.Exit(1)
	}

	// Execute the plugin command
	cmd := exec.Command("bash", commandPath, serviceName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("Error executing command: %v\n", err)
		os.Exit(1)
	}
}

// showServicesHelp shows services help
func showServicesHelp() {
	fmt.Println("Service management commands:")
	fmt.Println("")
	fmt.Println("  gokku services:list                    List all services")
	fmt.Println("  gokku services:create <plugin> --name <service>  Create service from plugin")
	fmt.Println("  gokku services:link <service> -a <app> Link service to app")
	fmt.Println("  gokku services:unlink <service> -a <app>  Unlink service from app")
	fmt.Println("  gokku services:destroy <service>       Destroy service")
	fmt.Println("  gokku services:info <service>          Show service information")
	fmt.Println("  gokku services:logs <service>          Show service logs")
	fmt.Println("")
	fmt.Println("Service commands:")
	fmt.Println("  gokku <plugin>:<command> <service>     Execute plugin command on service")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  gokku services:create postgres --name postgres-api")
	fmt.Println("  gokku services:create postgres:14 --name postgres-api")
	fmt.Println("  gokku services:create redis:7 --name redis-cache")
	fmt.Println("  gokku services:link postgres-api -a api-production")
	fmt.Println("  gokku postgres:backup postgres-api > backup.sql")
}

