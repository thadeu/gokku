package commands

import (
	"fmt"
	"os"
	"strings"

	"gokku/internal"
	"gokku/internal/services"
)

func useConfigWithContext(ctx *internal.ExecutionContext, args []string) {
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

	if ctx == nil {
		fmt.Println("Error: Execution context is required")
		fmt.Println("Usage: gokku config <set|get|list|unset> [args...] -a <app>")
		os.Exit(1)
	}

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

	if ctx.ServerExecution {
		executeAsServerMode(ctx, subcommand, remainingArgs[1:])
	} else {
		executeAsClientMode(ctx, subcommand, remainingArgs[1:])
	}
}

func executeAsServerMode(ctx *internal.ExecutionContext, subcommand string, args []string) {
	appName := ctx.GetAppName()
	configService := services.NewConfigService(ctx.BaseDir)

	if subcommand == "" {
		subcommand = "list"
	}

	switch subcommand {
	case "set":
		if len(args) < 1 {
			fmt.Println("Usage: gokku config set KEY=VALUE [KEY2=VALUE2...] -a <app>")
			os.Exit(1)
		}

		if err := configService.SetEnvVar(appName, args); err != nil {
			fmt.Printf("Error setting config: %v\n", err)
			os.Exit(1)
		}

		for _, arg := range args {
			fmt.Println(arg)
		}
	case "get":
		if len(args) < 1 {
			fmt.Println("Error: KEY is required for config get")
			fmt.Println("Usage: gokku config get KEY -a <app>")
			fmt.Println("")
			fmt.Println("Example:")
			fmt.Printf("  gokku config get PORT -a %s\n", appName)
			fmt.Println("")
			fmt.Println("To list all config variables, use:")
			fmt.Printf("  gokku config list -a %s\n", appName)
			os.Exit(1)
		}

		value, err := configService.GetEnvVar(appName, args[0])

		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("%s=%s\n", args[0], value)
	case "list":
		envVars := configService.ListEnvVars(appName)

		if len(envVars) == 0 {
			fmt.Println("No environment variables set")
			return
		}

		// Sort keys for consistent output
		keys := make([]string, 0, len(envVars))

		for k := range envVars {
			keys = append(keys, k)
		}

		// Sort alphabetically
		for i := 0; i < len(keys); i++ {
			for j := i + 1; j < len(keys); j++ {
				if keys[i] > keys[j] {
					keys[i], keys[j] = keys[j], keys[i]
				}
			}
		}

		for _, key := range keys {
			fmt.Printf("%s=%s\n", key, envVars[key])
		}
	case "unset":
		if len(args) < 1 {
			fmt.Println("Usage: gokku config unset KEY [KEY2...] -a <app>")
			os.Exit(1)
		}

		if err := configService.UnsetEnvVar(appName, args); err != nil {
			fmt.Printf("Error unsetting config: %v\n", err)
			os.Exit(1)
		}

		for _, key := range args {
			fmt.Printf("Unset %s\n", key)
		}
	default:
		fmt.Printf("Unknown subcommand: %s\n", subcommand)
		os.Exit(1)
	}

	// Auto-restart container after set/unset to apply changes
	if subcommand == "set" || subcommand == "unset" {
		fmt.Printf("\n-----> Restarting container to apply changes...\n")

		if err := configService.ReloadApp(appName); err != nil {
			fmt.Printf("Warning: Failed to restart container: %v\n", err)
			fmt.Printf("         Run 'gokku restart -a %s' manually.\n", appName)
		} else {
			fmt.Printf("✓ Container restarted with new configuration\n")
		}
	}
}

func executeAsClientMode(ctx *internal.ExecutionContext, subcommand string, args []string) {
	var cmd string

	if subcommand == "" {
		subcommand = "list"
	}

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
			fmt.Println("Error: KEY is required for config get")
			fmt.Println("Usage: gokku config get KEY -a <app>")
			fmt.Println("")
			fmt.Println("Example:")
			fmt.Printf("  gokku config get PORT -a %s\n", ctx.GetAppName())
			fmt.Println("")
			fmt.Println("To list all config variables, use:")
			fmt.Printf("  gokku config list -a %s\n", ctx.GetAppName())
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

	ctx.PrintConnectionInfo()

	if err := ctx.ExecuteCommand(cmd); err != nil {
		os.Exit(1)
	}
}
