package handlers

import (
	"fmt"
	"os"
	"os/exec"

	"infra/internal"
)

// handleDeploy deploys applications via git push
func handleDeploy(args []string) {
	remote, remainingArgs := internal.ExtractRemoteFlag(args)

	var app, env, remoteName string

	if remote != "" {
		remoteInfo, err := internal.GetRemoteInfo(remote)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		app = remoteInfo.App
		env = remoteInfo.Env
		remoteName = remote
	} else if len(remainingArgs) >= 2 {
		app = remainingArgs[0]
		env = remainingArgs[1]
		remoteName = fmt.Sprintf("%s-%s", app, env)
	} else {
		fmt.Println("Usage: gokku deploy <app> <env>")
		fmt.Println("   or: gokku deploy --remote <git-remote>")
		os.Exit(1)
	}

	// Determine branch based on environment
	branch := "main"
	if env == "staging" {
		branch = "staging"
	} else if env == "develop" {
		branch = "develop"
	}

	// Check if remote exists
	checkCmd := exec.Command("git", "remote", "get-url", remoteName)
	if err := checkCmd.Run(); err != nil {
		fmt.Printf("Error: git remote '%s' not found\n", remoteName)
		fmt.Println("\nAdd it with:")
		fmt.Printf("  git remote add %s user@host:/opt/gokku/repos/%s.git\n", remoteName, app)
		os.Exit(1)
	}

	fmt.Printf("Deploying %s to %s environment...\n", app, env)
	fmt.Printf("Remote: %s\n", remoteName)
	fmt.Printf("Branch: %s\n\n", branch)

	// Get remote info for auto-setup
	remoteInfo, err := internal.GetRemoteInfo(remoteName)
	if err != nil {
		fmt.Printf("Warning: Could not parse remote info: %v\n", err)
		fmt.Println("Proceeding without auto-setup...")
	} else {
		// Try to setup repository automatically
		if err := autoSetupRepository(remoteInfo); err != nil {
			fmt.Printf("Warning: Auto-setup failed: %v\n", err)
			fmt.Println("Repository may not exist on server yet.")
		}
	}

	// Push
	cmd := exec.Command("git", "push", remoteName, branch)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		fmt.Printf("\nDeploy failed: %v\n", err)
		fmt.Println("\nIf repository doesn't exist, run:")
		fmt.Printf("  gokku server setup %s %s --remote %s\n", app, env, remoteName)
		os.Exit(1)
	}

	fmt.Println("\n✓ Deploy complete!")
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

	// Create repository on server
	setupCmd := exec.Command("ssh", remoteInfo.Host, fmt.Sprintf(`
		sudo mkdir -p /opt/gokku/repos/%s.git && \
		sudo git init --bare /opt/gokku/repos/%s.git && \
		sudo chown -R %s /opt/gokku/repos/%s.git && \
		echo "Repository created successfully"
	`, remoteInfo.App, remoteInfo.App, os.Getenv("USER"), remoteInfo.App))

	setupCmd.Stdout = os.Stdout
	setupCmd.Stderr = os.Stderr

	if err := setupCmd.Run(); err != nil {
		return fmt.Errorf("failed to create repository: %v", err)
	}

	// Copy the smart hook to the server
	hookPath := "/opt/gokku/hooks/post-receive-systemd.template"
	copyCmd := exec.Command("scp", hookPath, fmt.Sprintf("%s:/tmp/post-receive-template", remoteInfo.Host))
	if err := copyCmd.Run(); err != nil {
		return fmt.Errorf("failed to copy hook template: %v", err)
	}

	// Configure the hook on server
	configCmd := exec.Command("ssh", remoteInfo.Host, fmt.Sprintf(`
		cp /tmp/post-receive-template /opt/gokku/repos/%s.git/hooks/post-receive && \
		chmod +x /opt/gokku/repos/%s.git/hooks/post-receive && \
		echo "Hook installed successfully"
	`, remoteInfo.App, remoteInfo.App))

	configCmd.Stdout = os.Stdout
	configCmd.Stderr = os.Stderr

	if err := configCmd.Run(); err != nil {
		return fmt.Errorf("failed to configure hook: %v", err)
	}

	fmt.Println("✓ Repository auto-setup complete")
	return nil
}
