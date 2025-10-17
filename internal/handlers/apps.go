package handlers

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

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

	// Check if we're running locally on the server (in /opt/gokku)
	if isRunningOnServer() {
		fmt.Printf("Running on server - creating app '%s' locally...\n", appName)
		if err := createAppLocally(appName); err != nil {
			fmt.Printf("Failed to create app locally: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("✓ App created successfully on server!")
		return
	}

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
			fmt.Println("This is normal if you haven't pushed your code yet.")
			fmt.Println("The app will be configured automatically on first deployment.")
			fmt.Println("Proceeding with basic repository setup...")
		}
	} else {
		fmt.Println("No local gokku.yml found. This is normal if you're setting up a new project.")
		fmt.Println("Make sure to commit your gokku.yml and Dockerfile before pushing.")
		fmt.Println("Proceeding with basic repository setup...")
	}

	// Run complete setup directly in Go
	if err := setupAppComplete(remoteInfo, appName, config); err != nil {
		fmt.Printf("Failed to setup app: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✓ App created successfully!")
	fmt.Println("")
	fmt.Println("Next steps:")
	fmt.Println("  1. Make sure your gokku.yml and Dockerfile are committed")
	fmt.Println("  2. Deploy with: git push", remote, "main")
	fmt.Println("     (This will automatically configure the app from your gokku.yml)")
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

	return setupAppCore(remoteInfo, appName, deployUser, true) // true = remote
}

// setupAppLocally creates an application directly on the server
func createAppLocally(appName string) error {
	fmt.Printf("Setting up app %s locally...\n", appName)

	baseDir := "/opt/gokku"

	// Check if we have write permissions to /opt/gokku before proceeding
	if !hasWritePermission(baseDir) {
		fmt.Printf("-----> No write permission to %s\n", baseDir)
		fmt.Printf("-----> This command requires write access to %s\n", baseDir)
		fmt.Printf("-----> Try running with sudo: sudo gokku apps create %s\n", appName)

		// Try to create directories with sudo
		if err := createDirectoriesWithSudo(appName); err != nil {
			return fmt.Errorf("failed to create directories with sudo: %v", err)
		}

		fmt.Println("-----> Directories created successfully with sudo")
		return nil
	}

	// For local setup, we use the current user
	currentUser := os.Getenv("USER")
	if currentUser == "" {
		currentUser = "gokku"
	}

	return setupAppCore(nil, appName, currentUser, false) // false = local
}

// hasWritePermission checks if the current user can write to the specified directory
func hasWritePermission(dir string) bool {
	testFile := filepath.Join(dir, ".gokku-permission-test")
	err := os.WriteFile(testFile, []byte("test"), 0644)
	if err != nil {
		return false
	}
	os.Remove(testFile)
	return true
}

// createDirectoriesWithSudo creates the necessary directories using sudo
func createDirectoriesWithSudo(appName string) error {
	baseDir := "/opt/gokku"

	// Get current user for proper ownership
	currentUser := os.Getenv("USER")
	if currentUser == "" {
		currentUser = "gokku"
	}

	// Create all directories with sudo
	cmd := exec.Command("sudo", "mkdir", "-p",
		filepath.Join(baseDir, "repos"),
		filepath.Join(baseDir, "apps", appName, "releases"),
		filepath.Join(baseDir, "apps", appName, "shared"))

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("sudo mkdir failed: %v (output: %s)", err, string(output))
	}

	// Fix ownership of created directories
	ownershipCmd := exec.Command("sudo", "chown", "-R", fmt.Sprintf("%s:%s", currentUser, currentUser),
		filepath.Join(baseDir, "repos"),
		filepath.Join(baseDir, "apps", appName))

	if output, err := ownershipCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("sudo chown failed: %v (output: %s)", err, string(output))
	}

	// Initialize git repository with sudo
	repoDir := filepath.Join(baseDir, "repos", appName+".git")
	gitCmd := exec.Command("sudo", "git", "init", "--bare", repoDir, "--initial-branch=main")
	if output, err := gitCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("sudo git init failed: %v (output: %s)", err, string(output))
	}

	// Fix ownership of git repository
	repoOwnershipCmd := exec.Command("sudo", "chown", "-R", fmt.Sprintf("%s:%s", currentUser, currentUser), repoDir)
	if output, err := repoOwnershipCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("sudo chown repo failed: %v (output: %s)", err, string(output))
	}

	// Create .env file with sudo
	envFile := filepath.Join(baseDir, "apps", appName, "shared", ".env")
	envContent := fmt.Sprintf(`# App: %s
# Generated: %s
ZERO_DOWNTIME=0
`, appName, time.Now().Format("2006-01-02 15:04:05"))

	envCmd := exec.Command("sudo", "bash", "-c", fmt.Sprintf("cat > %s << 'EOF'\n%s\nEOF", envFile, envContent))
	if output, err := envCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("sudo env file creation failed: %v (output: %s)", err, string(output))
	}

	// Create post-receive hook with sudo
	hookDir := filepath.Join(repoDir, "hooks")
	hookFile := filepath.Join(hookDir, "post-receive")
	hookContent := fmt.Sprintf(`#!/bin/bash
set -e

APP_NAME="%s"

echo "-----> Received push for $APP_NAME"

# Check if repository has commits
if git rev-parse --verify HEAD >/dev/null 2>&1; then
    echo "-----> Repository has commits"

    # Get the current HEAD branch
    CURRENT_HEAD_REF=$(git symbolic-ref HEAD 2>/dev/null || echo "")
    CURRENT_HEAD_BRANCH=$(basename "$CURRENT_HEAD_REF" 2>/dev/null || echo "")

    echo "-----> Current HEAD points to: $CURRENT_HEAD_BRANCH"

    # Check if the current HEAD branch has commits
    if ! git log --oneline -1 "$CURRENT_HEAD_BRANCH" >/dev/null 2>&1; then
        echo "-----> Current HEAD branch '$CURRENT_HEAD_BRANCH' has no commits"
        echo "-----> Looking for branch with commits to set as HEAD..."

        # Find the branch that was just pushed (from stdin)
        PUSHED_BRANCH=""
        while read oldrev newrev refname; do
            if [[ "$newrev" != "0000000000000000000000000000000000000000" ]]; then
                branch=$(basename "$refname")
                echo "-----> Detected push to branch: $branch"
                if git log --oneline -1 "$branch" >/dev/null 2>&1; then
                    PUSHED_BRANCH="$branch"
                    break
                fi
            fi
        done

        # If we found a pushed branch with commits, switch HEAD to it
        if [[ -n "$PUSHED_BRANCH" ]]; then
            echo "-----> Switching HEAD from $CURRENT_HEAD_BRANCH to $PUSHED_BRANCH"
            git symbolic-ref HEAD "refs/heads/$PUSHED_BRANCH"
            CURRENT_HEAD_BRANCH="$PUSHED_BRANCH"
        else
            echo "-----> No suitable branch found with commits"
        fi
    else
        echo "-----> Current HEAD branch '$CURRENT_HEAD_BRANCH' has commits"
    fi

    echo "-----> Deploying from branch: $CURRENT_HEAD_BRANCH"
    echo "-----> Deploying $APP_NAME..."

    # Execute deployment using the centralized deploy command
    gokku deploy "$APP_NAME"

    echo "-----> Deployment completed"
else
    echo "-----> Repository is empty, skipping deployment"
    echo "-----> Run 'gokku deploy $APP_NAME' manually after your first push"
fi

echo "-----> Done"
`, appName)

	hookCmd := exec.Command("sudo", "bash", "-c", fmt.Sprintf(`
		mkdir -p %s
		cat > %s << 'EOF'
%s
EOF
		chmod +x %s
		chown %s:%s %s
	`, hookDir, hookFile, hookContent, hookFile, currentUser, currentUser, hookFile))

	if output, err := hookCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("sudo hook creation failed: %v (output: %s)", err, string(output))
	}

	return nil
}

// setupAppCore performs the core app setup logic (shared between local and remote)
func setupAppCore(remoteInfo *internal.RemoteInfo, appName, deployUser string, isRemote bool) error {
	if isRemote {
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

		// 5. Create initial .env file
		if err := createInitialEnvFile(remoteInfo, appName); err != nil {
			return fmt.Errorf("failed to create initial .env file: %v", err)
		}
	} else {
		// Local setup
		baseDir := "/opt/gokku"

		fmt.Println("-----> Creating directory structure...")

		// Create app directories
		appDir := filepath.Join(baseDir, "apps", appName)
		releasesDir := filepath.Join(appDir, "releases")
		sharedDir := filepath.Join(appDir, "shared")

		if err := os.MkdirAll(releasesDir, 0755); err != nil {
			return fmt.Errorf("failed to create releases directory: %v", err)
		}
		if err := os.MkdirAll(sharedDir, 0755); err != nil {
			return fmt.Errorf("failed to create shared directory: %v", err)
		}

		// Create repos directory if it doesn't exist
		reposDir := filepath.Join(baseDir, "repos")
		if err := os.MkdirAll(reposDir, 0755); err != nil {
			return fmt.Errorf("failed to create repos directory: %v", err)
		}

		fmt.Println("-----> Setting up git repository...")

		// Check if git is available
		if _, err := exec.LookPath("git"); err != nil {
			return fmt.Errorf("git command not found: %v", err)
		}
		fmt.Println("-----> Git command found")

		// Initialize git repository
		repoDir := filepath.Join(reposDir, appName+".git")
		fmt.Printf("-----> Repository directory: %s\n", repoDir)

		if _, err := os.Stat(filepath.Join(repoDir, "HEAD")); os.IsNotExist(err) {
			fmt.Printf("-----> Initializing git repository at %s\n", repoDir)

			// Remove directory if it exists but is not a git repo
			if _, err := os.Stat(repoDir); err == nil {
				fmt.Printf("-----> Removing existing directory: %s\n", repoDir)
				if err := os.RemoveAll(repoDir); err != nil {
					return fmt.Errorf("failed to remove existing directory: %v", err)
				}
			}

			// Try to create the directory first with proper permissions
			if err := os.MkdirAll(repoDir, 0755); err != nil {
				return fmt.Errorf("failed to create repository directory %s: %v", repoDir, err)
			}
			fmt.Printf("-----> Repository directory created: %s\n", repoDir)

			cmd := exec.Command("git", "init", "--bare", repoDir, "--initial-branch=main")
			output, err := cmd.CombinedOutput()
			if err != nil {
				fmt.Printf("-----> Git command failed. Output: %s\n", string(output))
				fmt.Printf("-----> Error: %v\n", err)

				// Provide helpful suggestions based on the error
				if strings.Contains(string(output), "Permission denied") {
					return fmt.Errorf("failed to initialize git repository: %v (output: %s). "+
						"Try running with sudo: sudo gokku apps create %s", err, string(output), appName)
				}
				return fmt.Errorf("failed to initialize git repository: %v (output: %s)", err, string(output))
			}
			fmt.Println("-----> Git repository initialized successfully")
		} else {
			fmt.Println("-----> Git repository already exists")
		}

		fmt.Println("-----> Creating initial .env file...")

		// Create initial .env file
		envFile := filepath.Join(sharedDir, ".env")
		envContent := fmt.Sprintf(`# App: %s
# Generated: %s
ZERO_DOWNTIME=0
`, appName, time.Now().Format("2006-01-02 15:04:05"))

		if err := os.WriteFile(envFile, []byte(envContent), 0644); err != nil {
			return fmt.Errorf("failed to create .env file: %v", err)
		}

		fmt.Println("-----> Setting up deployment hook...")

		// Create post-receive hook
		hookDir := filepath.Join(repoDir, "hooks")
		if err := os.MkdirAll(hookDir, 0755); err != nil {
			return fmt.Errorf("failed to create hooks directory: %v", err)
		}

		hookContent := fmt.Sprintf(`#!/bin/bash
set -e

APP_NAME="%s"

echo "-----> Received push for $APP_NAME"

# Check if repository has commits
if git rev-parse --verify HEAD >/dev/null 2>&1; then
    echo "-----> Repository has commits"

    # Get the current HEAD branch
    CURRENT_HEAD_REF=$(git symbolic-ref HEAD 2>/dev/null || echo "")
    CURRENT_HEAD_BRANCH=$(basename "$CURRENT_HEAD_REF" 2>/dev/null || echo "")

    echo "-----> Current HEAD points to: $CURRENT_HEAD_BRANCH"

    # Check if the current HEAD branch has commits
    if ! git log --oneline -1 "$CURRENT_HEAD_BRANCH" >/dev/null 2>&1; then
        echo "-----> Current HEAD branch '$CURRENT_HEAD_BRANCH' has no commits"
        echo "-----> Looking for branch with commits to set as HEAD..."

        # Find the branch that was just pushed (from stdin)
        PUSHED_BRANCH=""
        while read oldrev newrev refname; do
            if [[ "$newrev" != "0000000000000000000000000000000000000000" ]]; then
                branch=$(basename "$refname")
                echo "-----> Detected push to branch: $branch"
                if git log --oneline -1 "$branch" >/dev/null 2>&1; then
                    PUSHED_BRANCH="$branch"
                    break
                fi
            fi
        done

        # If we found a pushed branch with commits, switch HEAD to it
        if [[ -n "$PUSHED_BRANCH" ]]; then
            echo "-----> Switching HEAD from $CURRENT_HEAD_BRANCH to $PUSHED_BRANCH"
            git symbolic-ref HEAD "refs/heads/$PUSHED_BRANCH"
            CURRENT_HEAD_BRANCH="$PUSHED_BRANCH"
        else
            echo "-----> No suitable branch found with commits"
        fi
    else
        echo "-----> Current HEAD branch '$CURRENT_HEAD_BRANCH' has commits"
    fi

    echo "-----> Deploying from branch: $CURRENT_HEAD_BRANCH"
    echo "-----> Deploying $APP_NAME..."

    # Execute deployment using the centralized deploy command
    gokku deploy "$APP_NAME"

    echo "-----> Deployment completed"
else
    echo "-----> Repository is empty, skipping deployment"
    echo "-----> Run 'gokku deploy $APP_NAME' manually after your first push"
fi

echo "-----> Done"
`, appName)

		hookFile := filepath.Join(hookDir, "post-receive")
		if err := os.WriteFile(hookFile, []byte(hookContent), 0755); err != nil {
			return fmt.Errorf("failed to create post-receive hook: %v", err)
		}
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
	`, remoteInfo.BaseDir, remoteInfo.BaseDir, appName, deployUser, deployUser, remoteInfo.BaseDir))

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
			sudo git init --bare --initial-branch=main %s
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

echo "-----> Received push for $APP_NAME"

# Check if repository has commits
if git rev-parse --verify HEAD >/dev/null 2>&1; then
    echo "-----> Repository has commits"

    # Get the current HEAD branch
    CURRENT_HEAD_REF=$(git symbolic-ref HEAD 2>/dev/null || echo "")
    CURRENT_HEAD_BRANCH=$(basename "$CURRENT_HEAD_REF" 2>/dev/null || echo "")

    echo "-----> Current HEAD points to: $CURRENT_HEAD_BRANCH"

    # Check if the current HEAD branch has commits
    if ! git log --oneline -1 "$CURRENT_HEAD_BRANCH" >/dev/null 2>&1; then
        echo "-----> Current HEAD branch '$CURRENT_HEAD_BRANCH' has no commits"
        echo "-----> Looking for branch with commits to set as HEAD..."

        # Find the branch that was just pushed (from stdin)
        PUSHED_BRANCH=""
        while read oldrev newrev refname; do
            if [[ "$newrev" != "0000000000000000000000000000000000000000" ]]; then
                branch=$(basename "$refname")
                echo "-----> Detected push to branch: $branch"
                if git log --oneline -1 "$branch" >/dev/null 2>&1; then
                    PUSHED_BRANCH="$branch"
                    break
                fi
            fi
        done

        # If we found a pushed branch with commits, switch HEAD to it
        if [[ -n "$PUSHED_BRANCH" ]]; then
            echo "-----> Switching HEAD from $CURRENT_HEAD_BRANCH to $PUSHED_BRANCH"
            git symbolic-ref HEAD "refs/heads/$PUSHED_BRANCH"
            CURRENT_HEAD_BRANCH="$PUSHED_BRANCH"
        else
            echo "-----> No suitable branch found with commits"
        fi
    else
        echo "-----> Current HEAD branch '$CURRENT_HEAD_BRANCH' has commits"
    fi

    echo "-----> Deploying from branch: $CURRENT_HEAD_BRANCH"
    echo "-----> Deploying $APP_NAME..."

    # Execute deployment using the centralized deploy command
    gokku deploy "$APP_NAME"

    echo "-----> Deployment completed"
else
    echo "-----> Repository is empty, skipping deployment"
    echo "-----> Run 'gokku deploy $APP_NAME' manually after your first push"
fi

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

	envContent := fmt.Sprintf("# App: %s\n# Generated: %s\nZERO_DOWNTIME=0\n", appName, time.Now().Format("2006-01-02 15:04:05"))

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

// isRunningOnServer checks if we're running directly on the Gokku server
func isRunningOnServer() bool {
	// Check if /opt/gokku directory exists and is accessible
	if _, err := os.Stat("/opt/gokku"); err != nil {
		return false
	}

	// Check if we're in a git repository (if not, we're likely on server)
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	if err := cmd.Run(); err != nil {
		// No git repository found, likely running on server
		return true
	}

	return false
}

// autoSetupRepository attempts to create the repository on the server if it doesn't exist
// This is kept for backward compatibility but now uses the complete setup
func autoSetupRepository(remoteInfo *internal.RemoteInfo) error {
	config, _ := internal.LoadConfig()
	return setupAppComplete(remoteInfo, remoteInfo.App, config)
}
