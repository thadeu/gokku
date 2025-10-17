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
		fmt.Println("  parse-app-config <app-name>              Parse app configuration")
		fmt.Println("  get-post-deploy <app-name>               Get post-deploy commands")
		fmt.Println("  get-app-docker-network-mode <app-name>   Get Docker network mode")
		fmt.Println("  get-app-docker-ports <app-name>          Get Docker ports")
		fmt.Println("  get-global-config                         Get global configuration")
		fmt.Println("  validate-config                           Validate gokku.yml")
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
	case "get-app-docker-network-mode":
		if len(args) < 2 {
			fmt.Println("Usage: gokku tool get-app-docker-network-mode <app-name>")
			os.Exit(1)
		}
		handleGetAppDockerNetworkMode(args[1])
	case "get-app-docker-ports":
		if len(args) < 2 {
			fmt.Println("Usage: gokku tool get-app-docker-ports <app-name>")
			os.Exit(1)
		}
		handleGetAppDockerPorts(args[1])
	case "get-global-config":
		handleGetGlobalConfig()
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

func handleGetAppDockerNetworkMode(appName string) {
	config, err := internal.LoadConfig()

	if err != nil {
		fmt.Printf("ERROR: App not found: %v\n", err)
		os.Exit(1)
	}

	appConfig := config.GetAppConfig(appName)

	networkMode := "bridge"

	if appConfig != nil && appConfig.Network != nil && appConfig.Network.Mode != "" {
		networkMode = appConfig.Network.Mode
	}

	fmt.Println(networkMode)
}

func handleGetAppDockerPorts(appName string) {
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

	// Output each port mapping on a separate line
	if app.Build != nil && app.Build.Ports != nil && len(app.Build.Ports) > 0 {
		for _, port := range app.Build.Ports {
			fmt.Println(port)
		}
	}
}

func handleGetGlobalConfig() {
	cfg, err := internal.LoadServerConfig()
	if err != nil {
		fmt.Printf("ERROR: Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Export global configuration as environment variables
	fmt.Printf("export GOKKU_PROJECT_NAME=\"%s\"\n", cfg.Project.Name)
	fmt.Printf("export GOKKU_BASE_DIR=\"%s\"\n", "/opt/gokku")
	fmt.Printf("export GOKKU_BUILD_WORKDIR=\"%s\"\n", cfg.Apps[0].Build.Workdir)
	fmt.Printf("export GOKKU_KEEP_RELEASES=\"%d\"\n", 5)
	fmt.Printf("export GOKKU_PORT_STRATEGY=\"%s\"\n", "manual")
}
