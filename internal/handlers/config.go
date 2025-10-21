package handlers

import (
	"fmt"
	"os"
	"strings"

	"infra/internal"
)

// handleConfigWithContext manages environment variable configuration using context
func handleConfigWithContext(ctx *internal.ExecutionContext, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: gokku config <set|get|list|unset> [KEY[=VALUE]] [options]")
		fmt.Println("")
		fmt.Println("Options:")
		fmt.Println("  -a, --app <app>           App name")
		fmt.Println("")
		fmt.Println("Examples:")
		fmt.Println("  # Client mode (from local machine)")
		fmt.Println("  gokku config set PORT=8080 -a api-production")
		fmt.Println("  gokku config list -a api-production")
		fmt.Println("")
		fmt.Println("  # Server mode (on server)")
		fmt.Println("  gokku config set PORT=8080 -a api")
		fmt.Println("  gokku config list -a api")
		os.Exit(1)
	}

	// Validate that app is required
	if err := ctx.ValidateAppRequired(); err != nil {
		ctx.PrintUsageError("config", err.Error())
	}

	// Extract remaining args (without -a flag)
	_, remainingArgs := internal.ExtractAppFlag(args)

	if len(remainingArgs) < 1 {
		fmt.Println("Usage: gokku config <set|get|list|unset> [args...] -a <app>")
		os.Exit(1)
	}

	subcommand := remainingArgs[0]

	// Check if we're running locally on server or remotely
	if ctx.ServerExecution {
		// Server mode - use internal functions directly
		executeAsServerMode(ctx, subcommand, remainingArgs[1:])
	} else {
		// Client mode - execute command on remote server
		executeAsClientMode(ctx, subcommand, remainingArgs[1:])
	}
}

// handleexecuteAsServerMode handles config commands when running on server
func executeAsServerMode(ctx *internal.ExecutionContext, subcommand string, args []string) {
	appName := ctx.GetAppName()
	envFile := fmt.Sprintf("%s/apps/%s/shared/.env", ctx.BaseDir, appName)

	switch subcommand {
	case "set":
		if len(args) < 1 {
			fmt.Println("Usage: gokku config set KEY=VALUE [KEY2=VALUE2...] -a <app>")
			os.Exit(1)
		}
		internal.EnvSet(envFile, args)
	case "get":
		if len(args) < 1 {
			fmt.Println("Usage: gokku config get KEY -a <app>")
			os.Exit(1)
		}
		internal.EnvGet(envFile, args[0])
	case "list":
		internal.EnvList(envFile)
	case "unset":
		if len(args) < 1 {
			fmt.Println("Usage: gokku config unset KEY [KEY2...] -a <app>")
			os.Exit(1)
		}
		internal.EnvUnset(envFile, args)
	default:
		fmt.Printf("Unknown subcommand: %s\n", subcommand)
		os.Exit(1)
	}

	// Auto-restart container after set/unset to apply changes
	if subcommand == "set" || subcommand == "unset" {
		fmt.Printf("\n-----> Restarting container to apply changes...\n")
		restartCmd := fmt.Sprintf("gokku restart -a %s", appName)
		if err := ctx.ExecuteCommand(restartCmd); err != nil {
			fmt.Printf("Warning: Failed to restart container. Run 'gokku restart -a %s' manually.\n", appName)
		} else {
			fmt.Printf("✓ Container restarted with new configuration\n")
		}
	}
}

// executeAsClientMode handles config commands when running from client
func executeAsClientMode(ctx *internal.ExecutionContext, subcommand string, args []string) {
	// Build command to execute
	var cmd string
	switch subcommand {
	case "set":
		if len(args) < 1 {
			fmt.Println("Usage: gokku config set KEY=VALUE [KEY2=VALUE2...] -a <app>")
			os.Exit(1)
		}
		pairs := strings.Join(args, " ")
		cmd = fmt.Sprintf("gokku config set %s --app %s", pairs, ctx.GetAppName())
	case "get":
		if len(args) < 1 {
			fmt.Println("Usage: gokku config get KEY -a <app>")
			os.Exit(1)
		}
		key := args[0]
		cmd = fmt.Sprintf("gokku config get %s --app %s", key, ctx.GetAppName())
	case "list":
		cmd = fmt.Sprintf("gokku config list --app %s", ctx.GetAppName())
	case "unset":
		if len(args) < 1 {
			fmt.Println("Usage: gokku config unset KEY [KEY2...] -a <app>")
			os.Exit(1)
		}
		keys := strings.Join(args, " ")
		cmd = fmt.Sprintf("gokku config unset %s --app %s", keys, ctx.GetAppName())
	default:
		fmt.Printf("Unknown subcommand: %s\n", subcommand)
		os.Exit(1)
	}

	// Print connection info for remote execution
	ctx.PrintConnectionInfo()

	// Execute command
	if err := ctx.ExecuteCommand(cmd); err != nil {
		os.Exit(1)
	}

	// Auto-restart container after set/unset to apply changes
	if subcommand == "set" || subcommand == "unset" {
		fmt.Printf("\n-----> Restarting container to apply changes...\n")
		restartCmd := fmt.Sprintf("gokku restart -a %s", ctx.GetAppName())

		if err := ctx.ExecuteCommand(restartCmd); err != nil {
			fmt.Printf("Warning: Failed to restart container. Run 'gokku restart -a %s' manually.\n", ctx.AppName)
		} else {
			fmt.Printf("✓ Container restarted with new configuration\n")
		}
	}
}
