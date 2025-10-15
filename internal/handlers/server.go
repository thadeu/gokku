package handlers

import (
	"fmt"
	"os"

	"infra/internal"
)

// handleServer manages server configuration commands
func handleServer(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: gokku server <add|list|remove|set-default>")
		os.Exit(1)
	}

	config, err := internal.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	subcommand := args[0]

	switch subcommand {
	case "add":
		if len(args) < 3 {
			fmt.Println("Usage: gokku server add <name> <host>")
			os.Exit(1)
		}
		name := args[1]
		host := args[2]

		// Check if server already exists
		for _, s := range config.Servers {
			if s.Name == name {
				fmt.Printf("Server '%s' already exists\n", name)
				os.Exit(1)
			}
		}

		server := internal.Server{
			Name:    name,
			Host:    host,
			BaseDir: "/opt/gokku",
			Default: len(config.Servers) == 0, // First server is default
		}
		config.Servers = append(config.Servers, server)

		if err := internal.SaveConfig(config); err != nil {
			fmt.Printf("Error saving config: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✓ Server '%s' added\n", name)
		if server.Default {
			fmt.Println("  Set as default server")
		}

	case "list":
		if len(config.Servers) == 0 {
			fmt.Println("No servers configured")
			fmt.Println("\nAdd a server:")
			fmt.Println("  gokku server add production ubuntu@ec2.compute.amazonaws.com")
			return
		}

		fmt.Println("Configured servers:")
		for _, server := range config.Servers {
			defaultMarker := ""
			if server.Default {
				defaultMarker = " (default)"
			}
			fmt.Printf("  • %s: %s%s\n", server.Name, server.Host, defaultMarker)
		}

	case "remove":
		if len(args) < 2 {
			fmt.Println("Usage: gokku server remove <name>")
			os.Exit(1)
		}
		name := args[1]

		found := false
		newServers := []internal.Server{}
		for _, s := range config.Servers {
			if s.Name != name {
				newServers = append(newServers, s)
			} else {
				found = true
			}
		}

		if !found {
			fmt.Printf("Server '%s' not found\n", name)
			os.Exit(1)
		}

		config.Servers = newServers
		if err := internal.SaveConfig(config); err != nil {
			fmt.Printf("Error saving config: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✓ Server '%s' removed\n", name)

	case "set-default":
		if len(args) < 2 {
			fmt.Println("Usage: gokku server set-default <name>")
			os.Exit(1)
		}
		name := args[1]

		found := false
		for i := range config.Servers {
			if config.Servers[i].Name == name {
				config.Servers[i].Default = true
				found = true
			} else {
				config.Servers[i].Default = false
			}
		}

		if !found {
			fmt.Printf("Server '%s' not found\n", name)
			os.Exit(1)
		}

		if err := internal.SaveConfig(config); err != nil {
			fmt.Printf("Error saving config: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✓ '%s' set as default server\n", name)

	default:
		fmt.Printf("Unknown subcommand: %s\n", subcommand)
		os.Exit(1)
	}
}
