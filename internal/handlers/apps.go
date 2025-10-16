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
	if len(args) == 0 {
		handleAppsList()
		return
	}

	subcommand := args[0]
	switch subcommand {
	case "list", "ls":
		handleAppsList()
	case "create":
		handleAppsCreate(args[1:])
	case "destroy", "rm":
		handleAppsDestroy(args[1:])
	default:
		fmt.Println("Usage: gokku apps <command> [options]")
		fmt.Println("")
		fmt.Println("Commands:")
		fmt.Println("  list, ls              List all applications")
		fmt.Println("  create <app> [env]    Create application and setup deployment")
		fmt.Println("  destroy, rm <app>     Destroy application")
		fmt.Println("")
		fmt.Println("Options:")
		fmt.Println("  --remote <remote>     Use specific git remote")
		os.Exit(1)
	}
}

// handleAppsList lists applications on the server
func handleAppsList() {
	config, err := internal.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	server := internal.GetDefaultServer(config)
	if server == nil {
		fmt.Println("No servers configured")
		fmt.Println("Add a server: gokku server add production ubuntu@ec2.compute.amazonaws.com")
		os.Exit(1)
	}

	fmt.Printf("Listing apps on %s...\n", server.Name)

	cmd := exec.Command("ssh", server.Host, fmt.Sprintf("ls -1 %s/repos 2>/dev/null | sed 's/.git//'", server.BaseDir))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

// handleAppsCreate creates an application and sets up deployment
func handleAppsCreate(args []string) {
	remote, remainingArgs := internal.ExtractRemoteFlag(args)

	if len(remainingArgs) < 1 {
		fmt.Println("Usage: gokku apps create <app> [environment] [--remote <remote>]")
		fmt.Println("")
		fmt.Println("Examples:")
		fmt.Println("  gokku apps create myapp                    # uses default environment")
		fmt.Println("  gokku apps create myapp production         # explicit environment")
		fmt.Println("  gokku apps create myapp --remote myremote  # uses git remote")
		os.Exit(1)
	}

	appName := remainingArgs[0]
	envName := "production" // default

	if len(remainingArgs) >= 2 {
		envName = remainingArgs[1]
	}

	if remote == "" {
		// Try to find a remote that matches the app name pattern
		remote = fmt.Sprintf("%s-%s", appName, envName)
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

	fmt.Printf("Creating app %s (%s) on %s...\n", appName, envName, remoteInfo.Host)

	// Setup repository automatically
	if err := autoSetupRepository(remoteInfo); err != nil {
		fmt.Printf("Failed to setup repository: %v\n", err)
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

// autoSetupRepository attempts to create the repository on the server if it doesn't exist
func autoSetupRepository(remoteInfo *internal.RemoteInfo) error {
	fmt.Printf("Checking repository status on %s...\n", remoteInfo.Host)

	// Test SSH connection and check if repo exists
	testCmd := exec.Command("ssh", remoteInfo.Host, fmt.Sprintf("test -d /opt/gokku/repos/%s.git", remoteInfo.App))
	if err := testCmd.Run(); err == nil {
		fmt.Println("✓ Repository exists")
		return nil
	}

	fmt.Printf("Repository doesn't exist, creating /opt/gokku/repos/%s.git...\n", remoteInfo.App)

	// Get the user from SSH connection (more reliable than os.Getenv)
	userCmd := exec.Command("ssh", remoteInfo.Host, "whoami")
	userOutput, err := userCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get user from server: %v", err)
	}
	serverUser := strings.TrimSpace(string(userOutput))

	// Create repository on server
	setupCmd := exec.Command("ssh", remoteInfo.Host, fmt.Sprintf(`
		set -e
		echo "Creating repository directory..."
		sudo mkdir -p /opt/gokku/repos
		sudo mkdir -p /opt/gokku/repos/%s.git
		echo "Initializing git repository..."
		sudo git init --bare /opt/gokku/repos/%s.git
		echo "Setting permissions..."
		sudo chown -R %s /opt/gokku/repos/%s.git
		echo "Repository created successfully"
	`, remoteInfo.App, remoteInfo.App, serverUser, remoteInfo.App))

	setupCmd.Stdout = os.Stdout
	setupCmd.Stderr = os.Stderr

	if err := setupCmd.Run(); err != nil {
		return fmt.Errorf("failed to create repository: %v", err)
	}

	// Copy the smart hook from local Gokku installation
	// First try to find the hook in the local Gokku installation
	localHookPaths := []string{
		"./hooks/post-receive-systemd.template",                      // If running from source
		"/usr/local/share/gokku/hooks/post-receive-systemd.template", // Standard install location
		"/opt/gokku/hooks/post-receive-systemd.template",             // Alternative location
	}

	var hookPath string
	for _, path := range localHookPaths {
		if _, err := os.Stat(path); err == nil {
			hookPath = path
			break
		}
	}

	if hookPath == "" {
		// Create a basic hook inline if template not found
		fmt.Println("Hook template not found locally, creating basic hook...")
		configCmd := exec.Command("ssh", remoteInfo.Host, fmt.Sprintf(`
			cat > /opt/gokku/repos/%s.git/hooks/post-receive << 'EOF'
#!/bin/bash
echo "==> Gokku post-receive hook executed"
echo "==> Auto-setup will be handled by push logic"
EOF
			chmod +x /opt/gokku/repos/%s.git/hooks/post-receive
			echo "Basic hook created"
		`, remoteInfo.App, remoteInfo.App))

		configCmd.Stdout = os.Stdout
		configCmd.Stderr = os.Stderr

		if err := configCmd.Run(); err != nil {
			return fmt.Errorf("failed to create basic hook: %v", err)
		}
	} else {
		// Copy the actual hook template
		copyCmd := exec.Command("scp", hookPath, fmt.Sprintf("%s:/opt/gokku/repos/%s.git/hooks/post-receive", remoteInfo.Host, remoteInfo.App))
		copyCmd.Stdout = os.Stdout
		copyCmd.Stderr = os.Stderr

		if err := copyCmd.Run(); err != nil {
			return fmt.Errorf("failed to copy hook template: %v", err)
		}

		// Make it executable
		execCmd := exec.Command("ssh", remoteInfo.Host, fmt.Sprintf("chmod +x /opt/gokku/repos/%s.git/hooks/post-receive", remoteInfo.App))
		if err := execCmd.Run(); err != nil {
			return fmt.Errorf("failed to make hook executable: %v", err)
		}
	}

	fmt.Println("✓ Repository auto-setup complete")
	return nil
}
