package handlers

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"infra/internal"
	"infra/internal/lang"
)

// handleDeploy deploys applications directly or via git push
func handleDeploy(args []string) {
	remote, remainingArgs := internal.ExtractRemoteFlag(args)

	var app, remoteName string
	var isDirectDeploy bool

	if remote != "" {
		// Remote deployment via git push (legacy mode)
		remoteInfo, err := internal.GetRemoteInfo(remote)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		app = remoteInfo.App
		remoteName = remote
		isDirectDeploy = false
	} else if len(remainingArgs) >= 1 {
		// Direct deployment (new mode)
		app = remainingArgs[0]
		isDirectDeploy = true
	} else {
		fmt.Println("Usage: gokku deploy <app>")
		fmt.Println("   or: gokku deploy --remote <git-remote>")
		os.Exit(1)
	}

	if isDirectDeploy {
		// Direct deployment - execute deployment logic directly
		fmt.Printf("Deploying %s directly...\n", app)
		if err := executeDirectDeployment(app); err != nil {
			fmt.Printf("Deploy failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("\n✓ Deploy complete!")
		return
	}

	// Legacy mode - git push deployment
	fmt.Printf("Deploying %s via git push...\n", app)
	fmt.Printf("Remote: %s\n", remoteName)

	// Check if remote exists
	checkCmd := exec.Command("git", "remote", "get-url", remoteName)
	if err := checkCmd.Run(); err != nil {
		fmt.Printf("Error: git remote '%s' not found\n", remoteName)
		fmt.Println("\nAdd it with:")
		fmt.Printf("  git remote add %s user@host:/opt/gokku/repos/%s.git\n", remoteName, app)
		os.Exit(1)
	}

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

	// Get current branch
	branchCmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	branchOut, _ := branchCmd.Output()
	branch := strings.TrimSpace(string(branchOut))

	fmt.Printf("Branch: %s\n\n", branch)

	// Push current branch
	cmd := exec.Command("git", "push", branch, remoteName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		fmt.Printf("\nDeploy failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n✓ Deploy complete!")
}

// executeDirectDeployment performs deployment directly without git push
func executeDirectDeployment(appName string) error {
	baseDir := "/opt/gokku"
	appDir := filepath.Join(baseDir, "apps", appName)
	reposDir := filepath.Join(baseDir, "repos", appName+".git")

	// Check if app exists
	if _, err := os.Stat(appDir); os.IsNotExist(err) {
		return fmt.Errorf("app '%s' not found. Create it first with: gokku apps create %s", appName, appName)
	}

	// Check if repository exists
	if _, err := os.Stat(reposDir); os.IsNotExist(err) {
		return fmt.Errorf("repository for app '%s' not found", appName)
	}

	// Create release directory
	releaseTag := time.Now().Format("20060102-150405")
	releaseDir := filepath.Join(appDir, "releases", releaseTag)

	fmt.Printf("-----> Creating release: %s\n", releaseTag)
	if err := os.MkdirAll(releaseDir, 0755); err != nil {
		return fmt.Errorf("failed to create release directory: %v", err)
	}

	// Extract code from git repository
	fmt.Println("-----> Extracting code...")
	if err := extractCodeFromRepo(reposDir, releaseDir); err != nil {
		return fmt.Errorf("failed to extract code: %v", err)
	}

	// Load app configuration
	config, err := internal.LoadServerConfig()
	if err != nil {
		return fmt.Errorf("failed to load server config: %v", err)
	}

	app, err := config.GetApp(appName)
	if err != nil {
		return fmt.Errorf("app not found in config: %v", err)
	}

	// Create language handler
	lang, err := lang.NewLang(app, releaseDir)
	if err != nil {
		return fmt.Errorf("failed to create language handler: %v", err)
	}

	fmt.Printf("-----> Detected language: %s\n", app.Lang)

	// Update environment file if needed
	envFile := filepath.Join(appDir, "shared", ".env")
	if err := updateEnvironmentFile(envFile, appName); err != nil {
		fmt.Printf("Warning: Failed to update environment file: %v\n", err)
	}

	// Link environment file to release
	releaseEnvFile := filepath.Join(releaseDir, ".env")
	if err := os.Symlink(envFile, releaseEnvFile); err != nil {
		return fmt.Errorf("failed to link environment file: %v", err)
	}

	// Update current symlink
	currentLink := filepath.Join(appDir, "current")
	if err := os.Remove(currentLink); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove current symlink: %v", err)
	}
	if err := os.Symlink(releaseDir, currentLink); err != nil {
		return fmt.Errorf("failed to create current symlink: %v", err)
	}

	// Build application using language handler
	if err := lang.Build(app, releaseDir); err != nil {
		return fmt.Errorf("build failed: %v", err)
	}

	// Deploy application using language handler
	if err := lang.Deploy(app, releaseDir); err != nil {
		return fmt.Errorf("deploy failed: %v", err)
	}

	// Cleanup old releases using language handler
	if err := lang.Cleanup(app); err != nil {
		fmt.Printf("Warning: Failed to cleanup old releases: %v\n", err)
	}

	// Execute post-deploy commands
	if err := executePostDeployCommands(appName, releaseDir); err != nil {
		return fmt.Errorf("post-deploy commands failed: %v", err)
	}

	return nil
}

// extractCodeFromRepo extracts code from git repository to release directory
func extractCodeFromRepo(repoDir, releaseDir string) error {
	// Check if repository has any commits
	checkCmd := exec.Command("git", "--git-dir", repoDir, "rev-list", "--count", "HEAD")
	output, err := checkCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to check repository state: %v, output: %s", err, string(output))
	}

	commitCount := strings.TrimSpace(string(output))
	if commitCount == "0" {
		return fmt.Errorf("repository has no commits yet - cannot extract code")
	}

	// Extract code from HEAD
	cmd := exec.Command("git", "--git-dir", repoDir, "--work-tree", releaseDir, "checkout", "-f", "HEAD")
	output, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git checkout failed: %v, output: %s", err, string(output))
	}
	return nil
}

// updateEnvironmentFile updates environment file from gokku.yml if needed
func updateEnvironmentFile(envFile, appName string) error {
	// Load server config to get default env vars
	config, err := internal.LoadServerConfig()
	if err != nil {
		return fmt.Errorf("failed to load server config: %v", err)
	}

	app, err := config.GetApp(appName)
	if err != nil {
		return fmt.Errorf("app not found in config: %v", err)
	}

	// Load existing env vars
	envVars := internal.LoadEnvFile(envFile)

	// Add default env vars from config if not already set
	if app.Environments != nil {
		for _, env := range app.Environments {
			if env.DefaultEnvVars != nil {
				for key, value := range env.DefaultEnvVars {
					if _, exists := envVars[key]; !exists {
						envVars[key] = value
					}
				}
			}
		}
	}

	// Save updated env vars
	return internal.SaveEnvFile(envFile, envVars)
}

// executePostDeployCommands runs post-deploy commands if configured
func executePostDeployCommands(appName, releaseDir string) error {
	// Get post-deploy commands using tool command
	cmd := exec.Command("gokku", "tool", "get-post-deploy", appName)
	output, err := cmd.Output()
	if err != nil {
		// No post-deploy commands configured
		return nil
	}

	commands := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(commands) == 0 || (len(commands) == 1 && commands[0] == "") {
		return nil
	}

	fmt.Println("-----> Running post-deploy commands...")

	for _, cmdStr := range commands {
		if strings.TrimSpace(cmdStr) == "" {
			continue
		}

		fmt.Printf("       Running: %s\n", cmdStr)

		// Execute command in release directory
		cmd := exec.Command("bash", "-c", cmdStr)
		cmd.Dir = releaseDir
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("post-deploy command failed '%s': %v, output: %s", cmdStr, err, string(output))
		}
	}

	fmt.Println("-----> Post-deploy commands completed")
	return nil
}
