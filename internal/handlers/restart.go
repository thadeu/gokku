package handlers

import (
	"fmt"
	"os"
	"os/exec"

	"infra/internal"
)

// handleRestart restarts services/containers
func handleRestart(args []string) {
	remote, remainingArgs := internal.ExtractRemoteFlag(args)

	var app, env, host string
	var localExecution bool

	if remote != "" {
		remoteInfo, err := internal.GetRemoteInfo(remote)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		app = remoteInfo.App
		env = remoteInfo.Env
		host = remoteInfo.Host
	} else if len(remainingArgs) >= 2 {
		// Check if running on server - allow local execution
		if internal.IsRunningOnServer() {
			localExecution = true
			app = remainingArgs[0]
			env = remainingArgs[1]
		} else {
			// Client without --remote - show error
			fmt.Println("Error: Local restart commands can only be run on the server")
			fmt.Println("")
			fmt.Println("For client usage, use --remote flag:")
			fmt.Println("  gokku restart <app> <env> --remote <git-remote>")
			fmt.Println("")
			fmt.Println("Or run this command directly on your server.")
			os.Exit(1)
		}
	} else {
		fmt.Println("Usage: gokku restart <app> <env>")
		fmt.Println("   or: gokku restart --remote <git-remote>")
		os.Exit(1)
	}

	serviceName := fmt.Sprintf("%s-%s", app, env)
	fmt.Printf("Restarting %s...\n", serviceName)

	if localExecution {
		// Local execution on server - recreate container with new env
		baseDir := "/opt/gokku"
		if envVar := os.Getenv("GOKKU_BASE_DIR"); envVar != "" {
			baseDir = envVar
		}

		envFile := fmt.Sprintf("%s/apps/%s/%s/shared/.env", baseDir, app, env)
		appDir := fmt.Sprintf("%s/apps/%s/%s", baseDir, app, env)

		// Source docker-helpers and recreate container
		restartScript := fmt.Sprintf(`
			source /opt/gokku/scripts/docker-helpers.sh
			recreate_active_container "%s" "%s" "%s"
		`, app, envFile, appDir)

		cmd := exec.Command("bash", "-c", restartScript)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Printf("Error recreating container: %v\n", err)
			os.Exit(1)
		}
	} else {
		// Remote execution via SSH - recreate container with new env
		sshCmd := fmt.Sprintf(`
			source /opt/gokku/scripts/docker-helpers.sh
			ENV_FILE="/opt/gokku/apps/%s/%s/shared/.env"
			APP_DIR="/opt/gokku/apps/%s/%s"
			recreate_active_container "%s" "$ENV_FILE" "$APP_DIR"
		`, app, env, app, env, app)

		cmd := exec.Command("ssh", host, sshCmd)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()
	}
}
