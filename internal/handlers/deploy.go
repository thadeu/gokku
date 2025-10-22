package handlers

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"infra/internal"
	"infra/internal/lang"

	"gopkg.in/yaml.v3"
)

// handleDeploy deploys applications directly or via git push
func handleDeploy(args []string) {
	app, remainingArgs := internal.ExtractAppFlag(args)

	var appName, remoteName string
	var isDirectDeploy bool

	// Check if running on server
	isServerMode := internal.IsServerMode()

	if app != "" {
		if isServerMode {
			// Server mode: -a flag means direct deployment with app name
			appName = app
			isDirectDeploy = true
		} else {
			// Client mode: -a flag means git remote
			remoteInfo, err := internal.GetRemoteInfo(app)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
			appName = remoteInfo.App
			remoteName = app
			isDirectDeploy = false
		}
	} else if len(remainingArgs) >= 1 {
		// Direct deployment (new mode)
		appName = remainingArgs[0]
		isDirectDeploy = true
	} else {
		fmt.Println("Usage: gokku deploy <app>")
		fmt.Println("   or: gokku deploy -a <app>")
		os.Exit(1)
	}

	if isDirectDeploy {
		// Direct deployment - execute deployment logic directly
		fmt.Printf("-----> Deploying %s directly...\n", appName)

		// Check if repository exists and has commits
		baseDir := "/opt/gokku"
		reposDir := filepath.Join(baseDir, "repos", appName+".git")
		if _, err := os.Stat(reposDir); os.IsNotExist(err) {
			fmt.Printf("Error: Repository for app '%s' not found at %s\n", appName, reposDir)
			fmt.Printf("Make sure the repository exists on the server.\n")
			os.Exit(1)
		}

		// Check if repository has commits before attempting deployment
		// For direct deployments, we expect the repository to have commits
		// (user should have pushed code first)
		if !hasCommits(reposDir) {
			fmt.Printf("Error: Repository has no commits yet. You need to push code first.\n")
			fmt.Printf("From your local repository, run:\n")
			fmt.Printf("  git remote add %s user@host:/opt/gokku/repos/%s.git\n", appName, appName)
			fmt.Printf("  git push -u %s main\n", appName)
			os.Exit(1)
		}

		if err := executeDirectDeployment(appName); err != nil {
			fmt.Printf("Deploy failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("\n✓ Deploy complete!")
		return
	}

	// Legacy mode - git push deployment
	fmt.Printf("----->Deploying %s via git push...\n", appName)
	fmt.Printf("Remote: %s\n", remoteName)

	// Check if remote exists
	checkCmd := exec.Command("git", "remote", "get-url", remoteName)
	if err := checkCmd.Run(); err != nil {
		fmt.Printf("Error: git remote '%s' not found\n", remoteName)
		fmt.Println("\nAdd it with:")
		fmt.Printf("  git remote add %s user@host:/opt/gokku/repos/%s.git\n", remoteName, appName)
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
	cmd := exec.Command("git", "push", remoteName, branch)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		// Check if it's a signal interruption
		if internal.IsSignalInterruption(err) {
			fmt.Printf("\nDeploy interrupted by user\n")
			os.Exit(0)
		}
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

	// Check if app exists - if not, we'll create it during initial setup
	appExists := true
	if _, err := os.Stat(appDir); os.IsNotExist(err) {
		appExists = false
	}

	// Check if repository exists
	if _, err := os.Stat(reposDir); os.IsNotExist(err) {
		return fmt.Errorf("repository for app '%s' not found", appName)
	}

	// Create app directory if it doesn't exist
	if !appExists {
		fmt.Printf("-----> App '%s' not found, will be created during initial setup\n", appName)
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
	if err := extractCodeFromRepo(appName, reposDir, releaseDir); err != nil {
		return fmt.Errorf("failed to extract code: %v", err)
	}

	// Copy gokku.yml to app directory if it doesn't exist
	appConfigPath := filepath.Join(appDir, "gokku.yml")
	releaseConfigPath := filepath.Join(releaseDir, "gokku.yml")

	// Check if this is the first deployment and handle initial setup
	if !appExists {
		if _, err := os.Stat(releaseConfigPath); err == nil {
			fmt.Println("-----> Initial setup detected - configuring application...")
			if err := handleInitialSetup(appName, releaseConfigPath, releaseDir); err != nil {
				return fmt.Errorf("failed to setup initial configuration: %v", err)
			}
		} else {
			return fmt.Errorf("gokku.yml not found in release directory - cannot setup app without configuration")
		}
	} else {
		// App exists, just copy config if needed
		if _, err := os.Stat(releaseConfigPath); err == nil {
			if err := copyFile(releaseConfigPath, appConfigPath); err != nil {
				return fmt.Errorf("failed to copy gokku.yml to app directory: %v", err)
			}
		}
	}

	// Load app configuration
	app, err := internal.LoadAppConfig(appName)
	if err != nil {
		return fmt.Errorf("failed to load app config: %v", err)
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
	fmt.Println("-----> Building application...")

	// Force rebuild without cache if this is a Docker build
	if app.Build != nil && app.Build.Type == "docker" {
		dockerfilePath := filepath.Join(releaseDir, "Dockerfile")

		if _, err := os.Stat(dockerfilePath); err == nil {
			fmt.Println("-----> Forcing fresh Docker build (no cache)...")
			// Remove any existing image with the same name
			imageTag := fmt.Sprintf("%s:latest", app.Name)
			exec.Command("docker", "rmi", imageTag).Run() // Ignore errors if image doesn't exist

			// Build with no cache
			cmd := exec.Command("docker", "build", "--no-cache", "-t", imageTag, releaseDir)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				// Check if it's a signal interruption
				if internal.IsSignalInterruption(err) {
					return fmt.Errorf("docker build interrupted by user")
				}
				return fmt.Errorf("docker build failed: %v", err)
			}
		} else {
			// Use language handler for non-Docker builds
			if err := lang.Build(app, releaseDir); err != nil {
				return fmt.Errorf("build failed: %v", err)
			}
		}
	} else {
		// Use language handler for non-Docker builds
		if err := lang.Build(app, releaseDir); err != nil {
			return fmt.Errorf("build failed: %v", err)
		}
	}

	fmt.Println("-----> Build complete!")

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

// hasCommits checks if a git repository has any commits
func hasCommits(repoDir string) bool {
	checkCmd := exec.Command("git", "--git-dir", repoDir, "rev-parse", "--short", "HEAD")
	return checkCmd.Run() == nil
}

// extractCodeFromRepo extracts code from git repository to release directory
func extractCodeFromRepo(appName string, repoDir, releaseDir string) error {
	// Check if repository has any commits
	if !hasCommits(repoDir) {
		return fmt.Errorf("repository has no commits yet - you need to push code first. Run: git push <remote> <branch>")
	}

	// Step 1: Extract only gokku.yml to read configuration
	fmt.Println("-----> Reading app configuration...")
	tmpCmd := exec.Command("git", "--git-dir", repoDir, "show", "HEAD:gokku.yml")
	gokkuYmlContent, err := tmpCmd.Output()
	if err != nil {
		// No gokku.yml in repo, do full checkout
		fmt.Println("-----> No gokku.yml found, extracting full repository...")
		cmd := exec.Command("git", "--git-dir", repoDir, "--work-tree", releaseDir, "checkout", "-f", "HEAD")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("git checkout failed: %v, output: %s", err, string(output))
		}
		return nil
	}

	// Parse config directly from the extracted content
	var serverConfig internal.ServerConfig
	if err := yaml.Unmarshal(gokkuYmlContent, &serverConfig); err != nil {
		fmt.Printf("-----> Error parsing app config: %v, extracting full repository...\n", err)
		cmd := exec.Command("git", "--git-dir", repoDir, "--work-tree", releaseDir, "checkout", "-f", "HEAD")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("git checkout failed: %v, output: %s", err, string(output))
		}
		return nil
	}

	// Find the app in the config
	var app *internal.App
	for _, a := range serverConfig.Apps {
		if a.Name == appName {
			app = &a
			break
		}
	}

	if app == nil {
		fmt.Printf("-----> App '%s' not found in config, extracting full repository...\n", appName)
		cmd := exec.Command("git", "--git-dir", repoDir, "--work-tree", releaseDir, "checkout", "-f", "HEAD")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("git checkout failed: %v, output: %s", err, string(output))
		}
		return nil
	}

	if app.Build == nil || app.Build.Workdir == "" {
		// No workdir specified, do full checkout
		fmt.Println("-----> No workdir specified, extracting full repository...")
		os.RemoveAll(releaseDir) // Clean temp files
		os.MkdirAll(releaseDir, 0755)
		cmd := exec.Command("git", "--git-dir", repoDir, "--work-tree", releaseDir, "checkout", "-f", "HEAD")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("git checkout failed: %v, output: %s", err, string(output))
		}

		// Initial setup is now handled in executeDirectDeployment

		return nil
	}

	workdir := strings.TrimPrefix(app.Build.Workdir, "./")
	workdir = strings.TrimPrefix(workdir, "/")

	fmt.Printf("-----> Workdir configured: '%s'\n", workdir)
	fmt.Printf("-----> Using selective extraction (git archive)\n")

	// Step 2: Extract only what we need using git archive
	// Clean and prepare directory
	os.RemoveAll(releaseDir)
	os.MkdirAll(releaseDir, 0755)

	// Extract gokku.yml and workdir using git archive
	archiveCmd := exec.Command("git", "--git-dir", repoDir, "archive", "HEAD", "gokku.yml", workdir)
	untar := exec.Command("tar", "-x", "-C", releaseDir)

	// Pipe git archive output to tar
	pipe, err := archiveCmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create pipe: %v", err)
	}
	untar.Stdin = pipe

	if err := archiveCmd.Start(); err != nil {
		return fmt.Errorf("failed to start git archive: %v", err)
	}

	if err := untar.Start(); err != nil {
		return fmt.Errorf("failed to start tar: %v", err)
	}

	if err := archiveCmd.Wait(); err != nil {
		return fmt.Errorf("git archive failed: %v", err)
	}

	if err := untar.Wait(); err != nil {
		return fmt.Errorf("tar extraction failed: %v", err)
	}

	// Initial setup is now handled in executeDirectDeployment

	return nil
}

// handleInitialSetup handles the initial setup when gokku.yml is found in the project
func handleInitialSetup(appName string, gokkuYmlPath, releaseDir string) error {
	fmt.Println("-----> Initial setup detected - configuring application...")

	// Create app directories
	appDir := filepath.Join("/opt/gokku", "apps", appName)
	releasesDir := filepath.Join(appDir, "releases")
	sharedDir := filepath.Join(appDir, "shared")

	// Create directories
	if err := os.MkdirAll(releasesDir, 0755); err != nil {
		return fmt.Errorf("failed to create releases directory for %s: %v", appName, err)
	}

	if err := os.MkdirAll(sharedDir, 0755); err != nil {
		return fmt.Errorf("failed to create shared directory for %s: %v", appName, err)
	}

	// Copy gokku.yml to app-specific config location
	appConfigPath := filepath.Join("/opt/gokku", "apps", appName, "gokku.yml")
	if err := copyFile(gokkuYmlPath, appConfigPath); err != nil {
		return fmt.Errorf("failed to copy gokku.yml to app config: %v", err)
	}

	// Also copy gokku.yml to current release directory
	currentConfigPath := filepath.Join(releaseDir, "gokku.yml")
	if err := copyFile(gokkuYmlPath, currentConfigPath); err != nil {
		return fmt.Errorf("failed to copy gokku.yml to current release: %v", err)
	}

	fmt.Printf("-----> Created directories for app '%s'\n", appName)
	fmt.Println("-----> Initial setup complete!")
	return nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	// Create destination directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	// Copy permissions
	sourceInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	return os.Chmod(dst, sourceInfo.Mode())
}

// updateEnvironmentFile updates environment file from gokku.yml if needed
func updateEnvironmentFile(envFile, appName string) error {
	// Load server config to get default env vars
	app, err := internal.LoadAppConfig(appName)

	if err != nil {
		return fmt.Errorf("failed to load server config: %v", err)
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
	app, err := internal.LoadAppConfig(appName)

	if err != nil {
		return fmt.Errorf("failed to load app config: %v", err)
	}

	if app.Deployment == nil || len(app.Deployment.PostDeploy) == 0 {
		return nil
	}

	fmt.Println("-----> Running post-deploy commands...")

	for _, cmdStr := range app.Deployment.PostDeploy {
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
