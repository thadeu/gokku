package handlers

import (
	"encoding/json"
	"fmt"
	"os"

	"infra/internal"
)

func handleTool(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: gokku tool <command> [args...]")
		fmt.Println("Commands:")
		fmt.Println("  parse-app-config <app-name>    Parse app configuration")
		fmt.Println("  get-post-deploy <app-name>     Get post-deploy commands")
		fmt.Println("  validate-config                Validate gokku.yml")
		os.Exit(1)
	}

	command := args[0]
	switch command {
	case "parse-app-config":
		if len(args) < 2 {
			fmt.Println("Usage: gokku internal parse-app-config <app-name>")
			os.Exit(1)
		}
		handleParseAppConfig(args[1])
	case "get-post-deploy":
		if len(args) < 2 {
			fmt.Println("Usage: gokku internal get-post-deploy <app-name>")
			os.Exit(1)
		}
		handleGetPostDeploy(args[1])
	case "validate-config":
		handleValidateConfig()
	default:
		fmt.Printf("Unknown internal command: %s\n", command)
		os.Exit(1)
	}
}

func handleParseAppConfig(appName string) {
	cfg, err := internal.LoadServerConfig()
	if err != nil {
		fmt.Printf("ERROR: Failed to load config: %v\n", err)
		os.Exit(1)
	}

	app, err := cfg.GetApp(appName)
	if err != nil {
		fmt.Printf("ERROR: App not found: %v\n", err)
		os.Exit(1)
	}

	// Output as JSON for easy parsing by bash
	jsonData, err := json.MarshalIndent(app, "", "  ")
	if err != nil {
		fmt.Printf("ERROR: Failed to marshal app config: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(jsonData))
}

func handleGetPostDeploy(appName string) {
	cfg, err := internal.LoadServerConfig()
	if err != nil {
		fmt.Printf("ERROR: Failed to load config: %v\n", err)
		os.Exit(1)
	}

	app, err := cfg.GetApp(appName)
	if err != nil {
		fmt.Printf("ERROR: App not found: %v\n", err)
		os.Exit(1)
	}

	if app.Deployment == nil || len(app.Deployment.PostDeploy) == 0 {
		// No post-deploy commands, exit silently
		return
	}

	// Output each command on a separate line
	for _, cmd := range app.Deployment.PostDeploy {
		fmt.Println(cmd)
	}
}

func handleValidateConfig() {
	cfg, err := internal.LoadServerConfig()
	if err != nil {
		fmt.Printf("ERROR: Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Validate the configuration
	if err := cfg.Validate(); err != nil {
		fmt.Printf("ERROR: Invalid configuration: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Configuration is valid")
}
