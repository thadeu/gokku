package commands

import (
	"encoding/json"
	"fmt"
	"os"

	"gokku/internal"
)

func useTool(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: gokku tool <command> [args...]")
		fmt.Println("Commands:")
		fmt.Println("  parse-app-config <app-name>              Parse app configuration")
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
	default:
		fmt.Printf("Unknown internal command: %s\n", command)
		os.Exit(1)
	}
}

func handleParseAppConfig(appName string) {
	app, err := internal.LoadAppConfig(appName)

	if err != nil {
		fmt.Printf("ERROR: Failed to load config: %v\n", err)
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
