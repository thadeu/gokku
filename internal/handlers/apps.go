package handlers

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"infra/internal"
)

// handleApps manages applications on the server
func handleApps(args []string) {
	subcommand := args[0]
	switch subcommand {
	case "list", "ls":
		handleAppsList(args[1:])
	case "create":
		handleAppsCreate(args[1:])
	case "destroy", "rm":
		handleAppsDestroy(args[1:])
	default:
		fmt.Println("Usage: gokku apps <command> [options]")
		fmt.Println("")
		fmt.Println("Commands:")
		fmt.Println("  list, ls              List all applications")
		fmt.Println("  create <app>          Create application and setup deployment")
		fmt.Println("  destroy, rm <app>     Destroy application")
		fmt.Println("")
		fmt.Println("Options:")
		fmt.Println("  --remote <remote>     Use specific git remote")
		os.Exit(1)
	}
}

// handleAppsList lists applications on the server
func handleAppsList(args []string) {
	remote, remainingArgs := internal.ExtractRemoteFlag(args)

	if len(remainingArgs) < 1 {
		fmt.Println("Usage: gokku apps list [--remote <remote>]")
		os.Exit(1)
	}

	appName := remainingArgs[0]
	remoteInfo, err := internal.GetRemoteInfo(remote)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	config, err := internal.LoadConfig()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	appConfig := config.GetAppConfig(appName)

	if appConfig == nil {
		fmt.Printf("Error: App '%s' not found in gokku.yml\n", appName)
		os.Exit(1)
	}

	// List apps from /opt/gokku/apps directory with detailed information
	cmd := exec.Command("ssh", remoteInfo.Host, fmt.Sprintf(`
		if [ -d "%s/apps" ]; then
			echo "App Name                    Status    Releases    Current Release"
			echo "================================================================"
			ls -1 %s/apps 2>/dev/null | while read app; do
				if [ -d "%s/apps/$app" ]; then
					# Get app status
					if docker ps --format '{{.Names}}' | grep -q "^$app"; then
						status="running"
					elif docker ps -a --format '{{.Names}}' | grep -q "^$app"; then
						status="stopped"
					else
						status="not deployed"
					fi

					# Count releases
					releases_count=0
					if [ -d "%s/apps/$app/releases" ]; then
						releases_count=$(ls -1 %s/apps/$app/releases 2>/dev/null | wc -l)
					fi

					# Get current release
					current_release="none"
					if [ -L "%s/apps/$app/current" ]; then
						current_release=$(basename $(readlink %s/apps/$app/current) 2>/dev/null || echo "none")
					fi

					printf "%%-25s %%-10s %%-10s %%s\n" "$app" "$status" "$releases_count" "$current_release"
				fi
			done
		else
			echo "No apps directory found at %s/apps"
		fi
	`, remoteInfo.BaseDir, remoteInfo.BaseDir, remoteInfo.BaseDir, remoteInfo.BaseDir, remoteInfo.BaseDir, remoteInfo.BaseDir, remoteInfo.BaseDir, remoteInfo.BaseDir))

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("Error listing apps: %v\n", err)
		os.Exit(1)
	}
}

// handleAppsCreate creates an application and sets up deployment
func handleAppsCreate(args []string) {
	remote, remainingArgs := internal.ExtractRemoteFlag(args)

	if len(remainingArgs) < 1 {
		fmt.Println("Usage: gokku apps create <app> [--remote <remote>]")
		fmt.Println("")
		fmt.Println("Examples:")
		fmt.Println("  gokku apps create myapp")
		fmt.Println("  gokku apps create myapp --remote myremote")
		os.Exit(1)
	}

	appName := remainingArgs[0]

	if remote == "" {
		remote = appName
		fmt.Printf("Using remote: %s\n", remote)
	}

	// Parse remote to get connection info
	remoteInfo, err := internal.GetRemoteInfo(remote)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		fmt.Println("")
		fmt.Println("Make sure the git remote exists:")
		fmt.Printf("  git remote add %s user@host:/opt/gokku/repos/%s.git\n", remote, appName)
		os.Exit(1)
	}

	fmt.Printf("Creating app %s on %s...\n", appName, remoteInfo.Host)

	// Load configuration to validate app exists
	config, err := internal.LoadConfig()
	if err != nil {
		fmt.Printf("Warning: Could not load config: %v\n", err)
		fmt.Println("Proceeding with basic setup...")
	}

	// Validate app exists in config if available
	if config != nil {
		if !appExistsInConfig(config, appName) {
			fmt.Printf("Warning: App '%s' not found in gokku.yml\n", appName)
			fmt.Println("Proceeding with basic setup...")
		}
	}

	// Run complete setup directly in Go
	if err := setupAppComplete(remoteInfo, appName, config); err != nil {
		fmt.Printf("Failed to setup app: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✓ App created successfully!")
	fmt.Println("")
	fmt.Println("Next steps:")
	fmt.Println("  1. Make sure your gokku.yml is committed")
	fmt.Println("  2. Deploy with: git push", remote, "main")
}

// handleAppsDestroy destroys an application
func handleAppsDestroy(args []string) {
	remote, remainingArgs := internal.ExtractRemoteFlag(args)

	if len(remainingArgs) < 1 {
		fmt.Println("Usage: gokku apps destroy <app> [--remote <remote>]")
		os.Exit(1)
	}

	appName := remainingArgs[0]

	if remote == "" {
		fmt.Println("Error: --remote is required for destroy command")
		fmt.Println("This prevents accidental deletion of the wrong app")
		os.Exit(1)
	}

	remoteInfo, err := internal.GetRemoteInfo(remote)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Destroying app %s on %s...\n", appName, remoteInfo.Host)
	fmt.Printf("This will permanently delete the app and all its data.\n")
	fmt.Printf("Continue? (y/N): ")

	var response string
	fmt.Scanln(&response)
	if response != "y" && response != "Y" {
		fmt.Println("Aborted.")
		return
	}

	// Remove app directory and repository
	destroyCmd := exec.Command("ssh", remoteInfo.Host, fmt.Sprintf(`
		set -e
		echo "Removing app directory..."
		sudo rm -rf /opt/gokku/apps/%s
		echo "Removing repository..."
		sudo rm -rf /opt/gokku/repos/%s.git
		echo "App destroyed successfully"
	`, appName, appName))

	destroyCmd.Stdout = os.Stdout
	destroyCmd.Stderr = os.Stderr

	if err := destroyCmd.Run(); err != nil {
		fmt.Printf("Failed to destroy app: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✓ App destroyed successfully!")
}

// appExistsInConfig checks if an app exists in the configuration
func appExistsInConfig(config *internal.Config, appName string) bool {
	return config.GetAppConfig(appName) != nil
}

// setupAppComplete performs the complete app setup directly in Go
func setupAppComplete(remoteInfo *internal.RemoteInfo, appName string, config *internal.Config) error {
	fmt.Printf("Setting up app %s on %s...\n", appName, remoteInfo.Host)

	// Get deploy user from SSH connection
	deployUser, err := getDeployUser(remoteInfo.Host)
	if err != nil {
		return fmt.Errorf("failed to get deploy user: %v", err)
	}

	// 1. Create directory structure
	if err := createDirectoryStructure(remoteInfo, appName, deployUser); err != nil {
		return fmt.Errorf("failed to create directory structure: %v", err)
	}

	// 2. Setup git repository
	if err := setupGitRepository(remoteInfo, appName, deployUser); err != nil {
		return fmt.Errorf("failed to setup git repository: %v", err)
	}

	// 3. Setup git namespace and shortcuts
	if err := setupGitNamespace(remoteInfo, appName, deployUser); err != nil {
		return fmt.Errorf("failed to setup git namespace: %v", err)
	}

	// 4. Setup simple hook
	if err := setupSimpleHook(remoteInfo, appName); err != nil {
		return fmt.Errorf("failed to setup hook: %v", err)
	}

	// 6. Create initial .env file
	if err := createInitialEnvFile(remoteInfo, appName); err != nil {
		return fmt.Errorf("failed to create initial .env file: %v", err)
	}

	fmt.Println("✓ Complete setup finished")
	return nil
}

// getDeployUser gets the deploy user from SSH connection
func getDeployUser(host string) (string, error) {
	cmd := exec.Command("ssh", host, "whoami")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// createDirectoryStructure creates the necessary directory structure
func createDirectoryStructure(remoteInfo *internal.RemoteInfo, appName, deployUser string) error {
	fmt.Println("-----> Creating directory structure...")

	cmd := exec.Command("ssh", remoteInfo.Host, fmt.Sprintf(`
		set -e
		echo "Creating base directories..."
		sudo mkdir -p %s/repos
		sudo mkdir -p %s/apps/%s/{releases,shared}
		echo "Setting permissions..."
		sudo chown -R %s:%s %s
		echo "Directory structure created"
	`, remoteInfo.BaseDir, remoteInfo.BaseDir, appName, remoteInfo.BaseDir, deployUser, deployUser, remoteInfo.BaseDir))

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// setupGitRepository initializes the git repository
func setupGitRepository(remoteInfo *internal.RemoteInfo, appName, deployUser string) error {
	fmt.Println("-----> Setting up git repository...")

	repoDir := fmt.Sprintf("%s/repos/%s.git", remoteInfo.BaseDir, appName)

	cmd := exec.Command("ssh", remoteInfo.Host, fmt.Sprintf(`
		set -e
		if [ ! -d "%s/refs" ]; then
			echo "Initializing git repository..."
			cd %s
			sudo git init --bare %s
			sudo chown -R %s %s
			echo "Git repository initialized"
		else
			echo "Git repository already exists"
		fi
	`, repoDir, remoteInfo.BaseDir, repoDir, deployUser, repoDir))

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// setupGitNamespace creates git namespace and shortcuts
func setupGitNamespace(remoteInfo *internal.RemoteInfo, appName, deployUser string) error {
	fmt.Println("-----> Setting up git namespace...")

	cmd := exec.Command("ssh", remoteInfo.Host, fmt.Sprintf(`
		set -e
		USER_HOME=$(eval echo ~%s)

		# Create Git namespace directory if it doesn't exist
		if [ ! -d "$USER_HOME/.git-namespace" ]; then
			sudo -u %s mkdir -p $USER_HOME/.git-namespace
		fi

		# Create symlink for each app (allows short names)
		if [ ! -L "$USER_HOME/%s.git" ]; then
			sudo -u %s ln -sf %s/repos/%s.git $USER_HOME/%s.git
			echo "Created Git shortcut: $USER_HOME/%s.git -> %s/repos/%s.git"
		else
			echo "Git shortcut already exists"
		fi
	`, deployUser, deployUser, appName, deployUser, remoteInfo.BaseDir, appName, appName, appName, remoteInfo.BaseDir, appName))

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// setupSimpleHook creates a simple post-receive hook that delegates to gokku deploy
func setupSimpleHook(remoteInfo *internal.RemoteInfo, appName string) error {
	fmt.Println("-----> Setting up deployment hook...")

	// Simple hook that just calls gokku deploy
	hookContent := fmt.Sprintf(`#!/bin/bash
set -e

APP_NAME="%s"

echo "-----> Deploying $APP_NAME..."

# Execute deployment using the centralized deploy command
gokku deploy "$APP_NAME"

echo "-----> Done"
`, appName)

	// Write hook to server
	cmd := exec.Command("ssh", remoteInfo.Host, fmt.Sprintf(`
		cat > %s/repos/%s.git/hooks/post-receive << 'HOOK_EOF'
%s
HOOK_EOF
		chmod +x %s/repos/%s.git/hooks/post-receive
		echo "Hook configured"
	`, remoteInfo.BaseDir, appName, hookContent, remoteInfo.BaseDir, appName))

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// createInitialEnvFile creates the initial .env file
func createInitialEnvFile(remoteInfo *internal.RemoteInfo, appName string) error {
	fmt.Println("-----> Creating initial .env file...")

	envContent := fmt.Sprintf(`# App: %s
# Generated: %s
ZERO_DOWNTIME=0
`, appName)

	cmd := exec.Command("ssh", remoteInfo.Host, fmt.Sprintf(`
		cat > %s/apps/%s/shared/.env << 'ENV_EOF'
%s
ENV_EOF
		echo "Initial .env file created"
	`, remoteInfo.BaseDir, appName, envContent))

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// autoSetupRepository attempts to create the repository on the server if it doesn't exist
// This is kept for backward compatibility but now uses the complete setup
func autoSetupRepository(remoteInfo *internal.RemoteInfo) error {
	config, _ := internal.LoadConfig()
	return setupAppComplete(remoteInfo, remoteInfo.App, config)
}
