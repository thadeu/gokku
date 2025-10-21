package handlers

import (
	"fmt"
	"os"
	"strings"

	"infra/internal"
)

func handleRun(args []string) {
	fmt.Println("Error: This handler should not be called directly")
	os.Exit(1)
}

func handleRunWithContext(ctx *internal.ExecutionContext, args []string) {
	if err := ctx.ValidateAppRequired(); err != nil {
		ctx.PrintUsageError("run", err.Error())
	}

	_, remainingArgs := internal.ExtractAppFlag(args)

	if len(remainingArgs) < 1 {
		fmt.Println("Error: command is required")
		fmt.Println("Usage: gokku run <command> -a <app>")
		os.Exit(1)
	}

	command := strings.Join(remainingArgs, " ")

	containerName := ctx.GetAppName()
	dockerCommand := fmt.Sprintf("docker exec -it %s %s", containerName, command)

	ctx.PrintConnectionInfo()
	fmt.Printf("$ %s\n\n", command)

	// Execute command
	if err := ctx.ExecuteCommand(dockerCommand); err != nil {
		os.Exit(1)
	}
}
